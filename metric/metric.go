package metric

import (
	grpcprometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// Create a metrics registry.
	Registry = prometheus.NewRegistry()

	// Create some standard server metrics.
	GrpcMetrics = grpcprometheus.NewServerMetrics(
		func(o *prometheus.CounterOpts) {
			o.Namespace = "blast"
		},
	)

	// Raft node state metric
	RaftStateMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "raft",
		Name:      "state",
		Help:      "Node state. 0:Follower, 1:Candidate, 2:Leader, 3:Shutdown",
	}, []string{"id"})

	RaftTermMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "raft",
		Name:      "term",
		Help:      "Term.",
	}, []string{"id"})

	RaftLastLogIndexMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "raft",
		Name:      "last_log_index",
		Help:      "Last log index.",
	}, []string{"id"})

	RaftLastLogTermMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "raft",
		Name:      "last_log_term",
		Help:      "Last log term.",
	}, []string{"id"})

	RaftCommitIndexMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "raft",
		Name:      "commit_index",
		Help:      "Commit index.",
	}, []string{"id"})

	RaftAppliedIndexMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "raft",
		Name:      "applied_index",
		Help:      "Applied index.",
	}, []string{"id"})

	RaftFsmPendingMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "raft",
		Name:      "fsm_pending",
		Help:      "FSM pending.",
	}, []string{"id"})

	RaftLastSnapshotIndexMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "raft",
		Name:      "last_snapshot_index",
		Help:      "Last snapshot index.",
	}, []string{"id"})

	RaftLastSnapshotTermMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "raft",
		Name:      "last_snapshot_term",
		Help:      "Last snapshot term.",
	}, []string{"id"})

	RaftLatestConfigurationIndexMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "raft",
		Name:      "latest_configuration_index",
		Help:      "Latest configuration index.",
	}, []string{"id"})

	RaftNumPeersMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "raft",
		Name:      "num_peers",
		Help:      "Number of peers.",
	}, []string{"id"})

	RaftLastContactMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "raft",
		Name:      "last_copntact",
		Help:      "Last contact.",
	}, []string{"id"})

	RaftNumNodesMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "raft",
		Name:      "num_nodes",
		Help:      "Number of nodes.",
	}, []string{"id"})

	IndexCurOnDiskBytesMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "cur_on_disk_bytes",
		Help:      "cur_on_disk_bytes",
	}, []string{"id"})

	IndexCurOnDiskFilesMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "cur_on_disk_files",
		Help:      "cur_on_disk_files",
	}, []string{"id"})

	IndexCurRootEpochMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "cur_root_epoch",
		Help:      "cur_root_epoch",
	}, []string{"id"})

	IndexLastMergedEpochMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "last_merged_epoch",
		Help:      "last_merged_epoch",
	}, []string{"id"})

	IndexLastPersistedEpochMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "last_persisted_epoch",
		Help:      "last_persisted_epoch",
	}, []string{"id"})

	IndexMaxBatchIntroTimeMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "max_batch_intro_time",
		Help:      "max_batch_intro_time",
	}, []string{"id"})

	IndexMaxFileMergeZapTimeMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "max_file_merge_zap_time",
		Help:      "max_file_merge_zap_time",
	}, []string{"id"})

	IndexMaxMemMergeZapTimeMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "max_mem_merge_zap_time",
		Help:      "max_mem_merge_zap_time",
	}, []string{"id"})

	IndexTotAnalysisTimeMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "tot_analysis_time",
		Help:      "tot_analysis_time",
	}, []string{"id"})

	IndexTotBatchIntroTimeMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "tot_batch_intro_time",
		Help:      "tot_batch_intro_time",
	}, []string{"id"})

	IndexTotBatchesMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "tot_batches",
		Help:      "tot_batches",
	}, []string{"id"})

	IndexTotBatchesEmptyMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "tot_batches_empty",
		Help:      "tot_batches_empty",
	}, []string{"id"})

	IndexTotDeletesMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "tot_deletes",
		Help:      "tot_deletes",
	}, []string{"id"})

	IndexTotFileMergeIntroductionsMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "tot_file_merge_introductions",
		Help:      "tot_file_merge_introductions",
	}, []string{"id"})

	IndexTotFileMergeIntroductionsDoneMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "tot_file_merge_introductions_done",
		Help:      "tot_file_merge_introductions_done",
	}, []string{"id"})

	IndexTotFileMergeIntroductionsSkippedMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "tot_file_merge_introductions_skipped",
		Help:      "tot_file_merge_introductions_skipped",
	}, []string{"id"})

	IndexTotFileMergeLoopBegMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "tot_file_merge_loop_beg",
		Help:      "tot_file_merge_loop_beg",
	}, []string{"id"})

	IndexTotFileMergeLoopEndMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "tot_file_merge_loop_end",
		Help:      "tot_file_merge_loop_end",
	}, []string{"id"})

	IndexTotFileMergeLoopErrMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "tot_file_merge_loop_err",
		Help:      "tot_file_merge_loop_err",
	}, []string{"id"})

	IndexTotFileMergePlanMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "tot_file_merge_plan",
		Help:      "tot_file_merge_plan",
	}, []string{"id"})

	IndexTotFileMergePlanErrMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "tot_file_merge_plan_err",
		Help:      "tot_file_merge_plan_err",
	}, []string{"id"})

	IndexTotFileMergePlanNoneMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "tot_file_merge_plan_none",
		Help:      "tot_file_merge_plan_none",
	}, []string{"id"})

	IndexTotFileMergePlanOkMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "tot_file_merge_plan_ok",
		Help:      "tot_file_merge_plan_ok",
	}, []string{"id"})

	IndexTotFileMergePlanTasksMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "tot_file_merge_plan_tasks",
		Help:      "tot_file_merge_plan_tasks",
	}, []string{"id"})

	IndexTotFileMergePlanTasksDoneMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "tot_file_merge_plan_tasks_done",
		Help:      "tot_file_merge_plan_tasks_done",
	}, []string{"id"})

	IndexTotFileMergePlanTasksErrMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "tot_file_merge_plan_tasks_err",
		Help:      "tot_file_merge_plan_tasks_err",
	}, []string{"id"})

	IndexTotFileMergePlanTasksSegmentsMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "tot_file_merge_plan_tasks_segments",
		Help:      "tot_file_merge_plan_tasks_segments",
	}, []string{"id"})

	IndexTotFileMergePlanTasksSegmentsEmptyMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "tot_file_merge_plan_tasks_segments_empty",
		Help:      "tot_file_merge_plan_tasks_segments_empty",
	}, []string{"id"})

	IndexTotFileMergeSegmentsMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "tot_file_merge_segments",
		Help:      "tot_file_merge_segments",
	}, []string{"id"})

	IndexTotFileMergeSegmentsEmptyMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "tot_file_merge_segments_empty",
		Help:      "tot_file_merge_segments_empty",
	}, []string{"id"})

	IndexTotFileMergeWrittenBytesMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "tot_file_merge_written_bytes",
		Help:      "tot_file_merge_written_bytes",
	}, []string{"id"})

	IndexTotFileMergeZapBegMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "tot_file_merge_zap_beg",
		Help:      "tot_file_merge_zap_beg",
	}, []string{"id"})

	IndexTotFileMergeZapEndMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "tot_file_merge_zap_end",
		Help:      "tot_file_merge_zap_end",
	}, []string{"id"})

	IndexTotFileMergeZapTimeMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "tot_file_merge_zap_time",
		Help:      "tot_file_merge_zap_time",
	}, []string{"id"})

	IndexTotFileSegmentsAtRootMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "tot_file_segments_at_root",
		Help:      "tot_file_segments_at_root",
	}, []string{"id"})

	IndexTotIndexTimeMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "tot_index_time",
		Help:      "tot_index_time",
	}, []string{"id"})

	IndexTotIndexedPlainTextBytesMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "tot_indexed_plain_text_bytes",
		Help:      "tot_indexed_plain_text_bytes",
	}, []string{"id"})

	IndexTotIntroduceLoopMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "tot_introduce_loop",
		Help:      "tot_introduce_loop",
	}, []string{"id"})

	IndexTotIntroduceMergeBegMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "tot_introduce_merge_beg",
		Help:      "tot_introduce_merge_beg",
	}, []string{"id"})

	IndexTotIntroduceMergeEndMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "tot_introduce_merge_end",
		Help:      "tot_introduce_merge_end",
	}, []string{"id"})

	IndexTotIntroducePersistBegMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "tot_introduce_persist_beg",
		Help:      "tot_introduce_persist_beg",
	}, []string{"id"})

	IndexTotIntroducePersistEndMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "tot_introduce_persist_end",
		Help:      "tot_introduce_persist_end",
	}, []string{"id"})

	IndexTotIntroduceRevertBegMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "tot_introduce_revert_beg",
		Help:      "tot_introduce_revert_beg",
	}, []string{"id"})

	IndexTotIntroduceRevertEndMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "tot_introduce_revert_end",
		Help:      "tot_introduce_revert_end",
	}, []string{"id"})

	IndexTotIntroduceSegmentBegMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "tot_introduce_segment_beg",
		Help:      "tot_introduce_segment_beg",
	}, []string{"id"})

	IndexTotIntroduceSegmentEndMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "tot_introduce_segment_end",
		Help:      "tot_introduce_segment_end",
	}, []string{"id"})

	IndexTotIntroducedItemsMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "tot_introduced_items",
		Help:      "tot_introduced_items",
	}, []string{"id"})

	IndexTotIntroducedSegmentsBatchMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "tot_introduced_segments_batch",
		Help:      "tot_introduced_segments_batch",
	}, []string{"id"})

	IndexTotIntroducedSegmentsMergeMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "tot_introduced_segments_merge",
		Help:      "tot_introduced_segments_merge",
	}, []string{"id"})

	IndexTotItemsToPersistMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "tot_items_to_persist",
		Help:      "tot_items_to_persist",
	}, []string{"id"})

	IndexTotMemMergeBegMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "tot_mem_merge_beg",
		Help:      "tot_mem_merge_beg",
	}, []string{"id"})

	IndexTotMemMergeDoneMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "tot_mem_merge_done",
		Help:      "tot_mem_merge_done",
	}, []string{"id"})

	IndexTotMemMergeErrMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "tot_mem_merge_err",
		Help:      "tot_mem_merge_err",
	}, []string{"id"})

	IndexTotMemMergeSegmentsMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "tot_mem_merge_segments",
		Help:      "tot_mem_merge_segments",
	}, []string{"id"})

	IndexTotMemMergeZapBegMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "tot_mem_merge_zap_beg",
		Help:      "tot_mem_merge_zap_beg",
	}, []string{"id"})

	IndexTotMemMergeZapEndMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "tot_mem_merge_zap_end",
		Help:      "tot_mem_merge_zap_end",
	}, []string{"id"})

	IndexTotMemMergeZapTimeMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "tot_mem_merge_zap_time",
		Help:      "tot_mem_merge_zap_time",
	}, []string{"id"})

	IndexTotMemorySegmentsAtRootMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "tot_memory_segments_at_root",
		Help:      "tot_memory_segments_at_root",
	}, []string{"id"})

	IndexTotOnErrorsMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "tot_on_errors",
		Help:      "tot_on_errors",
	}, []string{"id"})

	IndexTotPersistLoopBegMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "tot_persist_loop_beg",
		Help:      "tot_persist_loop_beg",
	}, []string{"id"})

	IndexTotPersistLoopEndMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "tot_persist_loop_end",
		Help:      "tot_persist_loop_end",
	}, []string{"id"})

	IndexTotPersistLoopErrMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "tot_persist_loop_err",
		Help:      "tot_persist_loop_err",
	}, []string{"id"})

	IndexTotPersistLoopProgressMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "tot_persist_loop_progress",
		Help:      "tot_persist_loop_progress",
	}, []string{"id"})

	IndexTotPersistLoopWaitMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "tot_persist_loop_wait",
		Help:      "tot_persist_loop_wait",
	}, []string{"id"})

	IndexTotPersistLoopWaitNotifiedMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "tot_persist_loop_wait_notified",
		Help:      "tot_persist_loop_wait_notified",
	}, []string{"id"})

	IndexTotPersistedItemsMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "tot_persisted_items",
		Help:      "tot_persisted_items",
	}, []string{"id"})

	IndexTotPersistedSegmentsMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "tot_persisted_segments",
		Help:      "tot_persisted_segments",
	}, []string{"id"})

	IndexTotPersisterMergerNapBreakMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "tot_persister_merger_nap_break",
		Help:      "tot_persister_merger_nap_break",
	}, []string{"id"})

	IndexTotPersisterNapPauseCompletedMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "tot_persister_nap_pause_completed",
		Help:      "tot_persister_nap_pause_completed",
	}, []string{"id"})

	IndexTotPersisterSlowMergerPauseMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "tot_persister_slow_merger_pause",
		Help:      "tot_persister_slow_merger_pause",
	}, []string{"id"})

	IndexTotPersisterSlowMergerResumeMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "tot_persister_slow_merger_resume",
		Help:      "tot_persister_slow_merger_resume",
	}, []string{"id"})

	IndexTotTermSearchersFinishedMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "tot_term_searchers_finished",
		Help:      "tot_term_searchers_finished",
	}, []string{"id"})

	IndexTotTermSearchersStartedMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "tot_term_searchers_started",
		Help:      "tot_term_searchers_started",
	}, []string{"id"})

	IndexTotUpdatesMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "tot_updates",
		Help:      "tot_updates",
	}, []string{"id"})

	IndexAnalysisTimeMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "analysis_time",
		Help:      "analysis_time",
	}, []string{"id"})

	IndexBatchesMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "batches",
		Help:      "batches",
	}, []string{"id"})

	IndexDeletesMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "deletes",
		Help:      "deletes",
	}, []string{"id"})

	IndexErrorsMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "errors",
		Help:      "errors",
	}, []string{"id"})

	IndexIndexTimeMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "index_time",
		Help:      "index_time",
	}, []string{"id"})

	IndexNumBytesUsedDiskMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "num_bytes_used_disk",
		Help:      "num_bytes_used_disk",
	}, []string{"id"})

	IndexNumFilesOnDiskMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "num_files_on_disk",
		Help:      "num_files_on_disk",
	}, []string{"id"})

	IndexNumItemsIntroducedMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "num_items_introduced",
		Help:      "num_items_introduced",
	}, []string{"id"})

	IndexNumItemsPersistedMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "num_items_persisted",
		Help:      "num_items_persisted",
	}, []string{"id"})

	IndexNumPersisterNapMergerBreakMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "num_persister_nap_merger_break",
		Help:      "num_persister_nap_merger_break",
	}, []string{"id"})

	IndexNumPersisterNapPauseCompletedMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "num_persister_nap_pause_completed",
		Help:      "num_persister_nap_pause_completed",
	}, []string{"id"})

	IndexNumPlainTextBytesIndexedMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "num_plain_text_bytes_indexed",
		Help:      "num_plain_text_bytes_indexed",
	}, []string{"id"})

	IndexNumRecsToPersistMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "num_recs_to_persist",
		Help:      "num_recs_to_persist",
	}, []string{"id"})

	IndexNumRootFilesegmentsMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "num_root_filesegments",
		Help:      "num_root_filesegments",
	}, []string{"id"})

	IndexNumRootMemorysegmentsMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "num_root_memorysegments",
		Help:      "num_root_memorysegments",
	}, []string{"id"})

	IndexTermSearchersFinishedMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "term_searchers_finished",
		Help:      "term_searchers_finished",
	}, []string{"id"})

	IndexTermSearchersStartedMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "term_searchers_started",
		Help:      "term_searchers_started",
	}, []string{"id"})

	IndexTotalCompactionWrittenBytesMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "total_compaction_written_bytes",
		Help:      "total_compaction_written_bytes",
	}, []string{"id"})

	IndexUpdatesMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "updates",
		Help:      "updates",
	}, []string{"id"})

	SearchTimeMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "search_time",
		Help:      "search_time",
	}, []string{"id"})

	SearchesMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "blast",
		Subsystem: "index",
		Name:      "searches",
		Help:      "searches",
	}, []string{"id"})
)

