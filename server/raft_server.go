package server

import (
	"encoding/json"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"time"

	raftbadgerdb "github.com/bbva/raft-badger"
	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/mapping"
	"github.com/dgraph-io/badger/v2"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/hashicorp/raft"
	"github.com/mosuka/blast/errors"
	"github.com/mosuka/blast/marshaler"
	"github.com/mosuka/blast/metric"
	"github.com/mosuka/blast/protobuf"
	"go.uber.org/zap"
)

type RaftServer struct {
	id            string
	raftAddress   string
	dataDirectory string
	bootstrap     bool
	logger        *zap.Logger

	fsm *RaftFSM

	transport *raft.NetworkTransport
	raft      *raft.Raft

	watchClusterStopCh chan struct{}
	watchClusterDoneCh chan struct{}

	applyCh chan *protobuf.Event
}

func NewRaftServer(id string, raftAddress string, dataDirectory string, indexMapping *mapping.IndexMappingImpl, bootstrap bool, logger *zap.Logger) (*RaftServer, error) {
	indexPath := filepath.Join(dataDirectory, "index")
	fsm, err := NewRaftFSM(indexPath, indexMapping, logger)
	if err != nil {
		logger.Error("failed to create FSM", zap.String("index_path", indexPath), zap.Error(err))
		return nil, err
	}

	return &RaftServer{
		id:            id,
		raftAddress:   raftAddress,
		dataDirectory: dataDirectory,
		bootstrap:     bootstrap,
		fsm:           fsm,
		logger:        logger,

		watchClusterStopCh: make(chan struct{}),
		watchClusterDoneCh: make(chan struct{}),

		applyCh: make(chan *protobuf.Event, 1024),
	}, nil
}

func (s *RaftServer) Start() error {
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(s.id)
	config.SnapshotThreshold = 1024
	config.LogOutput = ioutil.Discard

	addr, err := net.ResolveTCPAddr("tcp", s.raftAddress)
	if err != nil {
		s.logger.Error("failed to resolve TCP address", zap.String("raft_address", s.raftAddress), zap.Error(err))
		return err
	}

	s.transport, err = raft.NewTCPTransport(s.raftAddress, addr, 3, 10*time.Second, ioutil.Discard)
	if err != nil {
		s.logger.Error("failed to create TCP transport", zap.String("raft_address", s.raftAddress), zap.Error(err))
		return err
	}

	// create snapshot store
	snapshotStore, err := raft.NewFileSnapshotStore(s.dataDirectory, 2, ioutil.Discard)
	if err != nil {
		s.logger.Error("failed to create file snapshot store", zap.String("path", s.dataDirectory), zap.Error(err))
		return err
	}

	logStorePath := filepath.Join(s.dataDirectory, "raft", "log")
	err = os.MkdirAll(logStorePath, 0755)
	if err != nil {
		s.logger.Fatal(err.Error())
		return err
	}
	logStoreBadgerOpts := badger.DefaultOptions(logStorePath)
	logStoreBadgerOpts.ValueDir = logStorePath
	logStoreBadgerOpts.SyncWrites = false
	logStoreBadgerOpts.Logger = nil
	logStoreOpts := raftbadgerdb.Options{
		Path:          logStorePath,
		BadgerOptions: &logStoreBadgerOpts,
	}
	raftLogStore, err := raftbadgerdb.New(logStoreOpts)
	if err != nil {
		s.logger.Fatal(err.Error())
		return err
	}

	stableStorePath := filepath.Join(s.dataDirectory, "raft", "stable")
	err = os.MkdirAll(stableStorePath, 0755)
	if err != nil {
		s.logger.Fatal(err.Error())
		return err
	}
	stableStoreBadgerOpts := badger.DefaultOptions(stableStorePath)
	stableStoreBadgerOpts.ValueDir = stableStorePath
	stableStoreBadgerOpts.SyncWrites = false
	stableStoreBadgerOpts.Logger = nil
	stableStoreOpts := raftbadgerdb.Options{
		Path:          stableStorePath,
		BadgerOptions: &stableStoreBadgerOpts,
	}
	raftStableStore, err := raftbadgerdb.New(stableStoreOpts)
	if err != nil {
		s.logger.Fatal(err.Error())
		return err
	}

	// create raft
	s.raft, err = raft.NewRaft(config, s.fsm, raftLogStore, raftStableStore, snapshotStore, s.transport)
	if err != nil {
		s.logger.Error("failed to create raft", zap.Any("config", config), zap.Error(err))
		return err
	}

	if s.bootstrap {
		configuration := raft.Configuration{
			Servers: []raft.Server{
				{
					ID:      config.LocalID,
					Address: s.transport.LocalAddr(),
				},
			},
		}
		s.raft.BootstrapCluster(configuration)
	}

	go func() {
		s.startWatchCluster(500 * time.Millisecond)
	}()

	s.logger.Info("Raft server started", zap.String("raft_address", s.raftAddress))
	return nil
}

