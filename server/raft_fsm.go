package server

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"sync"
	"time"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/mapping"
	"github.com/golang/protobuf/proto"
	"github.com/hashicorp/raft"
	"github.com/mosuka/blast/errors"
	"github.com/mosuka/blast/marshaler"
	"github.com/mosuka/blast/protobuf"
	"github.com/mosuka/blast/storage"
	"go.uber.org/zap"
)

type ApplyResponse struct {
	count int
	error error
}

type RaftFSM struct {
	logger *zap.Logger

	index      *storage.Index
	metadata   map[string]*protobuf.Metadata
	nodesMutex sync.RWMutex

	applyCh chan *protobuf.Event
}

func NewRaftFSM(path string, indexMapping *mapping.IndexMappingImpl, logger *zap.Logger) (*RaftFSM, error) {
	index, err := storage.NewIndex(path, indexMapping, logger)
	if err != nil {
		logger.Error("failed to create index store", zap.String("path", path), zap.Error(err))
		return nil, err
	}

	return &RaftFSM{
		logger:   logger,
		index:    index,
		metadata: make(map[string]*protobuf.Metadata, 0),
		applyCh:  make(chan *protobuf.Event, 1024),
	}, nil
}

func (f *RaftFSM) Close() error {
	f.applyCh <- nil
	f.logger.Info("apply channel has closed")

	if err := f.index.Close(); err != nil {
		f.logger.Error("failed to close index store", zap.Error(err))
		return err
	}

	f.logger.Info("Index has closed")

	return nil
}

func (f *RaftFSM) get(id string) (map[string]interface{}, error) {
	return f.index.Get(id)
}

func (f *RaftFSM) search(searchRequest *bleve.SearchRequest) (*bleve.SearchResult, error) {
	return f.index.Search(searchRequest)
}

func (f *RaftFSM) set(id string, fields map[string]interface{}) error {
	return f.index.Index(id, fields)
}

func (f *RaftFSM) delete(id string) error {
	return f.index.Delete(id)
}

func (f *RaftFSM) bulkIndex(docs []map[string]interface{}) (int, error) {
	return f.index.BulkIndex(docs)
}

func (f *RaftFSM) bulkDelete(ids []string) (int, error) {
	return f.index.BulkDelete(ids)
}

func (f *RaftFSM) getMetadata(id string) *protobuf.Metadata {
	if metadata, exists := f.metadata[id]; exists {
		return metadata
	} else {
		f.logger.Debug("metadata not found", zap.String("id", id))
		return nil
	}
}

func (f *RaftFSM) setMetadata(id string, metadata *protobuf.Metadata) error {
	f.nodesMutex.Lock()
	defer f.nodesMutex.Unlock()

	f.metadata[id] = metadata

	return nil
}

func (f *RaftFSM) deleteMetadata(id string) error {
	f.nodesMutex.Lock()
	defer f.nodesMutex.Unlock()

	if _, exists := f.metadata[id]; exists {
		delete(f.metadata, id)
	}

	return nil
}