func init() {
	// Register standard server metrics and customized metrics to registry.
	Registry.MustRegister(
		GrpcMetrics,
		RaftStateMetric,
		RaftTermMetric,
		RaftLastLogIndexMetric,
		RaftLastLogTermMetric,
		RaftCommitIndexMetric,
		RaftAppliedIndexMetric,
		RaftFsmPendingMetric,
		RaftLastSnapshotIndexMetric,
		RaftLastSnapshotTermMetric,
		RaftLatestConfigurationIndexMetric,
		RaftNumPeersMetric,
		RaftLastContactMetric,
		RaftNumNodesMetric,
		IndexCurOnDiskBytesMetric,
		IndexCurOnDiskFilesMetric,
		IndexCurRootEpochMetric,
		IndexLastMergedEpochMetric,
		IndexLastPersistedEpochMetric,
		IndexMaxBatchIntroTimeMetric,
		IndexMaxFileMergeZapTimeMetric,
		IndexMaxMemMergeZapTimeMetric,
		IndexTotAnalysisTimeMetric,
		IndexTotBatchIntroTimeMetric,
		IndexTotBatchesMetric,
		IndexTotBatchesEmptyMetric,
		IndexTotDeletesMetric,
		IndexTotFileMergeIntroductionsMetric,
		IndexTotFileMergeIntroductionsDoneMetric,
		IndexTotFileMergeIntroductionsSkippedMetric,
		IndexTotFileMergeLoopBegMetric,
		IndexTotFileMergeLoopEndMetric,
		IndexTotFileMergeLoopErrMetric,
		IndexTotFileMergePlanMetric,
		IndexTotFileMergePlanErrMetric,
		IndexTotFileMergePlanNoneMetric,
		IndexTotFileMergePlanOkMetric,
		IndexTotFileMergePlanTasksMetric,
		IndexTotFileMergePlanTasksDoneMetric,
		IndexTotFileMergePlanTasksErrMetric,
		IndexTotFileMergePlanTasksSegmentsMetric,
		IndexTotFileMergePlanTasksSegmentsEmptyMetric,
		IndexTotFileMergeSegmentsMetric,
		IndexTotFileMergeSegmentsEmptyMetric,
		IndexTotFileMergeWrittenBytesMetric,
		IndexTotFileMergeZapBegMetric,
		IndexTotFileMergeZapEndMetric,
		IndexTotFileMergeZapTimeMetric,
		IndexTotFileSegmentsAtRootMetric,
		IndexTotIndexTimeMetric,
		IndexTotIndexedPlainTextBytesMetric,
		IndexTotIntroduceLoopMetric,
		IndexTotIntroduceMergeBegMetric,
		IndexTotIntroduceMergeEndMetric,
		IndexTotIntroducePersistBegMetric,
		IndexTotIntroducePersistEndMetric,
		IndexTotIntroduceRevertBegMetric,
		IndexTotIntroduceRevertEndMetric,
		IndexTotIntroduceSegmentBegMetric,
		IndexTotIntroduceSegmentEndMetric,
		IndexTotIntroducedItemsMetric,
		IndexTotIntroducedSegmentsBatchMetric,
		IndexTotIntroducedSegmentsMergeMetric,
		IndexTotItemsToPersistMetric,
		IndexTotMemMergeBegMetric,
		IndexTotMemMergeDoneMetric,
		IndexTotMemMergeErrMetric,
		IndexTotMemMergeSegmentsMetric,
		IndexTotMemMergeZapBegMetric,
		IndexTotMemMergeZapEndMetric,
		IndexTotMemMergeZapTimeMetric,
		IndexTotMemorySegmentsAtRootMetric,
		IndexTotOnErrorsMetric,
		IndexTotPersistLoopBegMetric,
		IndexTotPersistLoopEndMetric,
		IndexTotPersistLoopErrMetric,
		IndexTotPersistLoopProgressMetric,
		IndexTotPersistLoopWaitMetric,
		IndexTotPersistLoopWaitNotifiedMetric,
		IndexTotPersistedItemsMetric,
		IndexTotPersistedSegmentsMetric,
		IndexTotPersisterMergerNapBreakMetric,
		IndexTotPersisterNapPauseCompletedMetric,
		IndexTotPersisterSlowMergerPauseMetric,
		IndexTotPersisterSlowMergerResumeMetric,
		IndexTotTermSearchersFinishedMetric,
		IndexTotTermSearchersStartedMetric,
		IndexTotUpdatesMetric,
		IndexAnalysisTimeMetric,
		IndexBatchesMetric,
		IndexDeletesMetric,
		IndexErrorsMetric,
		IndexIndexTimeMetric,
		IndexNumBytesUsedDiskMetric,
		IndexNumFilesOnDiskMetric,
		IndexNumItemsIntroducedMetric,
		IndexNumItemsPersistedMetric,
		IndexNumPersisterNapMergerBreakMetric,
		IndexNumPersisterNapPauseCompletedMetric,
		IndexNumPlainTextBytesIndexedMetric,
		IndexNumRecsToPersistMetric,
		IndexNumRootFilesegmentsMetric,
		IndexNumRootMemorysegmentsMetric,
		IndexTermSearchersFinishedMetric,
		IndexTermSearchersStartedMetric,
		IndexTotalCompactionWrittenBytesMetric,
		IndexUpdatesMetric,
		SearchTimeMetric,
		SearchesMetric,
	)
	GrpcMetrics.EnableHandlingTimeHistogram(
		func(o *prometheus.HistogramOpts) {
			o.Namespace = "blast"
		},
	)
}