func (s *RaftServer) Stop() error {
	s.applyCh <- nil
	s.logger.Info("apply channel has closed")

	s.stopWatchCluster()

	if err := s.fsm.Close(); err != nil {
		s.logger.Error("failed to close FSM", zap.Error(err))
	}
	s.logger.Info("Raft FSM Closed")

	if future := s.raft.Shutdown(); future.Error() != nil {
		s.logger.Info("failed to shutdown Raft", zap.Error(future.Error()))
	}
	s.logger.Info("Raft has shutdown", zap.String("raft_address", s.raftAddress))

	return nil
}

func (s *RaftServer) startWatchCluster(checkInterval time.Duration) {
	s.logger.Info("start to update cluster info")

	defer func() {
		close(s.watchClusterDoneCh)
	}()

	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()

	timeout := 60 * time.Second
	if err := s.WaitForDetectLeader(timeout); err != nil {
		if err == errors.ErrTimeout {
			s.logger.Error("leader detection timed out", zap.Duration("timeout", timeout), zap.Error(err))
		} else {
			s.logger.Error("failed to detect leader", zap.Error(err))
		}
	}

	for {
		select {
		case <-s.watchClusterStopCh:
			s.logger.Info("received a request to stop updating a cluster")
			return
		case <-s.raft.LeaderCh():
			s.logger.Info("became a leader", zap.String("leaderAddr", string(s.raft.Leader())))
		case event := <-s.fsm.applyCh:
			s.applyCh <- event
		case <-ticker.C:
			raftStats := s.raft.Stats()

			switch raftStats["state"] {
			case "Follower":
				metric.RaftStateMetric.WithLabelValues(s.id).Set(float64(raft.Follower))
			case "Candidate":
				metric.RaftStateMetric.WithLabelValues(s.id).Set(float64(raft.Candidate))
			case "Leader":
				metric.RaftStateMetric.WithLabelValues(s.id).Set(float64(raft.Leader))
			case "Shutdown":
				metric.RaftStateMetric.WithLabelValues(s.id).Set(float64(raft.Shutdown))
			}

			if term, err := strconv.ParseFloat(raftStats["term"], 64); err == nil {
				metric.RaftTermMetric.WithLabelValues(s.id).Set(term)
			}

			if lastLogIndex, err := strconv.ParseFloat(raftStats["last_log_index"], 64); err == nil {
				metric.RaftLastLogIndexMetric.WithLabelValues(s.id).Set(lastLogIndex)
			}

			if lastLogTerm, err := strconv.ParseFloat(raftStats["last_log_term"], 64); err == nil {
				metric.RaftLastLogTermMetric.WithLabelValues(s.id).Set(lastLogTerm)
			}

			if commitIndex, err := strconv.ParseFloat(raftStats["commit_index"], 64); err == nil {
				metric.RaftCommitIndexMetric.WithLabelValues(s.id).Set(commitIndex)
			}

			if appliedIndex, err := strconv.ParseFloat(raftStats["applied_index"], 64); err == nil {
				metric.RaftAppliedIndexMetric.WithLabelValues(s.id).Set(appliedIndex)
			}

			if fsmPending, err := strconv.ParseFloat(raftStats["fsm_pending"], 64); err == nil {
				metric.RaftFsmPendingMetric.WithLabelValues(s.id).Set(fsmPending)
			}

			if lastSnapshotIndex, err := strconv.ParseFloat(raftStats["last_snapshot_index"], 64); err == nil {
				metric.RaftLastSnapshotIndexMetric.WithLabelValues(s.id).Set(lastSnapshotIndex)
			}

			if lastSnapshotTerm, err := strconv.ParseFloat(raftStats["last_snapshot_term"], 64); err == nil {
				metric.RaftLastSnapshotTermMetric.WithLabelValues(s.id).Set(lastSnapshotTerm)
			}

			if latestConfigurationIndex, err := strconv.ParseFloat(raftStats["latest_configuration_index"], 64); err == nil {
				metric.RaftLatestConfigurationIndexMetric.WithLabelValues(s.id).Set(latestConfigurationIndex)
			}

			if numPeers, err := strconv.ParseFloat(raftStats["num_peers"], 64); err == nil {
				metric.RaftNumPeersMetric.WithLabelValues(s.id).Set(numPeers)
			}

			if lastContact, err := strconv.ParseFloat(raftStats["last_contact"], 64); err == nil {
				metric.RaftLastContactMetric.WithLabelValues(s.id).Set(lastContact)
			}

			if nodes, err := s.Nodes(); err == nil {
				metric.RaftNumNodesMetric.WithLabelValues(s.id).Set(float64(len(nodes)))
			}

			indexStats := s.fsm.Stats()

			tmpIndex := indexStats["index"].(map[string]interface{})

			metric.IndexCurOnDiskBytesMetric.WithLabelValues(s.id).Set(float64(tmpIndex["CurOnDiskBytes"].(uint64)))

			metric.IndexCurOnDiskFilesMetric.WithLabelValues(s.id).Set(float64(tmpIndex["CurOnDiskFiles"].(uint64)))

			metric.IndexCurRootEpochMetric.WithLabelValues(s.id).Set(float64(tmpIndex["CurRootEpoch"].(uint64)))

			metric.IndexLastMergedEpochMetric.WithLabelValues(s.id).Set(float64(tmpIndex["LastMergedEpoch"].(uint64)))

			metric.IndexLastPersistedEpochMetric.WithLabelValues(s.id).Set(float64(tmpIndex["LastPersistedEpoch"].(uint64)))

			metric.IndexMaxBatchIntroTimeMetric.WithLabelValues(s.id).Set(float64(tmpIndex["MaxBatchIntroTime"].(uint64)))

			metric.IndexMaxFileMergeZapTimeMetric.WithLabelValues(s.id).Set(float64(tmpIndex["MaxFileMergeZapTime"].(uint64)))

			metric.IndexMaxMemMergeZapTimeMetric.WithLabelValues(s.id).Set(float64(tmpIndex["MaxMemMergeZapTime"].(uint64)))

			metric.IndexTotAnalysisTimeMetric.WithLabelValues(s.id).Set(float64(tmpIndex["TotAnalysisTime"].(uint64)))

			metric.IndexTotBatchIntroTimeMetric.WithLabelValues(s.id).Set(float64(tmpIndex["TotBatchIntroTime"].(uint64)))

			metric.IndexTotBatchesMetric.WithLabelValues(s.id).Set(float64(tmpIndex["TotBatches"].(uint64)))

			metric.IndexTotBatchesEmptyMetric.WithLabelValues(s.id).Set(float64(tmpIndex["TotBatchesEmpty"].(uint64)))

			metric.IndexTotDeletesMetric.WithLabelValues(s.id).Set(float64(tmpIndex["TotDeletes"].(uint64)))

			metric.IndexTotFileMergeIntroductionsMetric.WithLabelValues(s.id).Set(float64(tmpIndex["TotFileMergeIntroductions"].(uint64)))

			metric.IndexTotFileMergeIntroductionsDoneMetric.WithLabelValues(s.id).Set(float64(tmpIndex["TotFileMergeIntroductionsDone"].(uint64)))

			metric.IndexTotFileMergeIntroductionsSkippedMetric.WithLabelValues(s.id).Set(float64(tmpIndex["TotFileMergeIntroductionsSkipped"].(uint64)))

			metric.IndexTotFileMergeLoopBegMetric.WithLabelValues(s.id).Set(float64(tmpIndex["TotFileMergeLoopBeg"].(uint64)))

			metric.IndexTotFileMergeLoopEndMetric.WithLabelValues(s.id).Set(float64(tmpIndex["TotFileMergeLoopEnd"].(uint64)))

			metric.IndexTotFileMergeLoopErrMetric.WithLabelValues(s.id).Set(float64(tmpIndex["TotFileMergeLoopErr"].(uint64)))

			metric.IndexTotFileMergePlanMetric.WithLabelValues(s.id).Set(float64(tmpIndex["TotFileMergePlan"].(uint64)))

			metric.IndexTotFileMergePlanErrMetric.WithLabelValues(s.id).Set(float64(tmpIndex["TotFileMergePlanErr"].(uint64)))

			metric.IndexTotFileMergePlanNoneMetric.WithLabelValues(s.id).Set(float64(tmpIndex["TotFileMergePlanNone"].(uint64)))

			metric.IndexTotFileMergePlanOkMetric.WithLabelValues(s.id).Set(float64(tmpIndex["TotFileMergePlanOk"].(uint64)))

			metric.IndexTotFileMergePlanTasksMetric.WithLabelValues(s.id).Set(float64(tmpIndex["TotFileMergePlanTasks"].(uint64)))

			metric.IndexTotFileMergePlanTasksDoneMetric.WithLabelValues(s.id).Set(float64(tmpIndex["TotFileMergePlanTasksDone"].(uint64)))

			metric.IndexTotFileMergePlanTasksErrMetric.WithLabelValues(s.id).Set(float64(tmpIndex["TotFileMergePlanTasksErr"].(uint64)))

			metric.IndexTotFileMergePlanTasksSegmentsMetric.WithLabelValues(s.id).Set(float64(tmpIndex["TotFileMergePlanTasksSegments"].(uint64)))

			metric.IndexTotFileMergePlanTasksSegmentsEmptyMetric.WithLabelValues(s.id).Set(float64(tmpIndex["TotFileMergePlanTasksSegmentsEmpty"].(uint64)))

			metric.IndexTotFileMergeSegmentsMetric.WithLabelValues(s.id).Set(float64(tmpIndex["TotFileMergeSegments"].(uint64)))

			metric.IndexTotFileMergeSegmentsEmptyMetric.WithLabelValues(s.id).Set(float64(tmpIndex["TotFileMergeSegmentsEmpty"].(uint64)))

			metric.IndexTotFileMergeWrittenBytesMetric.WithLabelValues(s.id).Set(float64(tmpIndex["TotFileMergeWrittenBytes"].(uint64)))

			metric.IndexTotFileMergeZapBegMetric.WithLabelValues(s.id).Set(float64(tmpIndex["TotFileMergeZapBeg"].(uint64)))

			metric.IndexTotFileMergeZapEndMetric.WithLabelValues(s.id).Set(float64(tmpIndex["TotFileMergeZapEnd"].(uint64)))

			metric.IndexTotFileMergeZapTimeMetric.WithLabelValues(s.id).Set(float64(tmpIndex["TotFileMergeZapTime"].(uint64)))

			metric.IndexTotFileSegmentsAtRootMetric.WithLabelValues(s.id).Set(float64(tmpIndex["TotFileSegmentsAtRoot"].(uint64)))

			metric.IndexTotIndexTimeMetric.WithLabelValues(s.id).Set(float64(tmpIndex["TotIndexTime"].(uint64)))

			metric.IndexTotIndexedPlainTextBytesMetric.WithLabelValues(s.id).Set(float64(tmpIndex["TotIndexedPlainTextBytes"].(uint64)))

			metric.IndexTotIntroduceLoopMetric.WithLabelValues(s.id).Set(float64(tmpIndex["TotIntroduceLoop"].(uint64)))

			metric.IndexTotIntroduceMergeBegMetric.WithLabelValues(s.id).Set(float64(tmpIndex["TotIntroduceMergeBeg"].(uint64)))

			metric.IndexTotIntroduceMergeEndMetric.WithLabelValues(s.id).Set(float64(tmpIndex["TotIntroduceMergeEnd"].(uint64)))

			metric.IndexTotIntroducePersistBegMetric.WithLabelValues(s.id).Set(float64(tmpIndex["TotIntroducePersistBeg"].(uint64)))

			metric.IndexTotIntroducePersistEndMetric.WithLabelValues(s.id).Set(float64(tmpIndex["TotIntroducePersistEnd"].(uint64)))

			metric.IndexTotIntroduceRevertBegMetric.WithLabelValues(s.id).Set(float64(tmpIndex["TotIntroduceRevertBeg"].(uint64)))

			metric.IndexTotIntroduceRevertEndMetric.WithLabelValues(s.id).Set(float64(tmpIndex["TotIntroduceRevertEnd"].(uint64)))

			metric.IndexTotIntroduceSegmentBegMetric.WithLabelValues(s.id).Set(float64(tmpIndex["TotIntroduceSegmentBeg"].(uint64)))

			metric.IndexTotIntroduceSegmentEndMetric.WithLabelValues(s.id).Set(float64(tmpIndex["TotIntroduceSegmentEnd"].(uint64)))

			metric.IndexTotIntroducedItemsMetric.WithLabelValues(s.id).Set(float64(tmpIndex["TotIntroducedItems"].(uint64)))

			metric.IndexTotIntroducedSegmentsBatchMetric.WithLabelValues(s.id).Set(float64(tmpIndex["TotIntroducedSegmentsBatch"].(uint64)))

			metric.IndexTotIntroducedSegmentsMergeMetric.WithLabelValues(s.id).Set(float64(tmpIndex["TotIntroducedSegmentsMerge"].(uint64)))

			metric.IndexTotItemsToPersistMetric.WithLabelValues(s.id).Set(float64(tmpIndex["TotItemsToPersist"].(uint64)))

			metric.IndexTotMemMergeBegMetric.WithLabelValues(s.id).Set(float64(tmpIndex["TotMemMergeBeg"].(uint64)))

			metric.IndexTotMemMergeDoneMetric.WithLabelValues(s.id).Set(float64(tmpIndex["TotMemMergeDone"].(uint64)))

			metric.IndexTotMemMergeErrMetric.WithLabelValues(s.id).Set(float64(tmpIndex["TotMemMergeErr"].(uint64)))

			metric.IndexTotMemMergeSegmentsMetric.WithLabelValues(s.id).Set(float64(tmpIndex["TotMemMergeSegments"].(uint64)))

			metric.IndexTotMemMergeZapBegMetric.WithLabelValues(s.id).Set(float64(tmpIndex["TotMemMergeZapBeg"].(uint64)))

			metric.IndexTotMemMergeZapEndMetric.WithLabelValues(s.id).Set(float64(tmpIndex["TotMemMergeZapEnd"].(uint64)))

			metric.IndexTotMemMergeZapTimeMetric.WithLabelValues(s.id).Set(float64(tmpIndex["TotMemMergeZapTime"].(uint64)))

			metric.IndexTotMemorySegmentsAtRootMetric.WithLabelValues(s.id).Set(float64(tmpIndex["TotMemorySegmentsAtRoot"].(uint64)))

			metric.IndexTotOnErrorsMetric.WithLabelValues(s.id).Set(float64(tmpIndex["TotOnErrors"].(uint64)))

			metric.IndexTotPersistLoopBegMetric.WithLabelValues(s.id).Set(float64(tmpIndex["TotPersistLoopBeg"].(uint64)))

			metric.IndexTotPersistLoopEndMetric.WithLabelValues(s.id).Set(float64(tmpIndex["TotPersistLoopEnd"].(uint64)))

			metric.IndexTotPersistLoopErrMetric.WithLabelValues(s.id).Set(float64(tmpIndex["TotPersistLoopErr"].(uint64)))

			metric.IndexTotPersistLoopProgressMetric.WithLabelValues(s.id).Set(float64(tmpIndex["TotPersistLoopProgress"].(uint64)))

			metric.IndexTotPersistLoopWaitMetric.WithLabelValues(s.id).Set(float64(tmpIndex["TotPersistLoopWait"].(uint64)))

			metric.IndexTotPersistLoopWaitNotifiedMetric.WithLabelValues(s.id).Set(float64(tmpIndex["TotPersistLoopWaitNotified"].(uint64)))

			metric.IndexTotPersistedItemsMetric.WithLabelValues(s.id).Set(float64(tmpIndex["TotPersistedItems"].(uint64)))

			metric.IndexTotPersistedSegmentsMetric.WithLabelValues(s.id).Set(float64(tmpIndex["TotPersistedSegments"].(uint64)))

			metric.IndexTotPersisterMergerNapBreakMetric.WithLabelValues(s.id).Set(float64(tmpIndex["TotPersisterMergerNapBreak"].(uint64)))

			metric.IndexTotPersisterNapPauseCompletedMetric.WithLabelValues(s.id).Set(float64(tmpIndex["TotPersisterNapPauseCompleted"].(uint64)))

			metric.IndexTotPersisterSlowMergerPauseMetric.WithLabelValues(s.id).Set(float64(tmpIndex["TotPersisterSlowMergerPause"].(uint64)))

			metric.IndexTotPersisterSlowMergerResumeMetric.WithLabelValues(s.id).Set(float64(tmpIndex["TotPersisterSlowMergerResume"].(uint64)))

			metric.IndexTotTermSearchersFinishedMetric.WithLabelValues(s.id).Set(float64(tmpIndex["TotTermSearchersFinished"].(uint64)))

			metric.IndexTotTermSearchersStartedMetric.WithLabelValues(s.id).Set(float64(tmpIndex["TotTermSearchersStarted"].(uint64)))

			metric.IndexTotUpdatesMetric.WithLabelValues(s.id).Set(float64(tmpIndex["TotUpdates"].(uint64)))

			metric.IndexAnalysisTimeMetric.WithLabelValues(s.id).Set(float64(tmpIndex["analysis_time"].(uint64)))

			metric.IndexBatchesMetric.WithLabelValues(s.id).Set(float64(tmpIndex["batches"].(uint64)))

			metric.IndexDeletesMetric.WithLabelValues(s.id).Set(float64(tmpIndex["deletes"].(uint64)))

			metric.IndexErrorsMetric.WithLabelValues(s.id).Set(float64(tmpIndex["errors"].(uint64)))

			metric.IndexIndexTimeMetric.WithLabelValues(s.id).Set(float64(tmpIndex["index_time"].(uint64)))

			metric.IndexNumBytesUsedDiskMetric.WithLabelValues(s.id).Set(float64(tmpIndex["num_bytes_used_disk"].(uint64)))

			metric.IndexNumFilesOnDiskMetric.WithLabelValues(s.id).Set(float64(tmpIndex["num_files_on_disk"].(uint64)))

			metric.IndexNumItemsIntroducedMetric.WithLabelValues(s.id).Set(float64(tmpIndex["num_items_introduced"].(uint64)))

			metric.IndexNumItemsPersistedMetric.WithLabelValues(s.id).Set(float64(tmpIndex["num_items_persisted"].(uint64)))

			metric.IndexNumPersisterNapMergerBreakMetric.WithLabelValues(s.id).Set(float64(tmpIndex["num_persister_nap_merger_break"].(uint64)))

			metric.IndexNumPersisterNapPauseCompletedMetric.WithLabelValues(s.id).Set(float64(tmpIndex["num_persister_nap_pause_completed"].(uint64)))

			metric.IndexNumPlainTextBytesIndexedMetric.WithLabelValues(s.id).Set(float64(tmpIndex["num_plain_text_bytes_indexed"].(uint64)))

			metric.IndexNumRecsToPersistMetric.WithLabelValues(s.id).Set(float64(tmpIndex["num_recs_to_persist"].(uint64)))

			metric.IndexNumRootFilesegmentsMetric.WithLabelValues(s.id).Set(float64(tmpIndex["num_root_filesegments"].(uint64)))

			metric.IndexNumRootMemorysegmentsMetric.WithLabelValues(s.id).Set(float64(tmpIndex["num_root_memorysegments"].(uint64)))

			metric.IndexTermSearchersFinishedMetric.WithLabelValues(s.id).Set(float64(tmpIndex["term_searchers_finished"].(uint64)))

			metric.IndexTermSearchersStartedMetric.WithLabelValues(s.id).Set(float64(tmpIndex["term_searchers_started"].(uint64)))

			metric.IndexTotalCompactionWrittenBytesMetric.WithLabelValues(s.id).Set(float64(tmpIndex["total_compaction_written_bytes"].(uint64)))

			metric.IndexUpdatesMetric.WithLabelValues(s.id).Set(float64(tmpIndex["updates"].(uint64)))

			metric.SearchTimeMetric.WithLabelValues(s.id).Set(float64(indexStats["search_time"].(uint64)))

			metric.SearchesMetric.WithLabelValues(s.id).Set(float64(indexStats["searches"].(uint64)))
		}
	}
}

