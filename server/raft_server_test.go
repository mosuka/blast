package server

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/hashicorp/raft"
	"github.com/mosuka/blast/errors"
	"github.com/mosuka/blast/log"
	"github.com/mosuka/blast/mapping"
	"github.com/mosuka/blast/protobuf"
	"github.com/mosuka/blast/util"
)

func Test_RaftServer_Close(t *testing.T) {
	curDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("%v", err)
	}

	rafAddress := fmt.Sprintf(":%d", util.TmpPort())

	dir := util.TmpDir()
	defer func() {
		_ = os.RemoveAll(dir)
	}()

	indexMapping, err := mapping.NewIndexMappingFromFile(filepath.Join(curDir, "../examples/example_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}

	logger := log.NewLogger("WARN", "", 500, 3, 30, false)

	server, err := NewRaftServer("node1", rafAddress, dir, indexMapping, true, logger)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if err := server.Start(); err != nil {
		t.Fatalf("%v", err)
	}
	defer func() {
		if err := server.Stop(); err != nil {
			t.Fatalf("%v", err)
		}
	}()

	time.Sleep(10 * time.Second)
}

func Test_RaftServer_LeaderAddress(t *testing.T) {
	curDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("%v", err)
	}

	raftAddress := fmt.Sprintf(":%d", util.TmpPort())

	dir := util.TmpDir()
	defer func() {
		_ = os.RemoveAll(dir)
	}()

	indexMapping, err := mapping.NewIndexMappingFromFile(filepath.Join(curDir, "../examples/example_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}

	logger := log.NewLogger("WARN", "", 500, 3, 30, false)

	server, err := NewRaftServer("node1", raftAddress, dir, indexMapping, true, logger)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if err := server.Start(); err != nil {
		t.Fatalf("%v", err)
	}
	defer func() {
		if err := server.Stop(); err != nil {
			t.Fatalf("%v", err)
		}
	}()

	leaderAddress, err := server.LeaderAddress(60 * time.Second)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if raftAddress != string(leaderAddress) {
		t.Fatalf("expected content to see %v, saw %v", raftAddress, string(leaderAddress))
	}
}

func Test_RaftServer_LeaderID(t *testing.T) {
	curDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("%v", err)
	}

	id := "node1"
	raftAddress := fmt.Sprintf(":%d", util.TmpPort())

	dir := util.TmpDir()
	defer func() {
		_ = os.RemoveAll(dir)
	}()

	indexMapping, err := mapping.NewIndexMappingFromFile(filepath.Join(curDir, "../examples/example_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}

	logger := log.NewLogger("WARN", "", 500, 3, 30, false)

	server, err := NewRaftServer(id, raftAddress, dir, indexMapping, true, logger)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if err := server.Start(); err != nil {
		t.Fatalf("%v", err)
	}
	defer func() {
		if err := server.Stop(); err != nil {
			t.Fatalf("%v", err)
		}
	}()

	leaderId, err := server.LeaderID(60 * time.Second)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if id != string(leaderId) {
		t.Fatalf("expected content to see %v, saw %v", id, string(leaderId))
	}
}

func Test_RaftServer_WaitForDetectLeader(t *testing.T) {
	curDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("%v", err)
	}

	id := "node1"
	raftAddress := fmt.Sprintf(":%d", util.TmpPort())

	dir := util.TmpDir()
	defer func() {
		_ = os.RemoveAll(dir)
	}()

	indexMapping, err := mapping.NewIndexMappingFromFile(filepath.Join(curDir, "../examples/example_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}

	logger := log.NewLogger("WARN", "", 500, 3, 30, false)

	server, err := NewRaftServer(id, raftAddress, dir, indexMapping, true, logger)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if err := server.Start(); err != nil {
		t.Fatalf("%v", err)
	}
	defer func() {
		if err := server.Stop(); err != nil {
			t.Fatalf("%v", err)
		}
	}()

	if err := server.WaitForDetectLeader(60 * time.Second); err != nil {
		t.Fatalf("%v", err)
	}
}

func Test_RaftServer_State(t *testing.T) {
	curDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("%v", err)
	}

	id := "node1"
	raftAddress := fmt.Sprintf(":%d", util.TmpPort())

	dir := util.TmpDir()
	defer func() {
		_ = os.RemoveAll(dir)
	}()

	indexMapping, err := mapping.NewIndexMappingFromFile(filepath.Join(curDir, "../examples/example_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}

	logger := log.NewLogger("WARN", "", 500, 3, 30, false)

	server, err := NewRaftServer(id, raftAddress, dir, indexMapping, true, logger)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if err := server.Start(); err != nil {
		t.Fatalf("%v", err)
	}
	defer func() {
		if err := server.Stop(); err != nil {
			t.Fatalf("%v", err)
		}
	}()

	if err := server.WaitForDetectLeader(60 * time.Second); err != nil {
		t.Fatalf("%v", err)
	}

	state := server.State()
	if raft.Leader != state {
		t.Fatalf("expected content to see %v, saw %v", raft.Leader, state)
	}
}

func Test_RaftServer_StateStr(t *testing.T) {
	curDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("%v", err)
	}

	id := "node1"
	raftAddress := fmt.Sprintf(":%d", util.TmpPort())

	dir := util.TmpDir()
	defer func() {
		_ = os.RemoveAll(dir)
	}()

	indexMapping, err := mapping.NewIndexMappingFromFile(filepath.Join(curDir, "../examples/example_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}

	logger := log.NewLogger("WARN", "", 500, 3, 30, false)

	server, err := NewRaftServer(id, raftAddress, dir, indexMapping, true, logger)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if err := server.Start(); err != nil {
		t.Fatalf("%v", err)
	}
	defer func() {
		if err := server.Stop(); err != nil {
			t.Fatalf("%v", err)
		}
	}()

	if err := server.WaitForDetectLeader(60 * time.Second); err != nil {
		t.Fatalf("%v", err)
	}

	state := server.StateStr()
	if raft.Leader.String() != state {
		t.Fatalf("expected content to see %v, saw %v", raft.Leader.String(), state)
	}
}

func Test_RaftServer_Exist(t *testing.T) {
	curDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("%v", err)
	}

	id := "node1"
	raftAddress := fmt.Sprintf(":%d", util.TmpPort())

	dir := util.TmpDir()
	defer func() {
		_ = os.RemoveAll(dir)
	}()

	indexMapping, err := mapping.NewIndexMappingFromFile(filepath.Join(curDir, "../examples/example_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}

	logger := log.NewLogger("WARN", "", 500, 3, 30, false)

	server, err := NewRaftServer(id, raftAddress, dir, indexMapping, true, logger)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if err := server.Start(); err != nil {
		t.Fatalf("%v", err)
	}
	defer func() {
		if err := server.Stop(); err != nil {
			t.Fatalf("%v", err)
		}
	}()

	if err := server.WaitForDetectLeader(60 * time.Second); err != nil {
		t.Fatalf("%v", err)
	}

	exist, err := server.Exist(id)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if !exist {
		t.Fatalf("expected content to see %v, saw %v", true, exist)
	}

	exist, err = server.Exist("non-existent-id")
	if err != nil {
		t.Fatalf("%v", err)
	}
	if exist {
		t.Fatalf("expected content to see %v, saw %v", false, exist)
	}
}

func Test_RaftServer_setMetadata(t *testing.T) {
	curDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("%v", err)
	}

	id := "node1"
	raftAddress := fmt.Sprintf(":%d", util.TmpPort())
	grpcAddress := fmt.Sprintf(":%d", util.TmpPort())
	httpAddress := fmt.Sprintf(":%d", util.TmpPort())

	dir := util.TmpDir()
	defer func() {
		_ = os.RemoveAll(dir)
	}()

	indexMapping, err := mapping.NewIndexMappingFromFile(filepath.Join(curDir, "../examples/example_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}

	logger := log.NewLogger("WARN", "", 500, 3, 30, false)

	server, err := NewRaftServer(id, raftAddress, dir, indexMapping, true, logger)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if err := server.Start(); err != nil {
		t.Fatalf("%v", err)
	}
	defer func() {
		if err := server.Stop(); err != nil {
			t.Fatalf("%v", err)
		}
	}()

	if err := server.WaitForDetectLeader(60 * time.Second); err != nil {
		t.Fatalf("%v", err)
	}

	metadata := &protobuf.Metadata{
		GrpcAddress: grpcAddress,
		HttpAddress: httpAddress,
	}

	if err := server.setMetadata(id, metadata); err != nil {
		t.Fatalf("%v", err)
	}
}

func Test_RaftServer_getMetadata(t *testing.T) {
	curDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("%v", err)
	}

	id := "node1"
	raftAddress := fmt.Sprintf(":%d", util.TmpPort())
	grpcAddress := fmt.Sprintf(":%d", util.TmpPort())
	httpAddress := fmt.Sprintf(":%d", util.TmpPort())

	dir := util.TmpDir()
	defer func() {
		_ = os.RemoveAll(dir)
	}()

	indexMapping, err := mapping.NewIndexMappingFromFile(filepath.Join(curDir, "../examples/example_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}

	logger := log.NewLogger("WARN", "", 500, 3, 30, false)

	server, err := NewRaftServer(id, raftAddress, dir, indexMapping, true, logger)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if err := server.Start(); err != nil {
		t.Fatalf("%v", err)
	}
	defer func() {
		if err := server.Stop(); err != nil {
			t.Fatalf("%v", err)
		}
	}()

	if err := server.WaitForDetectLeader(60 * time.Second); err != nil {
		t.Fatalf("%v", err)
	}

	metadata := &protobuf.Metadata{
		GrpcAddress: grpcAddress,
		HttpAddress: httpAddress,
	}

	if err := server.setMetadata(id, metadata); err != nil {
		t.Fatalf("%v", err)
	}

	m, err := server.getMetadata(id)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if grpcAddress != m.GrpcAddress {
		t.Fatalf("expected content to see %v, saw %v", grpcAddress, m.GrpcAddress)
	}
	if httpAddress != m.HttpAddress {
		t.Fatalf("expected content to see %v, saw %v", httpAddress, m.HttpAddress)
	}
}

func Test_RaftServer_deleteMetadata(t *testing.T) {
	curDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("%v", err)
	}

	id := "node1"
	raftAddress := fmt.Sprintf(":%d", util.TmpPort())
	grpcAddress := fmt.Sprintf(":%d", util.TmpPort())
	httpAddress := fmt.Sprintf(":%d", util.TmpPort())

	dir := util.TmpDir()
	defer func() {
		_ = os.RemoveAll(dir)
	}()

	indexMapping, err := mapping.NewIndexMappingFromFile(filepath.Join(curDir, "../examples/example_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}

	logger := log.NewLogger("WARN", "", 500, 3, 30, false)

	server, err := NewRaftServer(id, raftAddress, dir, indexMapping, true, logger)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if err := server.Start(); err != nil {
		t.Fatalf("%v", err)
	}
	defer func() {
		if err := server.Stop(); err != nil {
			t.Fatalf("%v", err)
		}
	}()

	if err := server.WaitForDetectLeader(60 * time.Second); err != nil {
		t.Fatalf("%v", err)
	}

	metadata := &protobuf.Metadata{
		GrpcAddress: grpcAddress,
		HttpAddress: httpAddress,
	}

	// set
	if err := server.setMetadata(id, metadata); err != nil {
		t.Fatalf("%v", err)
	}

	// get
	m, err := server.getMetadata(id)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if grpcAddress != m.GrpcAddress {
		t.Fatalf("expected content to see %v, saw %v", grpcAddress, m.GrpcAddress)
	}
	if httpAddress != m.HttpAddress {
		t.Fatalf("expected content to see %v, saw %v", httpAddress, m.HttpAddress)
	}

	// delete
	if err := server.deleteMetadata(id); err != nil {
		t.Fatalf("%v", err)
	}

	//get
	m, err = server.getMetadata(id)
	if err != nil {
		switch err {
		case errors.ErrNotFound:
			// ok
		default:
			t.Fatalf("%v", err)
		}
	}
	if err == nil {
		t.Fatalf("expected content to see %v, saw %v", nil, err)
	}
}

func Test_RaftServer_Join(t *testing.T) {
	curDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("%v", err)
	}

	id := "node1"
	raftAddress := fmt.Sprintf(":%d", util.TmpPort())
	grpcAddress := fmt.Sprintf(":%d", util.TmpPort())
	httpAddress := fmt.Sprintf(":%d", util.TmpPort())

	dir := util.TmpDir()
	defer func() {
		_ = os.RemoveAll(dir)
	}()

	indexMapping, err := mapping.NewIndexMappingFromFile(filepath.Join(curDir, "../examples/example_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}

	logger := log.NewLogger("WARN", "", 500, 3, 30, false)

	server, err := NewRaftServer(id, raftAddress, dir, indexMapping, true, logger)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if err := server.Start(); err != nil {
		t.Fatalf("%v", err)
	}
	defer func() {
		if err := server.Stop(); err != nil {
			t.Fatalf("%v", err)
		}
	}()

	if err := server.WaitForDetectLeader(60 * time.Second); err != nil {
		t.Fatalf("%v", err)
	}

	node := &protobuf.Node{
		RaftAddress: raftAddress,
		Metadata: &protobuf.Metadata{
			GrpcAddress: grpcAddress,
			HttpAddress: httpAddress,
		},
	}

	if err := server.Join(id, node); err != nil {
		switch err {
		case errors.ErrNodeAlreadyExists:
			// ok
		default:
			t.Fatalf("%v", err)
		}
	}
}

func Test_RaftServer_Node(t *testing.T) {
	curDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("%v", err)
	}

	id := "node1"
	raftAddress := fmt.Sprintf(":%d", util.TmpPort())
	grpcAddress := fmt.Sprintf(":%d", util.TmpPort())
	httpAddress := fmt.Sprintf(":%d", util.TmpPort())

	dir := util.TmpDir()
	defer func() {
		_ = os.RemoveAll(dir)
	}()

	indexMapping, err := mapping.NewIndexMappingFromFile(filepath.Join(curDir, "../examples/example_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}

	logger := log.NewLogger("WARN", "", 500, 3, 30, false)

	server, err := NewRaftServer(id, raftAddress, dir, indexMapping, true, logger)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if err := server.Start(); err != nil {
		t.Fatalf("%v", err)
	}
	defer func() {
		if err := server.Stop(); err != nil {
			t.Fatalf("%v", err)
		}
	}()

	if err := server.WaitForDetectLeader(60 * time.Second); err != nil {
		t.Fatalf("%v", err)
	}

	node := &protobuf.Node{
		RaftAddress: raftAddress,
		Metadata: &protobuf.Metadata{
			GrpcAddress: grpcAddress,
			HttpAddress: httpAddress,
		},
	}

	if err := server.Join(id, node); err != nil {
		switch err {
		case errors.ErrNodeAlreadyExists:
			// ok
		default:
			t.Fatalf("%v", err)
		}
	}

	n, err := server.Node()
	if err != nil {
		t.Fatalf("%v", err)
	}
	if raftAddress != n.RaftAddress {
		t.Fatalf("expected content to see %v, saw %v", raftAddress, n.RaftAddress)
	}
	if grpcAddress != n.Metadata.GrpcAddress {
		t.Fatalf("expected content to see %v, saw %v", grpcAddress, n.Metadata.GrpcAddress)
	}
	if httpAddress != n.Metadata.HttpAddress {
		t.Fatalf("expected content to see %v, saw %v", httpAddress, n.Metadata.HttpAddress)
	}
}

func Test_RaftServer_Cluster(t *testing.T) {
	curDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("%v", err)
	}

	id := "node1"
	raftAddress := fmt.Sprintf(":%d", util.TmpPort())
	grpcAddress := fmt.Sprintf(":%d", util.TmpPort())
	httpAddress := fmt.Sprintf(":%d", util.TmpPort())

	dir := util.TmpDir()
	defer func() {
		_ = os.RemoveAll(dir)
	}()

	indexMapping, err := mapping.NewIndexMappingFromFile(filepath.Join(curDir, "../examples/example_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}

	logger := log.NewLogger("WARN", "", 500, 3, 30, false)

	server, err := NewRaftServer(id, raftAddress, dir, indexMapping, true, logger)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if err := server.Start(); err != nil {
		t.Fatalf("%v", err)
	}
	defer func() {
		if err := server.Stop(); err != nil {
			t.Fatalf("%v", err)
		}
	}()

	if err := server.WaitForDetectLeader(60 * time.Second); err != nil {
		t.Fatalf("%v", err)
	}

	node := &protobuf.Node{
		RaftAddress: raftAddress,
		Metadata: &protobuf.Metadata{
			GrpcAddress: grpcAddress,
			HttpAddress: httpAddress,
		},
	}

	if err := server.Join(id, node); err != nil {
		switch err {
		case errors.ErrNodeAlreadyExists:
			// ok
		default:
			t.Fatalf("%v", err)
		}
	}

	// ----------

	id2 := "node2"
	raftAddress2 := fmt.Sprintf(":%d", util.TmpPort())
	grpcAddress2 := fmt.Sprintf(":%d", util.TmpPort())
	httpAddress2 := fmt.Sprintf(":%d", util.TmpPort())

	dir2 := util.TmpDir()
	defer func() {
		_ = os.RemoveAll(dir2)
	}()

	indexMapping2, err := mapping.NewIndexMappingFromFile(filepath.Join(curDir, "../examples/example_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}

	logger2 := log.NewLogger("WARN", "", 500, 3, 30, false)

	server2, err := NewRaftServer(id2, raftAddress2, dir2, indexMapping2, false, logger2)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if err := server2.Start(); err != nil {
		t.Fatalf("%v", err)
	}
	defer func() {
		if err := server2.Stop(); err != nil {
			t.Fatalf("%v", err)
		}
	}()

	node2 := &protobuf.Node{
		RaftAddress: raftAddress2,
		Metadata: &protobuf.Metadata{
			GrpcAddress: grpcAddress2,
			HttpAddress: httpAddress2,
		},
	}

	if err := server.Join(id2, node2); err != nil {
		switch err {
		case errors.ErrNodeAlreadyExists:
			// ok
		default:
			t.Fatalf("%v", err)
		}
	}

	// ----------

	id3 := "node3"
	raftAddress3 := fmt.Sprintf(":%d", util.TmpPort())
	grpcAddress3 := fmt.Sprintf(":%d", util.TmpPort())
	httpAddress3 := fmt.Sprintf(":%d", util.TmpPort())

	dir3 := util.TmpDir()
	defer func() {
		_ = os.RemoveAll(dir3)
	}()

	indexMapping3, err := mapping.NewIndexMappingFromFile(filepath.Join(curDir, "../examples/example_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}

	logger3 := log.NewLogger("WARN", "", 500, 3, 30, false)

	server3, err := NewRaftServer(id3, raftAddress3, dir3, indexMapping3, false, logger3)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if err := server3.Start(); err != nil {
		t.Fatalf("%v", err)
	}
	defer func() {
		if err := server3.Stop(); err != nil {
			t.Fatalf("%v", err)
		}
	}()

	node3 := &protobuf.Node{
		RaftAddress: raftAddress3,
		Metadata: &protobuf.Metadata{
			GrpcAddress: grpcAddress3,
			HttpAddress: httpAddress3,
		},
	}

	if err := server.Join(id3, node3); err != nil {
		switch err {
		case errors.ErrNodeAlreadyExists:
			// ok
		default:
			t.Fatalf("%v", err)
		}
	}

	ns, err := server.Nodes()
	if err != nil {
		t.Fatalf("%v", err)
	}
	if 3 != len(ns) {
		t.Fatalf("expected content to see %v, saw %v", 3, len(ns))
	}
	if raftAddress != ns[id].RaftAddress {
		t.Fatalf("expected content to see %v, saw %v", raftAddress, ns[id].RaftAddress)
	}
	if grpcAddress != ns[id].Metadata.GrpcAddress {
		t.Fatalf("expected content to see %v, saw %v", grpcAddress, ns[id].Metadata.GrpcAddress)
	}
	if httpAddress != ns[id].Metadata.HttpAddress {
		t.Fatalf("expected content to see %v, saw %v", httpAddress, ns[id].Metadata.HttpAddress)
	}
	if raftAddress2 != ns[id2].RaftAddress {
		t.Fatalf("expected content to see %v, saw %v", raftAddress2, ns[id2].RaftAddress)
	}
	if grpcAddress2 != ns[id2].Metadata.GrpcAddress {
		t.Fatalf("expected content to see %v, saw %v", grpcAddress2, ns[id2].Metadata.GrpcAddress)
	}
	if httpAddress2 != ns[id2].Metadata.HttpAddress {
		t.Fatalf("expected content to see %v, saw %v", httpAddress2, ns[id2].Metadata.HttpAddress)
	}
	if raftAddress3 != ns[id3].RaftAddress {
		t.Fatalf("expected content to see %v, saw %v", raftAddress3, ns[id3].RaftAddress)
	}
	if grpcAddress3 != ns[id3].Metadata.GrpcAddress {
		t.Fatalf("expected content to see %v, saw %v", grpcAddress3, ns[id3].Metadata.GrpcAddress)
	}
	if httpAddress3 != ns[id3].Metadata.HttpAddress {
		t.Fatalf("expected content to see %v, saw %v", httpAddress3, ns[id3].Metadata.HttpAddress)
	}

	time.Sleep(3 * time.Second)

	ns2, err := server2.Nodes()
	if err != nil {
		t.Fatalf("%v", err)
	}
	if 3 != len(ns2) {
		t.Fatalf("expected content to see %v, saw %v", 3, len(ns2))
	}
	if raftAddress != ns2[id].RaftAddress {
		t.Fatalf("expected content to see %v, saw %v", raftAddress, ns2[id].RaftAddress)
	}
	if grpcAddress != ns2[id].Metadata.GrpcAddress {
		t.Fatalf("expected content to see %v, saw %v", grpcAddress, ns2[id].Metadata.GrpcAddress)
	}
	if httpAddress != ns2[id].Metadata.HttpAddress {
		t.Fatalf("expected content to see %v, saw %v", httpAddress, ns2[id].Metadata.HttpAddress)
	}
	if raftAddress2 != ns2[id2].RaftAddress {
		t.Fatalf("expected content to see %v, saw %v", raftAddress2, ns2[id2].RaftAddress)
	}
	if grpcAddress2 != ns2[id2].Metadata.GrpcAddress {
		t.Fatalf("expected content to see %v, saw %v", grpcAddress2, ns2[id2].Metadata.GrpcAddress)
	}
	if httpAddress2 != ns2[id2].Metadata.HttpAddress {
		t.Fatalf("expected content to see %v, saw %v", httpAddress2, ns2[id2].Metadata.HttpAddress)
	}
	if raftAddress3 != ns2[id3].RaftAddress {
		t.Fatalf("expected content to see %v, saw %v", raftAddress3, ns2[id3].RaftAddress)
	}
	if grpcAddress3 != ns2[id3].Metadata.GrpcAddress {
		t.Fatalf("expected content to see %v, saw %v", grpcAddress3, ns2[id3].Metadata.GrpcAddress)
	}
	if httpAddress3 != ns2[id3].Metadata.HttpAddress {
		t.Fatalf("expected content to see %v, saw %v", httpAddress3, ns2[id3].Metadata.HttpAddress)
	}

	time.Sleep(3 * time.Second)

	ns3, err := server3.Nodes()
	if err != nil {
		t.Fatalf("%v", err)
	}
	if 3 != len(ns3) {
		t.Fatalf("expected content to see %v, saw %v", 3, len(ns3))
	}
	if raftAddress != ns3[id].RaftAddress {
		t.Fatalf("expected content to see %v, saw %v", raftAddress, ns3[id].RaftAddress)
	}
	if grpcAddress != ns3[id].Metadata.GrpcAddress {
		t.Fatalf("expected content to see %v, saw %v", grpcAddress, ns3[id].Metadata.GrpcAddress)
	}
	if httpAddress != ns3[id].Metadata.HttpAddress {
		t.Fatalf("expected content to see %v, saw %v", httpAddress, ns3[id].Metadata.HttpAddress)
	}
	if raftAddress2 != ns3[id2].RaftAddress {
		t.Fatalf("expected content to see %v, saw %v", raftAddress2, ns3[id2].RaftAddress)
	}
	if grpcAddress2 != ns3[id2].Metadata.GrpcAddress {
		t.Fatalf("expected content to see %v, saw %v", grpcAddress2, ns3[id2].Metadata.GrpcAddress)
	}
	if httpAddress2 != ns3[id2].Metadata.HttpAddress {
		t.Fatalf("expected content to see %v, saw %v", httpAddress2, ns3[id2].Metadata.HttpAddress)
	}
	if raftAddress3 != ns3[id3].RaftAddress {
		t.Fatalf("expected content to see %v, saw %v", raftAddress3, ns3[id3].RaftAddress)
	}
	if grpcAddress3 != ns3[id3].Metadata.GrpcAddress {
		t.Fatalf("expected content to see %v, saw %v", grpcAddress3, ns3[id3].Metadata.GrpcAddress)
	}
	if httpAddress3 != ns3[id3].Metadata.HttpAddress {
		t.Fatalf("expected content to see %v, saw %v", httpAddress3, ns3[id3].Metadata.HttpAddress)
	}
}

func Test_RaftServer_Leave(t *testing.T) {
	curDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("%v", err)
	}

	id := "node1"
	raftAddress := fmt.Sprintf(":%d", util.TmpPort())
	grpcAddress := fmt.Sprintf(":%d", util.TmpPort())
	httpAddress := fmt.Sprintf(":%d", util.TmpPort())

	dir := util.TmpDir()
	defer func() {
		_ = os.RemoveAll(dir)
	}()

	indexMapping, err := mapping.NewIndexMappingFromFile(filepath.Join(curDir, "../examples/example_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}

	logger := log.NewLogger("WARN", "", 500, 3, 30, false)

	server, err := NewRaftServer(id, raftAddress, dir, indexMapping, true, logger)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if err := server.Start(); err != nil {
		t.Fatalf("%v", err)
	}
	defer func() {
		if err := server.Stop(); err != nil {
			t.Fatalf("%v", err)
		}
	}()

	if err := server.WaitForDetectLeader(60 * time.Second); err != nil {
		t.Fatalf("%v", err)
	}

	node := &protobuf.Node{
		RaftAddress: raftAddress,
		Metadata: &protobuf.Metadata{
			GrpcAddress: grpcAddress,
			HttpAddress: httpAddress,
		},
	}

	if err := server.Join(id, node); err != nil {
		switch err {
		case errors.ErrNodeAlreadyExists:
			// ok
		default:
			t.Fatalf("%v", err)
		}
	}

	// ----------

	id2 := "node2"
	raftAddress2 := fmt.Sprintf(":%d", util.TmpPort())
	grpcAddress2 := fmt.Sprintf(":%d", util.TmpPort())
	httpAddress2 := fmt.Sprintf(":%d", util.TmpPort())

	dir2 := util.TmpDir()
	defer func() {
		_ = os.RemoveAll(dir2)
	}()

	indexMapping2, err := mapping.NewIndexMappingFromFile(filepath.Join(curDir, "../examples/example_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}

	logger2 := log.NewLogger("WARN", "", 500, 3, 30, false)

	server2, err := NewRaftServer(id2, raftAddress2, dir2, indexMapping2, false, logger2)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if err := server2.Start(); err != nil {
		t.Fatalf("%v", err)
	}
	defer func() {
		if err := server2.Stop(); err != nil {
			t.Fatalf("%v", err)
		}
	}()

	node2 := &protobuf.Node{
		RaftAddress: raftAddress2,
		Metadata: &protobuf.Metadata{
			GrpcAddress: grpcAddress2,
			HttpAddress: httpAddress2,
		},
	}

	if err := server.Join(id2, node2); err != nil {
		switch err {
		case errors.ErrNodeAlreadyExists:
			// ok
		default:
			t.Fatalf("%v", err)
		}
	}

	// ----------

	id3 := "node3"
	raftAddress3 := fmt.Sprintf(":%d", util.TmpPort())
	grpcAddress3 := fmt.Sprintf(":%d", util.TmpPort())
	httpAddress3 := fmt.Sprintf(":%d", util.TmpPort())

	dir3 := util.TmpDir()
	defer func() {
		_ = os.RemoveAll(dir3)
	}()

	indexMapping3, err := mapping.NewIndexMappingFromFile(filepath.Join(curDir, "../examples/example_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}

	logger3 := log.NewLogger("WARN", "", 500, 3, 30, false)

	server3, err := NewRaftServer(id3, raftAddress3, dir3, indexMapping3, false, logger3)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if err := server3.Start(); err != nil {
		t.Fatalf("%v", err)
	}
	defer func() {
		if err := server3.Stop(); err != nil {
			t.Fatalf("%v", err)
		}
	}()

	node3 := &protobuf.Node{
		RaftAddress: raftAddress3,
		Metadata: &protobuf.Metadata{
			GrpcAddress: grpcAddress3,
			HttpAddress: httpAddress3,
		},
	}

	if err := server.Join(id3, node3); err != nil {
		switch err {
		case errors.ErrNodeAlreadyExists:
			// ok
		default:
			t.Fatalf("%v", err)
		}
	}

	ns, err := server.Nodes()
	if err != nil {
		t.Fatalf("%v", err)
	}
	if 3 != len(ns) {
		t.Fatalf("expected content to see %v, saw %v", 3, len(ns))
	}
	if raftAddress != ns[id].RaftAddress {
		t.Fatalf("expected content to see %v, saw %v", raftAddress, ns[id].RaftAddress)
	}
	if grpcAddress != ns[id].Metadata.GrpcAddress {
		t.Fatalf("expected content to see %v, saw %v", grpcAddress, ns[id].Metadata.GrpcAddress)
	}
	if httpAddress != ns[id].Metadata.HttpAddress {
		t.Fatalf("expected content to see %v, saw %v", httpAddress, ns[id].Metadata.HttpAddress)
	}
	if raftAddress2 != ns[id2].RaftAddress {
		t.Fatalf("expected content to see %v, saw %v", raftAddress2, ns[id2].RaftAddress)
	}
	if grpcAddress2 != ns[id2].Metadata.GrpcAddress {
		t.Fatalf("expected content to see %v, saw %v", grpcAddress2, ns[id2].Metadata.GrpcAddress)
	}
	if httpAddress2 != ns[id2].Metadata.HttpAddress {
		t.Fatalf("expected content to see %v, saw %v", httpAddress2, ns[id2].Metadata.HttpAddress)
	}
	if raftAddress3 != ns[id3].RaftAddress {
		t.Fatalf("expected content to see %v, saw %v", raftAddress3, ns[id3].RaftAddress)
	}
	if grpcAddress3 != ns[id3].Metadata.GrpcAddress {
		t.Fatalf("expected content to see %v, saw %v", grpcAddress3, ns[id3].Metadata.GrpcAddress)
	}
	if httpAddress3 != ns[id3].Metadata.HttpAddress {
		t.Fatalf("expected content to see %v, saw %v", httpAddress3, ns[id3].Metadata.HttpAddress)
	}

	time.Sleep(3 * time.Second)

	ns2, err := server2.Nodes()
	if err != nil {
		t.Fatalf("%v", err)
	}
	if 3 != len(ns2) {
		t.Fatalf("expected content to see %v, saw %v", 3, len(ns2))
	}
	if raftAddress != ns2[id].RaftAddress {
		t.Fatalf("expected content to see %v, saw %v", raftAddress, ns2[id].RaftAddress)
	}
	if grpcAddress != ns2[id].Metadata.GrpcAddress {
		t.Fatalf("expected content to see %v, saw %v", grpcAddress, ns2[id].Metadata.GrpcAddress)
	}
	if httpAddress != ns2[id].Metadata.HttpAddress {
		t.Fatalf("expected content to see %v, saw %v", httpAddress, ns2[id].Metadata.HttpAddress)
	}
	if raftAddress2 != ns2[id2].RaftAddress {
		t.Fatalf("expected content to see %v, saw %v", raftAddress2, ns2[id2].RaftAddress)
	}
	if grpcAddress2 != ns2[id2].Metadata.GrpcAddress {
		t.Fatalf("expected content to see %v, saw %v", grpcAddress2, ns2[id2].Metadata.GrpcAddress)
	}
	if httpAddress2 != ns2[id2].Metadata.HttpAddress {
		t.Fatalf("expected content to see %v, saw %v", httpAddress2, ns2[id2].Metadata.HttpAddress)
	}
	if raftAddress3 != ns2[id3].RaftAddress {
		t.Fatalf("expected content to see %v, saw %v", raftAddress3, ns2[id3].RaftAddress)
	}
	if grpcAddress3 != ns2[id3].Metadata.GrpcAddress {
		t.Fatalf("expected content to see %v, saw %v", grpcAddress3, ns2[id3].Metadata.GrpcAddress)
	}
	if httpAddress3 != ns2[id3].Metadata.HttpAddress {
		t.Fatalf("expected content to see %v, saw %v", httpAddress3, ns2[id3].Metadata.HttpAddress)
	}

	time.Sleep(3 * time.Second)

	ns3, err := server3.Nodes()
	if err != nil {
		t.Fatalf("%v", err)
	}
	if 3 != len(ns3) {
		t.Fatalf("expected content to see %v, saw %v", 3, len(ns3))
	}
	if raftAddress != ns3[id].RaftAddress {
		t.Fatalf("expected content to see %v, saw %v", raftAddress, ns3[id].RaftAddress)
	}
	if grpcAddress != ns3[id].Metadata.GrpcAddress {
		t.Fatalf("expected content to see %v, saw %v", grpcAddress, ns3[id].Metadata.GrpcAddress)
	}
	if httpAddress != ns3[id].Metadata.HttpAddress {
		t.Fatalf("expected content to see %v, saw %v", httpAddress, ns3[id].Metadata.HttpAddress)
	}
	if raftAddress2 != ns3[id2].RaftAddress {
		t.Fatalf("expected content to see %v, saw %v", raftAddress2, ns3[id2].RaftAddress)
	}
	if grpcAddress2 != ns3[id2].Metadata.GrpcAddress {
		t.Fatalf("expected content to see %v, saw %v", grpcAddress2, ns3[id2].Metadata.GrpcAddress)
	}
	if httpAddress2 != ns3[id2].Metadata.HttpAddress {
		t.Fatalf("expected content to see %v, saw %v", httpAddress2, ns3[id2].Metadata.HttpAddress)
	}
	if raftAddress3 != ns3[id3].RaftAddress {
		t.Fatalf("expected content to see %v, saw %v", raftAddress3, ns3[id3].RaftAddress)
	}
	if grpcAddress3 != ns3[id3].Metadata.GrpcAddress {
		t.Fatalf("expected content to see %v, saw %v", grpcAddress3, ns3[id3].Metadata.GrpcAddress)
	}
	if httpAddress3 != ns3[id3].Metadata.HttpAddress {
		t.Fatalf("expected content to see %v, saw %v", httpAddress3, ns3[id3].Metadata.HttpAddress)
	}

	if err := server.Leave(id3); err != nil {
		t.Fatalf("%v", err)
	}

	ns, err = server.Nodes()
	if err != nil {
		t.Fatalf("%v", err)
	}
	if 2 != len(ns) {
		t.Fatalf("expected content to see %v, saw %v", 2, len(ns))
	}
	if raftAddress != ns[id].RaftAddress {
		t.Fatalf("expected content to see %v, saw %v", raftAddress, ns[id].RaftAddress)
	}
	if grpcAddress != ns[id].Metadata.GrpcAddress {
		t.Fatalf("expected content to see %v, saw %v", grpcAddress, ns[id].Metadata.GrpcAddress)
	}
	if httpAddress != ns[id].Metadata.HttpAddress {
		t.Fatalf("expected content to see %v, saw %v", httpAddress, ns[id].Metadata.HttpAddress)
	}
	if raftAddress2 != ns[id2].RaftAddress {
		t.Fatalf("expected content to see %v, saw %v", raftAddress2, ns[id2].RaftAddress)
	}
	if grpcAddress2 != ns[id2].Metadata.GrpcAddress {
		t.Fatalf("expected content to see %v, saw %v", grpcAddress2, ns[id2].Metadata.GrpcAddress)
	}
	if httpAddress2 != ns[id2].Metadata.HttpAddress {
		t.Fatalf("expected content to see %v, saw %v", httpAddress2, ns[id2].Metadata.HttpAddress)
	}
	if _, ok := ns[id3]; ok {
		t.Fatalf("expected content to see %v, saw %v", false, ok)
	}

	time.Sleep(3 * time.Second)

	ns2, err = server2.Nodes()
	if err != nil {
		t.Fatalf("%v", err)
	}
	if 2 != len(ns2) {
		t.Fatalf("expected content to see %v, saw %v", 2, len(ns2))
	}
	if raftAddress != ns2[id].RaftAddress {
		t.Fatalf("expected content to see %v, saw %v", raftAddress, ns2[id].RaftAddress)
	}
	if grpcAddress != ns2[id].Metadata.GrpcAddress {
		t.Fatalf("expected content to see %v, saw %v", grpcAddress, ns2[id].Metadata.GrpcAddress)
	}
	if httpAddress != ns2[id].Metadata.HttpAddress {
		t.Fatalf("expected content to see %v, saw %v", httpAddress, ns2[id].Metadata.HttpAddress)
	}
	if raftAddress2 != ns2[id2].RaftAddress {
		t.Fatalf("expected content to see %v, saw %v", raftAddress2, ns2[id2].RaftAddress)
	}
	if grpcAddress2 != ns2[id2].Metadata.GrpcAddress {
		t.Fatalf("expected content to see %v, saw %v", grpcAddress2, ns2[id2].Metadata.GrpcAddress)
	}
	if httpAddress2 != ns2[id2].Metadata.HttpAddress {
		t.Fatalf("expected content to see %v, saw %v", httpAddress2, ns2[id2].Metadata.HttpAddress)
	}
	if _, ok := ns2[id3]; ok {
		t.Fatalf("expected content to see %v, saw %v", false, ok)
	}
}

func Test_RaftServer_Set(t *testing.T) {
	curDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("%v", err)
	}

	rafAddress := fmt.Sprintf(":%d", util.TmpPort())

	dir := util.TmpDir()
	defer func() {
		_ = os.RemoveAll(dir)
	}()

	indexMapping, err := mapping.NewIndexMappingFromFile(filepath.Join(curDir, "../examples/example_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}

	logger := log.NewLogger("WARN", "", 500, 3, 30, false)

	server, err := NewRaftServer("node1", rafAddress, dir, indexMapping, true, logger)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if err := server.Start(); err != nil {
		t.Fatalf("%v", err)
	}
	defer func() {
		if err := server.Stop(); err != nil {
			t.Fatalf("%v", err)
		}
	}()

	if err := server.WaitForDetectLeader(10 * time.Second); err != nil {
		t.Fatalf("%v", err)
	}

	docId1 := "1"
	docFieldsMap1 := map[string]interface{}{
		"title":     "Search engine (computing)",
		"text":      "A search engine is an information retrieval system designed to help find information stored on a computer system. The search results are usually presented in a list and are commonly called hits. Search engines help to minimize the time required to find information and the amount of information which must be consulted, akin to other techniques for managing information overload. The most public, visible form of a search engine is a Web search engine which searches for information on the World Wide Web.",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"_type":     "example",
	}
	docFields1, err := json.Marshal(docFieldsMap1)
	if err != nil {
		t.Fatalf("%v", err)
	}

	setReq1 := &protobuf.SetRequest{
		Id:     docId1,
		Fields: docFields1,
	}

	if err := server.Set(setReq1); err != nil {
		t.Fatalf("%v", err)
	}
}

func Test_RaftServer_Get(t *testing.T) {
	curDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("%v", err)
	}

	rafAddress := fmt.Sprintf(":%d", util.TmpPort())

	dir := util.TmpDir()
	defer func() {
		_ = os.RemoveAll(dir)
	}()

	indexMapping, err := mapping.NewIndexMappingFromFile(filepath.Join(curDir, "../examples/example_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}

	logger := log.NewLogger("WARN", "", 500, 3, 30, false)

	server, err := NewRaftServer("node1", rafAddress, dir, indexMapping, true, logger)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if err := server.Start(); err != nil {
		t.Fatalf("%v", err)
	}
	defer func() {
		if err := server.Stop(); err != nil {
			t.Fatalf("%v", err)
		}
	}()

	if err := server.WaitForDetectLeader(10 * time.Second); err != nil {
		t.Fatalf("%v", err)
	}

	docId1 := "1"
	docFieldsMap1 := map[string]interface{}{
		"title":     "Search engine (computing)",
		"text":      "A search engine is an information retrieval system designed to help find information stored on a computer system. The search results are usually presented in a list and are commonly called hits. Search engines help to minimize the time required to find information and the amount of information which must be consulted, akin to other techniques for managing information overload. The most public, visible form of a search engine is a Web search engine which searches for information on the World Wide Web.",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"_type":     "example",
	}
	docFields1, err := json.Marshal(docFieldsMap1)
	if err != nil {
		t.Fatalf("%v", err)
	}

	setReq1 := &protobuf.SetRequest{
		Id:     docId1,
		Fields: docFields1,
	}

	if err := server.Set(setReq1); err != nil {
		t.Fatalf("%v", err)
	}

	f1, err := server.Get(docId1)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if docFieldsMap1["title"] != f1["title"] {
		t.Fatalf("expected content to see %v, saw %v", docFieldsMap1["title"], f1["title"])
	}
}

func Test_RaftServer_Delete(t *testing.T) {
	curDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("%v", err)
	}

	rafAddress := fmt.Sprintf(":%d", util.TmpPort())

	dir := util.TmpDir()
	defer func() {
		_ = os.RemoveAll(dir)
	}()

	indexMapping, err := mapping.NewIndexMappingFromFile(filepath.Join(curDir, "../examples/example_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}

	logger := log.NewLogger("WARN", "", 500, 3, 30, false)

	server, err := NewRaftServer("node1", rafAddress, dir, indexMapping, true, logger)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if err := server.Start(); err != nil {
		t.Fatalf("%v", err)
	}
	defer func() {
		if err := server.Stop(); err != nil {
			t.Fatalf("%v", err)
		}
	}()

	if err := server.WaitForDetectLeader(10 * time.Second); err != nil {
		t.Fatalf("%v", err)
	}

	docId1 := "1"
	docFieldsMap1 := map[string]interface{}{
		"title":     "Search engine (computing)",
		"text":      "A search engine is an information retrieval system designed to help find information stored on a computer system. The search results are usually presented in a list and are commonly called hits. Search engines help to minimize the time required to find information and the amount of information which must be consulted, akin to other techniques for managing information overload. The most public, visible form of a search engine is a Web search engine which searches for information on the World Wide Web.",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"_type":     "example",
	}
	docFields1, err := json.Marshal(docFieldsMap1)
	if err != nil {
		t.Fatalf("%v", err)
	}

	setReq1 := &protobuf.SetRequest{
		Id:     docId1,
		Fields: docFields1,
	}

	if err := server.Set(setReq1); err != nil {
		t.Fatalf("%v", err)
	}

	f1, err := server.Get(docId1)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if docFieldsMap1["title"] != f1["title"] {
		t.Fatalf("expected content to see %v, saw %v", docFieldsMap1["title"], f1["title"])
	}

	deleteReq1 := &protobuf.DeleteRequest{
		Id: docId1,
	}

	if err := server.Delete(deleteReq1); err != nil {
		t.Fatalf("%v", err)
	}

	f1, err = server.Get(docId1)
	if err != nil {
		switch err {
		case errors.ErrNotFound:
			//ok
		default:
			t.Fatalf("%v", err)
		}
	}
	if f1 != nil {
		t.Fatalf("expected content to see %v, saw %v", nil, f1)
	}
}

func Test_RaftServer_Snapshot(t *testing.T) {
	curDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("%v", err)
	}

	rafAddress := fmt.Sprintf(":%d", util.TmpPort())

	dir := util.TmpDir()
	defer func() {
		_ = os.RemoveAll(dir)
	}()

	indexMapping, err := mapping.NewIndexMappingFromFile(filepath.Join(curDir, "../examples/example_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}

	logger := log.NewLogger("WARN", "", 500, 3, 30, false)

	server, err := NewRaftServer("node1", rafAddress, dir, indexMapping, true, logger)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if err := server.Start(); err != nil {
		t.Fatalf("%v", err)
	}
	defer func() {
		if err := server.Stop(); err != nil {
			t.Fatalf("%v", err)
		}
	}()

	if err := server.WaitForDetectLeader(10 * time.Second); err != nil {
		t.Fatalf("%v", err)
	}

	docId1 := "1"
	docFieldsMap1 := map[string]interface{}{
		"title":     "Search engine (computing)",
		"text":      "A search engine is an information retrieval system designed to help find information stored on a computer system. The search results are usually presented in a list and are commonly called hits. Search engines help to minimize the time required to find information and the amount of information which must be consulted, akin to other techniques for managing information overload. The most public, visible form of a search engine is a Web search engine which searches for information on the World Wide Web.",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"_type":     "example",
	}
	docFields1, err := json.Marshal(docFieldsMap1)
	if err != nil {
		t.Fatalf("%v", err)
	}

	setReq1 := &protobuf.SetRequest{
		Id:     docId1,
		Fields: docFields1,
	}

	if err := server.Set(setReq1); err != nil {
		t.Fatalf("%v", err)
	}

	if err := server.Snapshot(); err != nil {
		t.Fatalf("%v", err)
	}
}