func (f *RaftFSM) Apply(l *raft.Log) interface{} {
	var event protobuf.Event
	err := proto.Unmarshal(l.Data, &event)
	if err != nil {
		f.logger.Error("failed to unmarshal message bytes to KVS command", zap.Error(err))
		return err
	}

	switch event.Type {
	case protobuf.Event_Join:
		data, err := marshaler.MarshalAny(event.Data)
		if err != nil {
			f.logger.Error("failed to marshal to request from KVS command request", zap.String("type", event.Type.String()), zap.Error(err))
			return &ApplyResponse{error: err}
		}
		if data == nil {
			err = errors.ErrNil
			f.logger.Error("request is nil", zap.String("type", event.Type.String()), zap.Error(err))
			return &ApplyResponse{error: err}
		}
		req := data.(*protobuf.SetMetadataRequest)

		if err := f.setMetadata(req.Id, req.Metadata); err != nil {
			return &ApplyResponse{error: err}
		}

		f.applyCh <- &event

		return &ApplyResponse{}
	case protobuf.Event_Leave:
		data, err := marshaler.MarshalAny(event.Data)
		if err != nil {
			f.logger.Error("failed to marshal to request from KVS command request", zap.String("type", event.Type.String()), zap.Error(err))
			return &ApplyResponse{error: err}
		}
		if data == nil {
			err = errors.ErrNil
			f.logger.Error("request is nil", zap.String("type", event.Type.String()), zap.Error(err))
			return &ApplyResponse{error: err}
		}
		req := *data.(*protobuf.DeleteMetadataRequest)

		if err := f.deleteMetadata(req.Id); err != nil {
			return &ApplyResponse{error: err}
		}

		f.applyCh <- &event

		return &ApplyResponse{}
	case protobuf.Event_Set:
		data, err := marshaler.MarshalAny(event.Data)
		if err != nil {
			f.logger.Error("failed to marshal event data to set request", zap.String("type", event.Type.String()), zap.Error(err))
			return &ApplyResponse{error: err}
		}
		if data == nil {
			err = errors.ErrNil
			f.logger.Error("request is nil", zap.String("type", event.Type.String()), zap.Error(err))
			return &ApplyResponse{error: err}
		}
		req := *data.(*protobuf.SetRequest)

		var fields map[string]interface{}
		if err := json.Unmarshal(req.Fields, &fields); err != nil {
			return &ApplyResponse{error: err}
		}

		if err := f.set(req.Id, fields); err != nil {
			return &ApplyResponse{error: err}
		}

		f.applyCh <- &event

		return &ApplyResponse{}
	case protobuf.Event_Delete:
		data, err := marshaler.MarshalAny(event.Data)
		if err != nil {
			f.logger.Error("failed to marshal event data to delete request", zap.String("type", event.Type.String()), zap.Error(err))
			return &ApplyResponse{error: err}
		}
		if data == nil {
			err = errors.ErrNil
			f.logger.Error("request is nil", zap.String("type", event.Type.String()), zap.Error(err))
			return &ApplyResponse{error: err}
		}
		req := *data.(*protobuf.DeleteRequest)

		if err := f.delete(req.Id); err != nil {
			return &ApplyResponse{error: err}
		}

		f.applyCh <- &event

		return &ApplyResponse{}
	case protobuf.Event_BulkIndex:
		data, err := marshaler.MarshalAny(event.Data)
		if err != nil {
			f.logger.Error("failed to marshal event data to set request", zap.String("type", event.Type.String()), zap.Error(err))
			return &ApplyResponse{count: -1, error: nil}
		}
		if data == nil {
			err = errors.ErrNil
			f.logger.Error("request is nil", zap.String("type", event.Type.String()), zap.Error(err))
			return &ApplyResponse{count: -1, error: nil}
		}
		req := *data.(*protobuf.BulkIndexRequest)

		docs := make([]map[string]interface{}, 0)
		for _, r := range req.Requests {
			var fields map[string]interface{}
			if err := json.Unmarshal(r.Fields, &fields); err != nil {
				f.logger.Error("failed to unmarshal bytes to map", zap.String("id", r.Id), zap.Error(err))
				continue
			}

			doc := map[string]interface{}{
				"id":     r.Id,
				"fields": fields,
			}
			docs = append(docs, doc)
		}

		count, err := f.bulkIndex(docs)
		if err != nil {
			return &ApplyResponse{count: count, error: err}
		}

		f.applyCh <- &event

		return &ApplyResponse{count: count, error: nil}
	case protobuf.Event_BulkDelete:
		data, err := marshaler.MarshalAny(event.Data)
		if err != nil {
			f.logger.Error("failed to marshal event data to set request", zap.String("type", event.Type.String()), zap.Error(err))
			return &ApplyResponse{count: -1, error: nil}
		}
		if data == nil {
			err = errors.ErrNil
			f.logger.Error("request is nil", zap.String("type", event.Type.String()), zap.Error(err))
			return &ApplyResponse{count: -1, error: nil}
		}
		req := *data.(*protobuf.BulkDeleteRequest)

		ids := make([]string, 0)
		for _, r := range req.Requests {
			ids = append(ids, r.Id)
		}

		count, err := f.bulkDelete(ids)
		if err != nil {
			return &ApplyResponse{count: count, error: err}
		}

		f.applyCh <- &event

		return &ApplyResponse{count: count, error: nil}
	default:
		err = errors.ErrUnsupportedEvent
		f.logger.Error("unsupported command", zap.String("type", event.Type.String()), zap.Error(err))
		return &ApplyResponse{error: err}
	}
}