func (s *RaftServer) stopWatchCluster() {
	if s.watchClusterStopCh != nil {
		s.logger.Info("send a request to stop updating a cluster")
		close(s.watchClusterStopCh)
	}

	s.logger.Info("wait for the cluster update to stop")
	<-s.watchClusterDoneCh
	s.logger.Info("the cluster update has been stopped")
}

func (s *RaftServer) LeaderAddress(timeout time.Duration) (raft.ServerAddress, error) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	for {
		select {
		case <-ticker.C:
			leaderAddr := s.raft.Leader()
			if leaderAddr != "" {
				s.logger.Debug("detected a leader address", zap.String("raft_address", string(leaderAddr)))
				return leaderAddr, nil
			}
		case <-timer.C:
			err := errors.ErrTimeout
			s.logger.Error("failed to detect leader address", zap.Error(err))
			return "", err
		}
	}
}

func (s *RaftServer) LeaderID(timeout time.Duration) (raft.ServerID, error) {
	leaderAddr, err := s.LeaderAddress(timeout)
	if err != nil {
		s.logger.Error("failed to get leader address", zap.Error(err))
		return "", err
	}

	cf := s.raft.GetConfiguration()
	if err = cf.Error(); err != nil {
		s.logger.Error("failed to get Raft configuration", zap.Error(err))
		return "", err
	}

	for _, server := range cf.Configuration().Servers {
		if server.Address == leaderAddr {
			s.logger.Info("detected a leader ID", zap.String("id", string(server.ID)))
			return server.ID, nil
		}
	}

	err = errors.ErrNotFoundLeader
	s.logger.Error("failed to detect leader ID", zap.Error(err))
	return "", err
}

