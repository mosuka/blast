package server

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/hashicorp/raft"
	"github.com/mosuka/blast/errors"
	"github.com/mosuka/blast/log"
	"github.com/mosuka/blast/mapping"
	"github.com/mosuka/blast/marshaler"
	"github.com/mosuka/blast/protobuf"
	"github.com/mosuka/blast/util"
)

func Test_RaftFSM_Close(t *testing.T) {
	curDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("%v", err)
	}

	tmpDir := util.TmpDir()
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	dir := filepath.Join(tmpDir, "index")

	indexMapping, err := mapping.NewIndexMappingFromFile(filepath.Join(curDir, "../examples/example_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}

	logger := log.NewLogger("WARN", "", 500, 3, 30, false)

	fsm, err := NewRaftFSM(dir, indexMapping, logger)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if fsm == nil {
		t.Fatal("failed to create index")
	}

	if err := fsm.Close(); err != nil {
		t.Fatalf("%v", err)
	}
}

func Test_RaftFSM_Set(t *testing.T) {
	curDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("%v", err)
	}

	tmpDir := util.TmpDir()
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	dir := filepath.Join(tmpDir, "index")

	indexMapping, err := mapping.NewIndexMappingFromFile(filepath.Join(curDir, "../examples/example_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}

	logger := log.NewLogger("WARN", "", 500, 3, 30, false)

	fsm, err := NewRaftFSM(dir, indexMapping, logger)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if fsm == nil {
		t.Fatal("failed to create index")
	}

	id := "1"
	fields := map[string]interface{}{
		"title":     "Search engine (computing)",
		"text":      "A search engine is an information retrieval system designed to help find information stored on a computer system. The search results are usually presented in a list and are commonly called hits. Search engines help to minimize the time required to find information and the amount of information which must be consulted, akin to other techniques for managing information overload. The most public, visible form of a search engine is a Web search engine which searches for information on the World Wide Web.",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"_type":     "example",
	}

	if ret := fsm.set(id, fields); ret != nil {
		t.Fatal("failed to index document")
	}

	if err := fsm.Close(); err != nil {
		t.Fatalf("%v", err)
	}
}

func Test_RaftFSM_Get(t *testing.T) {
	curDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("%v", err)
	}

	tmpDir := util.TmpDir()
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	dir := filepath.Join(tmpDir, "index")

	indexMapping, err := mapping.NewIndexMappingFromFile(filepath.Join(curDir, "../examples/example_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}

	logger := log.NewLogger("WARN", "", 500, 3, 30, false)

	fsm, err := NewRaftFSM(dir, indexMapping, logger)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if fsm == nil {
		t.Fatal("failed to create index")
	}

	id := "1"
	fields := map[string]interface{}{
		"title":     "Search engine (computing)",
		"text":      "A search engine is an information retrieval system designed to help find information stored on a computer system. The search results are usually presented in a list and are commonly called hits. Search engines help to minimize the time required to find information and the amount of information which must be consulted, akin to other techniques for managing information overload. The most public, visible form of a search engine is a Web search engine which searches for information on the World Wide Web.",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"_type":     "example",
	}

	if ret := fsm.set(id, fields); ret != nil {
		t.Fatal("failed to index document")
	}

	f, err := fsm.get(id)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if fields["title"].(string) != f["title"].(string) {
		t.Fatalf("expected content to see %v, saw %v", fields["title"].(string), f["title"].(string))
	}

	if err := fsm.Close(); err != nil {
		t.Fatalf("%v", err)
	}
}

func Test_RaftFSM_Delete(t *testing.T) {
	curDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("%v", err)
	}

	tmpDir := util.TmpDir()
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	dir := filepath.Join(tmpDir, "index")

	indexMapping, err := mapping.NewIndexMappingFromFile(filepath.Join(curDir, "../examples/example_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}

	logger := log.NewLogger("WARN", "", 500, 3, 30, false)

	fsm, err := NewRaftFSM(dir, indexMapping, logger)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if fsm == nil {
		t.Fatal("failed to create index")
	}

	id := "1"
	fields := map[string]interface{}{
		"title":     "Search engine (computing)",
		"text":      "A search engine is an information retrieval system designed to help find information stored on a computer system. The search results are usually presented in a list and are commonly called hits. Search engines help to minimize the time required to find information and the amount of information which must be consulted, akin to other techniques for managing information overload. The most public, visible form of a search engine is a Web search engine which searches for information on the World Wide Web.",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"_type":     "example",
	}

	if ret := fsm.set(id, fields); ret != nil {
		t.Fatal("failed to index document")
	}

	f, err := fsm.get(id)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if fields["title"].(string) != f["title"].(string) {
		t.Fatalf("expected content to see %v, saw %v", fields["title"].(string), f["title"].(string))
	}

	if ret := fsm.delete(id); ret != nil {
		t.Fatal("failed to delete document")
	}

	f, err = fsm.get(id)
	if err != nil {
		switch err {
		case errors.ErrNotFound:
			// ok
		default:
			t.Fatal("failed to get document")
		}
	}

	if err := fsm.Close(); err != nil {
		t.Fatalf("%v", err)
	}
}

func Test_RaftFSM_SetMetadata(t *testing.T) {
	curDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("%v", err)
	}

	tmpDir := util.TmpDir()
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	dir := filepath.Join(tmpDir, "index")

	indexMapping, err := mapping.NewIndexMappingFromFile(filepath.Join(curDir, "../examples/example_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}

	logger := log.NewLogger("WARN", "", 500, 3, 30, false)

	fsm, err := NewRaftFSM(dir, indexMapping, logger)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if fsm == nil {
		t.Fatal("failed to create index")
	}

	id := "node1"
	metadata := &protobuf.Metadata{
		GrpcAddress: ":9000",
		HttpAddress: ":8000",
	}

	if ret := fsm.setMetadata(id, metadata); ret != nil {
		t.Fatal("failed to index document")
	}

	if err := fsm.Close(); err != nil {
		t.Fatalf("%v", err)
	}
}

func Test_RaftFSM_GetMetadata(t *testing.T) {
	curDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("%v", err)
	}

	tmpDir := util.TmpDir()
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	dir := filepath.Join(tmpDir, "index")

	indexMapping, err := mapping.NewIndexMappingFromFile(filepath.Join(curDir, "../examples/example_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}

	logger := log.NewLogger("WARN", "", 500, 3, 30, false)

	fsm, err := NewRaftFSM(dir, indexMapping, logger)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if fsm == nil {
		t.Fatal("failed to create index")
	}

	id := "node1"
	metadata := &protobuf.Metadata{
		GrpcAddress: ":9000",
		HttpAddress: ":8000",
	}

	if ret := fsm.setMetadata(id, metadata); ret != nil {
		t.Fatal("failed to index document")
	}

	m := fsm.getMetadata(id)
	if metadata.GrpcAddress != m.GrpcAddress {
		t.Fatalf("expected content to see %v, saw %v", metadata.GrpcAddress, m.GrpcAddress)
	}

	if err := fsm.Close(); err != nil {
		t.Fatalf("%v", err)
	}
}

func Test_RaftFSM_DeleteMetadata(t *testing.T) {
	curDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("%v", err)
	}

	tmpDir := util.TmpDir()
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	dir := filepath.Join(tmpDir, "index")

	indexMapping, err := mapping.NewIndexMappingFromFile(filepath.Join(curDir, "../examples/example_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}

	logger := log.NewLogger("WARN", "", 500, 3, 30, false)

	fsm, err := NewRaftFSM(dir, indexMapping, logger)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if fsm == nil {
		t.Fatal("failed to create index")
	}

	id := "node1"
	metadata := &protobuf.Metadata{
		GrpcAddress: ":9000",
		HttpAddress: ":8000",
	}

	if ret := fsm.setMetadata(id, metadata); ret != nil {
		t.Fatal("failed to set metadata")
	}

	m := fsm.getMetadata(id)
	if metadata.GrpcAddress != m.GrpcAddress {
		t.Fatalf("expected content to see %v, saw %v", metadata.GrpcAddress, m.GrpcAddress)
	}

	if ret := fsm.deleteMetadata(id); ret != nil {
		t.Fatal("failed to delete metadata")
	}

	m = fsm.getMetadata(id)
	if m != nil {
		t.Fatalf("expected content to see %v, saw %v", nil, m.GrpcAddress)
	}

	if err := fsm.Close(); err != nil {
		t.Fatalf("%v", err)
	}
}

func Test_RaftFSM_ApplyJoin(t *testing.T) {
	curDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("%v", err)
	}

	tmpDir := util.TmpDir()
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	dir := filepath.Join(tmpDir, "index")

	indexMapping, err := mapping.NewIndexMappingFromFile(filepath.Join(curDir, "../examples/example_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}

	logger := log.NewLogger("WARN", "", 500, 3, 30, false)

	fsm, err := NewRaftFSM(dir, indexMapping, logger)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if fsm == nil {
		t.Fatal("failed to create index")
	}

	data := &protobuf.SetMetadataRequest{
		Id: "node1",
		Metadata: &protobuf.Metadata{
			GrpcAddress: ":9000",
			HttpAddress: ":8000",
		},
	}

	dataAny := &any.Any{}
	if err := marshaler.UnmarshalAny(data, dataAny); err != nil {
		t.Fatal("failed to unmarshal data to any")
	}

	event := &protobuf.Event{
		Type: protobuf.Event_Join,
		Data: dataAny,
	}

	eventData, err := proto.Marshal(event)
	if err != nil {
		t.Fatal("failed to marshal event to bytes")
	}

	raftLog := &raft.Log{
		Data: eventData,
	}

	ret := fsm.Apply(raftLog)
	if ret.(*ApplyResponse).error != nil {
		t.Fatal("failed to apply data")
	}

	m := fsm.getMetadata(data.Id)
	if data.Metadata.GrpcAddress != m.GrpcAddress {
		t.Fatalf("expected content to see %v, saw %v", data.Metadata.GrpcAddress, m.GrpcAddress)
	}

	if err := fsm.Close(); err != nil {
		t.Fatalf("%v", err)
	}
}

func Test_RaftFSM_ApplyLeave(t *testing.T) {
	curDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("%v", err)
	}

	tmpDir := util.TmpDir()
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	dir := filepath.Join(tmpDir, "index")

	indexMapping, err := mapping.NewIndexMappingFromFile(filepath.Join(curDir, "../examples/example_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}

	logger := log.NewLogger("WARN", "", 500, 3, 30, false)

	fsm, err := NewRaftFSM(dir, indexMapping, logger)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if fsm == nil {
		t.Fatal("failed to create index")
	}

	// apply joni
	setData := &protobuf.SetMetadataRequest{
		Id: "node1",
		Metadata: &protobuf.Metadata{
			GrpcAddress: ":9000",
			HttpAddress: ":8000",
		},
	}

	setDataAny := &any.Any{}
	if err := marshaler.UnmarshalAny(setData, setDataAny); err != nil {
		t.Fatal("failed to unmarshal data to any")
	}

	joinEvent := &protobuf.Event{
		Type: protobuf.Event_Join,
		Data: setDataAny,
	}

	joinEventData, err := proto.Marshal(joinEvent)
	if err != nil {
		t.Fatal("failed to marshal event to bytes")
	}

	joinRaftLog := &raft.Log{
		Data: joinEventData,
	}

	ret := fsm.Apply(joinRaftLog)
	if ret.(*ApplyResponse).error != nil {
		t.Fatal("failed to apply data")
	}

	m := fsm.getMetadata(setData.Id)
	if setData.Metadata.GrpcAddress != m.GrpcAddress {
		t.Fatalf("expected content to see %v, saw %v", setData.Metadata.GrpcAddress, m.GrpcAddress)
	}

	// apply leave
	deleteData := &protobuf.DeleteMetadataRequest{
		Id: "node1",
	}

	deleteDataAny := &any.Any{}
	if err := marshaler.UnmarshalAny(deleteData, deleteDataAny); err != nil {
		t.Fatal("failed to unmarshal data to any")
	}

	leaveEvent := &protobuf.Event{
		Type: protobuf.Event_Leave,
		Data: deleteDataAny,
	}

	leaveEventData, err := proto.Marshal(leaveEvent)
	if err != nil {
		t.Fatal("failed to marshal event to bytes")
	}

	leaveRaftLog := &raft.Log{
		Data: leaveEventData,
	}

	ret = fsm.Apply(leaveRaftLog)
	if ret.(*ApplyResponse).error != nil {
		t.Fatal("failed to apply data")
	}

	m = fsm.getMetadata(deleteData.Id)
	if m != nil {
		t.Fatalf("expected content to see %v, saw %v", nil, m.GrpcAddress)
	}

	if err := fsm.Close(); err != nil {
		t.Fatalf("%v", err)
	}
}

func Test_RaftFSM_ApplySet(t *testing.T) {
	curDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("%v", err)
	}

	tmpDir := util.TmpDir()
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	dir := filepath.Join(tmpDir, "index")

	indexMapping, err := mapping.NewIndexMappingFromFile(filepath.Join(curDir, "../examples/example_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}

	logger := log.NewLogger("WARN", "", 500, 3, 30, false)

	fsm, err := NewRaftFSM(dir, indexMapping, logger)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if fsm == nil {
		t.Fatal("failed to create index")
	}

	fields := map[string]interface{}{
		"title":     "Search engine (computing)",
		"text":      "A search engine is an information retrieval system designed to help find information stored on a computer system. The search results are usually presented in a list and are commonly called hits. Search engines help to minimize the time required to find information and the amount of information which must be consulted, akin to other techniques for managing information overload. The most public, visible form of a search engine is a Web search engine which searches for information on the World Wide Web.",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"_type":     "example",
	}
	fieldsBytes, err := json.Marshal(fields)
	if err != nil {
		t.Fatalf("%v", err)
	}

	// apply set
	setData := &protobuf.SetRequest{
		Id:     "1",
		Fields: fieldsBytes,
	}

	setDataAny := &any.Any{}
	if err := marshaler.UnmarshalAny(setData, setDataAny); err != nil {
		t.Fatal("failed to unmarshal data to any")
	}

	setEvent := &protobuf.Event{
		Type: protobuf.Event_Set,
		Data: setDataAny,
	}

	setEventData, err := proto.Marshal(setEvent)
	if err != nil {
		t.Fatal("failed to marshal event to bytes")
	}

	setRaftLog := &raft.Log{
		Data: setEventData,
	}

	ret := fsm.Apply(setRaftLog)
	if ret.(*ApplyResponse).error != nil {
		t.Fatal("failed to apply data")
	}

	f, err := fsm.get(setData.Id)
	if err != nil {
		t.Fatal("failed to get document")
	}
	if fields["title"] != f["title"] {
		t.Fatalf("expected content to see %v, saw %v", fields["title"], f["title"])
	}

	if err := fsm.Close(); err != nil {
		t.Fatalf("%v", err)
	}
}

func Test_RaftFSM_ApplyDelete(t *testing.T) {
	curDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("%v", err)
	}

	tmpDir := util.TmpDir()
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	dir := filepath.Join(tmpDir, "index")

	indexMapping, err := mapping.NewIndexMappingFromFile(filepath.Join(curDir, "../examples/example_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}

	logger := log.NewLogger("WARN", "", 500, 3, 30, false)

	fsm, err := NewRaftFSM(dir, indexMapping, logger)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if fsm == nil {
		t.Fatal("failed to create index")
	}

	fields := map[string]interface{}{
		"title":     "Search engine (computing)",
		"text":      "A search engine is an information retrieval system designed to help find information stored on a computer system. The search results are usually presented in a list and are commonly called hits. Search engines help to minimize the time required to find information and the amount of information which must be consulted, akin to other techniques for managing information overload. The most public, visible form of a search engine is a Web search engine which searches for information on the World Wide Web.",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"_type":     "example",
	}
	fieldsBytes, err := json.Marshal(fields)
	if err != nil {
		t.Fatalf("%v", err)
	}

	// apply set
	setData := &protobuf.SetRequest{
		Id:     "1",
		Fields: fieldsBytes,
	}

	setDataAny := &any.Any{}
	if err := marshaler.UnmarshalAny(setData, setDataAny); err != nil {
		t.Fatal("failed to unmarshal data to any")
	}

	setEvent := &protobuf.Event{
		Type: protobuf.Event_Set,
		Data: setDataAny,
	}

	setEventData, err := proto.Marshal(setEvent)
	if err != nil {
		t.Fatal("failed to marshal event to bytes")
	}

	setRaftLog := &raft.Log{
		Data: setEventData,
	}

	ret := fsm.Apply(setRaftLog)
	if ret.(*ApplyResponse).error != nil {
		t.Fatal("failed to apply data")
	}

	f, err := fsm.get(setData.Id)
	if err != nil {
		t.Fatal("failed to get document")
	}
	if fields["title"] != f["title"] {
		t.Fatalf("expected content to see %v, saw %v", fields["title"], f["title"])
	}

	// apply delete
	deleteData := &protobuf.DeleteRequest{
		Id: "1",
	}

	deleteDataAny := &any.Any{}
	if err := marshaler.UnmarshalAny(deleteData, deleteDataAny); err != nil {
		t.Fatal("failed to unmarshal data to any")
	}

	deleteEvent := &protobuf.Event{
		Type: protobuf.Event_Delete,
		Data: deleteDataAny,
	}

	deleteEventData, err := proto.Marshal(deleteEvent)
	if err != nil {
		t.Fatal("failed to marshal event to bytes")
	}

	deleteRaftLog := &raft.Log{
		Data: deleteEventData,
	}

	ret = fsm.Apply(deleteRaftLog)
	if ret.(*ApplyResponse).error != nil {
		t.Fatal("failed to apply data")
	}

	f, err = fsm.get(deleteData.Id)
	if err != nil {
		switch err {
		case errors.ErrNotFound:
			// ok
		default:
			t.Fatal("failed to get document")
		}
	}

	if err := fsm.Close(); err != nil {
		t.Fatalf("%v", err)
	}
}