func (f *RaftFSM) Stats() map[string]interface{} {
	return f.index.Stats()
}

func (f *RaftFSM) Mapping() *mapping.IndexMappingImpl {
	return f.index.Mapping()
}

func (f *RaftFSM) Snapshot() (raft.FSMSnapshot, error) {
	return &KVSFSMSnapshot{
		index:  f.index,
		logger: f.logger,
	}, nil
}

func (f *RaftFSM) Restore(rc io.ReadCloser) error {
	start := time.Now()

	f.logger.Info("start to restore items")

	defer func() {
		err := rc.Close()
		if err != nil {
			f.logger.Error("failed to close reader", zap.Error(err))
		}
	}()

	data, err := ioutil.ReadAll(rc)
	if err != nil {
		f.logger.Error("failed to open reader", zap.Error(err))
		return err
	}

	count := uint64(0)

	buff := proto.NewBuffer(data)
	for {
		doc := &protobuf.Document{}
		err = buff.DecodeMessage(doc)
		if err == io.ErrUnexpectedEOF {
			f.logger.Debug("reached the EOF", zap.Error(err))
			break
		}
		if err != nil {
			f.logger.Error("failed to read document", zap.Error(err))
			return err
		}

		var fields map[string]interface{}
		if err := json.Unmarshal(doc.Fields, &fields); err != nil {
			f.logger.Error("failed to unmarshal fields bytes to map", zap.Error(err))
			continue
		}

		// apply item to store
		if err = f.index.Index(doc.Id, fields); err != nil {
			f.logger.Error("failed to index document", zap.Error(err))
			continue
		}

		f.logger.Debug("document restored", zap.String("id", doc.Id))
		count = count + 1
	}

	f.logger.Info("finished to restore items", zap.Uint64("count", count), zap.Float64("time", float64(time.Since(start))/float64(time.Second)))

	return nil
}

// ---------------------

type KVSFSMSnapshot struct {
	index  *storage.Index
	logger *zap.Logger
}

func (f *KVSFSMSnapshot) Persist(sink raft.SnapshotSink) error {
	start := time.Now()

	f.logger.Info("start to persist items")

	defer func() {
		if err := sink.Close(); err != nil {
			f.logger.Error("failed to close sink", zap.Error(err))
		}
	}()

	ch := f.index.SnapshotItems()

	count := uint64(0)

	for {
		doc := <-ch
		if doc == nil {
			f.logger.Debug("channel closed")
			break
		}

		count = count + 1

		buff := proto.NewBuffer([]byte{})
		if err := buff.EncodeMessage(doc); err != nil {
			f.logger.Error("failed to encode document", zap.Error(err))
			return err
		}

		if _, err := sink.Write(buff.Bytes()); err != nil {
			f.logger.Error("failed to write document", zap.Error(err))
			return err
		}
	}

	f.logger.Info("finished to persist items", zap.Uint64("count", count), zap.Float64("time", float64(time.Since(start))/float64(time.Second)))

	return nil
}

func (f *KVSFSMSnapshot) Release() {
	f.logger.Info("release")
}