func (s *RaftServer) WaitForDetectLeader(timeout time.Duration) error {
	if _, err := s.LeaderAddress(timeout); err != nil {
		s.logger.Error("failed to wait for detect leader", zap.Error(err))
		return err
	}

	return nil
}

func (s *RaftServer) State() raft.RaftState {
	return s.raft.State()
}

func (s *RaftServer) StateStr() string {
	return s.State().String()
}

func (s *RaftServer) Exist(id string) (bool, error) {
	exist := false

	cf := s.raft.GetConfiguration()
	if err := cf.Error(); err != nil {
		s.logger.Error("failed to get Raft configuration", zap.Error(err))
		return false, err
	}

	for _, server := range cf.Configuration().Servers {
		if server.ID == raft.ServerID(id) {
			s.logger.Debug("node already joined the cluster", zap.String("id", id))
			exist = true
			break
		}
	}

	return exist, nil
}

func (s *RaftServer) getMetadata(id string) (*protobuf.Metadata, error) {
	metadata := s.fsm.getMetadata(id)
	if metadata == nil {
		return nil, errors.ErrNotFound
	}

	return metadata, nil
}

func (s *RaftServer) setMetadata(id string, metadata *protobuf.Metadata) error {
	data := &protobuf.SetMetadataRequest{
		Id:       id,
		Metadata: metadata,
	}

	dataAny := &any.Any{}
	if err := marshaler.UnmarshalAny(data, dataAny); err != nil {
		s.logger.Error("failed to unmarshal request to the command data", zap.String("id", id), zap.Any("metadata", metadata), zap.Error(err))
		return err
	}

	event := &protobuf.Event{
		Type: protobuf.Event_Join,
		Data: dataAny,
	}

	msg, err := proto.Marshal(event)
	if err != nil {
		s.logger.Error("failed to marshal the command into the bytes as message", zap.String("id", id), zap.Any("metadata", metadata), zap.Error(err))
		return err
	}

	timeout := 60 * time.Second
	if future := s.raft.Apply(msg, timeout); future.Error() != nil {
		s.logger.Error("failed to apply message bytes", zap.Duration("timeout", timeout), zap.Error(future.Error()))
		return future.Error()
	}

	return nil
}

func (s *RaftServer) deleteMetadata(id string) error {
	data := &protobuf.DeleteMetadataRequest{
		Id: id,
	}

	dataAny := &any.Any{}
	if err := marshaler.UnmarshalAny(data, dataAny); err != nil {
		s.logger.Error("failed to unmarshal request to the command data", zap.String("id", id), zap.Error(err))
		return err
	}

	event := &protobuf.Event{
		Type: protobuf.Event_Leave,
		Data: dataAny,
	}

	msg, err := proto.Marshal(event)
	if err != nil {
		s.logger.Error("failed to marshal the command into the bytes as the message", zap.String("id", id), zap.Error(err))
		return err
	}

	timeout := 60 * time.Second
	if future := s.raft.Apply(msg, timeout); future.Error() != nil {
		s.logger.Error("failed to apply message bytes", zap.Duration("timeout", timeout), zap.Error(future.Error()))
		return future.Error()
	}

	return nil
}

func (s *RaftServer) Join(id string, node *protobuf.Node) error {
	exist, err := s.Exist(id)
	if err != nil {
		return err
	}

	if !exist {
		if future := s.raft.AddVoter(raft.ServerID(id), raft.ServerAddress(node.RaftAddress), 0, 0); future.Error() != nil {
			s.logger.Error("failed to add voter", zap.String("id", id), zap.String("raft_address", node.RaftAddress), zap.Error(future.Error()))
			return future.Error()
		}
		s.logger.Info("node has successfully joined", zap.String("id", id), zap.String("raft_address", node.RaftAddress))
	}

	if err := s.setMetadata(id, node.Metadata); err != nil {
		return err
	}

	if exist {
		return errors.ErrNodeAlreadyExists
	}

	return nil
}

func (s *RaftServer) Leave(id string) error {
	exist, err := s.Exist(id)
	if err != nil {
		return err
	}

	if exist {
		if future := s.raft.RemoveServer(raft.ServerID(id), 0, 0); future.Error() != nil {
			s.logger.Error("failed to remove server", zap.String("id", id), zap.Error(future.Error()))
			return future.Error()
		}
		s.logger.Info("node has successfully left", zap.String("id", id))
	}

	if err = s.deleteMetadata(id); err != nil {
		return err
	}

	if !exist {
		return errors.ErrNodeDoesNotExist
	}

	return nil
}

func (s *RaftServer) Node() (*protobuf.Node, error) {
	nodes, err := s.Nodes()
	if err != nil {
		return nil, err
	}

	node, ok := nodes[s.id]
	if !ok {
		return nil, errors.ErrNotFound
	}

	node.State = s.StateStr()

	return node, nil
}

func (s *RaftServer) Nodes() (map[string]*protobuf.Node, error) {
	cf := s.raft.GetConfiguration()
	if err := cf.Error(); err != nil {
		s.logger.Error("failed to get Raft configuration", zap.Error(err))
		return nil, err
	}

	nodes := make(map[string]*protobuf.Node, 0)
	for _, server := range cf.Configuration().Servers {
		metadata, _ := s.getMetadata(string(server.ID))

		nodes[string(server.ID)] = &protobuf.Node{
			RaftAddress: string(server.Address),
			Metadata:    metadata,
		}
	}

	return nodes, nil
}

func (s *RaftServer) Snapshot() error {
	if future := s.raft.Snapshot(); future.Error() != nil {
		s.logger.Error("failed to snapshot", zap.Error(future.Error()))
		return future.Error()
	}

	return nil
}

func (s *RaftServer) Get(id string) (map[string]interface{}, error) {
	return s.fsm.get(id)
}

func (s *RaftServer) Search(searchRequest *bleve.SearchRequest) (*bleve.SearchResult, error) {
	return s.fsm.search(searchRequest)
}

func (s *RaftServer) Set(req *protobuf.SetRequest) error {
	dataAny := &any.Any{}
	if err := marshaler.UnmarshalAny(req, dataAny); err != nil {
		s.logger.Error("failed to unmarshal document map to any", zap.Error(err))
		return err
	}

	event := &protobuf.Event{
		Type: protobuf.Event_Set,
		Data: dataAny,
	}

	msg, err := proto.Marshal(event)
	if err != nil {
		s.logger.Error("failed to marshal event to bytes", zap.Error(err))
		return err
	}

	timeout := 60 * time.Second
	if future := s.raft.Apply(msg, timeout); future.Error() != nil {
		s.logger.Error("failed to apply message bytes", zap.Duration("timeout", timeout), zap.Error(future.Error()))
		return future.Error()
	}

	return nil
}

func (s *RaftServer) Delete(req *protobuf.DeleteRequest) error {
	dataAny := &any.Any{}
	if err := marshaler.UnmarshalAny(req, dataAny); err != nil {
		s.logger.Error("failed to unmarshal id to any", zap.Error(err))
		return err
	}

	c := &protobuf.Event{
		Type: protobuf.Event_Delete,
		Data: dataAny,
	}

	msg, err := proto.Marshal(c)
	if err != nil {
		s.logger.Error("failed to marshal event to bytes", zap.Error(err))
		return err
	}

	timeout := 60 * time.Second
	if future := s.raft.Apply(msg, timeout); future.Error() != nil {
		s.logger.Error("failed to apply message bytes", zap.Duration("timeout", timeout), zap.Error(future.Error()))
		return future.Error()
	}

	return nil
}

func (s *RaftServer) BulkIndex(req *protobuf.BulkIndexRequest) error {
	dataAny := &any.Any{}
	if err := marshaler.UnmarshalAny(req, dataAny); err != nil {
		s.logger.Error("failed to unmarshal bulk index request to any", zap.Error(err))
		return err
	}

	event := &protobuf.Event{
		Type: protobuf.Event_BulkIndex,
		Data: dataAny,
	}

	msg, err := proto.Marshal(event)
	if err != nil {
		s.logger.Error("failed to marshal event to bytes", zap.Error(err))
		return err
	}

	timeout := 60 * time.Second
	if future := s.raft.Apply(msg, timeout); future.Error() != nil {
		s.logger.Error("failed to apply message bytes", zap.Duration("timeout", timeout), zap.Error(future.Error()))
		return future.Error()
	}

	return nil
}

func (s *RaftServer) BulkDelete(req *protobuf.BulkDeleteRequest) error {
	dataAny := &any.Any{}
	if err := marshaler.UnmarshalAny(req, dataAny); err != nil {
		s.logger.Error("failed to unmarshal set request to any", zap.Error(err))
		return err
	}

	event := &protobuf.Event{
		Type: protobuf.Event_BulkDelete,
		Data: dataAny,
	}

	msg, err := proto.Marshal(event)
	if err != nil {
		s.logger.Error("failed to marshal event to bytes", zap.Error(err))
		return err
	}

	timeout := 60 * time.Second
	if future := s.raft.Apply(msg, timeout); future.Error() != nil {
		s.logger.Error("failed to apply message bytes", zap.Duration("timeout", timeout), zap.Error(future.Error()))
		return future.Error()
	}

	return nil
}

func (s *RaftServer) Mapping() (*protobuf.MappingResponse, error) {
	resp := &protobuf.MappingResponse{}

	m := s.fsm.Mapping()

	fieldsBytes, err := json.Marshal(m)
	if err != nil {
		s.logger.Error("failed to marshal mapping to bytes", zap.Error(err))
		return resp, err
	}

	resp.Mapping = fieldsBytes

	return resp, nil
}
