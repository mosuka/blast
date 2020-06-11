package server

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/hashicorp/raft"

	"github.com/mosuka/blast/protobuf"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/mosuka/blast/log"
	"github.com/mosuka/blast/mapping"
	"github.com/mosuka/blast/util"
)

func Test_GRPCService_Start_Stop(t *testing.T) {
	curDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("%v", err)
	}

	tmpDir := util.TmpDir()
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	logger := log.NewLogger("WARN", "", 500, 3, 30, false)

	// Raft server
	rafAddress := fmt.Sprintf(":%d", util.TmpPort())
	dir := util.TmpDir()
	defer func() {
		_ = os.RemoveAll(dir)
	}()
	indexMapping, err := mapping.NewIndexMappingFromFile(filepath.Join(curDir, "../examples/example_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}
	raftServer, err := NewRaftServer("node1", rafAddress, dir, indexMapping, true, logger)
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer func() {
		if err := raftServer.Stop(); err != nil {
			t.Fatalf("%v", err)
		}
	}()
	if err := raftServer.Start(); err != nil {
		t.Fatalf("%v", err)
	}

	// gRPC service
	certificateFile := ""
	commonName := ""

	grpcService, err := NewGRPCService(raftServer, certificateFile, commonName, logger)
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer func() {
		if err := grpcService.Stop(); err != nil {
			t.Fatalf("%v", err)
		}
	}()

	if err := grpcService.Start(); err != nil {
		t.Fatalf("%v", err)
	}

	time.Sleep(3 * time.Second)
}

func Test_GRPCService_LivenessCheck(t *testing.T) {
	curDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("%v", err)
	}

	tmpDir := util.TmpDir()
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	logger := log.NewLogger("WARN", "", 500, 3, 30, false)

	// Raft server
	rafAddress := fmt.Sprintf(":%d", util.TmpPort())
	dir := util.TmpDir()
	defer func() {
		_ = os.RemoveAll(dir)
	}()
	indexMapping, err := mapping.NewIndexMappingFromFile(filepath.Join(curDir, "../examples/example_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}
	raftServer, err := NewRaftServer("node1", rafAddress, dir, indexMapping, true, logger)
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer func() {
		if err := raftServer.Stop(); err != nil {
			t.Fatalf("%v", err)
		}
	}()
	if err := raftServer.Start(); err != nil {
		t.Fatalf("%v", err)
	}

	// gRPC service
	certificateFile := ""
	commonName := ""

	grpcService, err := NewGRPCService(raftServer, certificateFile, commonName, logger)
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer func() {
		if err := grpcService.Stop(); err != nil {
			t.Fatalf("%v", err)
		}
	}()

	if err := grpcService.Start(); err != nil {
		t.Fatalf("%v", err)
	}

	time.Sleep(3 * time.Second)

	ctx := context.Background()
	req := &empty.Empty{}

	resp, err := grpcService.LivenessCheck(ctx, req)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if !resp.Alive {
		t.Fatalf("expected content to see %v, saw %v", true, resp.Alive)
	}
}

func Test_GRPCService_ReadinessCheck(t *testing.T) {
	curDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("%v", err)
	}

	tmpDir := util.TmpDir()
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	logger := log.NewLogger("WARN", "", 500, 3, 30, false)

	// Raft server
	rafAddress := fmt.Sprintf(":%d", util.TmpPort())
	dir := util.TmpDir()
	defer func() {
		_ = os.RemoveAll(dir)
	}()
	indexMapping, err := mapping.NewIndexMappingFromFile(filepath.Join(curDir, "../examples/example_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}
	raftServer, err := NewRaftServer("node1", rafAddress, dir, indexMapping, true, logger)
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer func() {
		if err := raftServer.Stop(); err != nil {
			t.Fatalf("%v", err)
		}
	}()
	if err := raftServer.Start(); err != nil {
		t.Fatalf("%v", err)
	}

	// gRPC service
	certificateFile := ""
	commonName := ""

	grpcService, err := NewGRPCService(raftServer, certificateFile, commonName, logger)
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer func() {
		if err := grpcService.Stop(); err != nil {
			t.Fatalf("%v", err)
		}
	}()

	if err := grpcService.Start(); err != nil {
		t.Fatalf("%v", err)
	}

	time.Sleep(3 * time.Second)

	ctx := context.Background()
	req := &empty.Empty{}

	resp, err := grpcService.ReadinessCheck(ctx, req)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if !resp.Ready {
		t.Fatalf("expected content to see %v, saw %v", true, resp.Ready)
	}
}

func Test_GRPCService_Join(t *testing.T) {
	curDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("%v", err)
	}

	tmpDir := util.TmpDir()
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	logger := log.NewLogger("WARN", "", 500, 3, 30, false)

	// Raft server
	raftAddress := fmt.Sprintf(":%d", util.TmpPort())
	dir := util.TmpDir()
	defer func() {
		_ = os.RemoveAll(dir)
	}()
	indexMapping, err := mapping.NewIndexMappingFromFile(filepath.Join(curDir, "../examples/example_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}
	raftServer, err := NewRaftServer("node1", raftAddress, dir, indexMapping, true, logger)
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer func() {
		if err := raftServer.Stop(); err != nil {
			t.Fatalf("%v", err)
		}
	}()
	if err := raftServer.Start(); err != nil {
		t.Fatalf("%v", err)
	}

	// gRPC service
	certificateFile := ""
	commonName := ""

	grpcService, err := NewGRPCService(raftServer, certificateFile, commonName, logger)
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer func() {
		if err := grpcService.Stop(); err != nil {
			t.Fatalf("%v", err)
		}
	}()

	if err := grpcService.Start(); err != nil {
		t.Fatalf("%v", err)
	}

	time.Sleep(3 * time.Second)

	grpcAddress := fmt.Sprintf(":%d", util.TmpPort())
	httpAddress := fmt.Sprintf(":%d", util.TmpPort())

	ctx := context.Background()
	req := &protobuf.JoinRequest{
		Id: "node1",
		Node: &protobuf.Node{
			RaftAddress: raftAddress,
			Metadata: &protobuf.Metadata{
				GrpcAddress: grpcAddress,
				HttpAddress: httpAddress,
			},
		},
	}

	_, err = grpcService.Join(ctx, req)
	if err != nil {
		t.Fatalf("%v", err)
	}
}

func Test_GRPCService_Node(t *testing.T) {
	curDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("%v", err)
	}

	tmpDir := util.TmpDir()
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	logger := log.NewLogger("WARN", "", 500, 3, 30, false)

	// Raft server
	raftAddress := fmt.Sprintf(":%d", util.TmpPort())
	dir := util.TmpDir()
	defer func() {
		_ = os.RemoveAll(dir)
	}()
	indexMapping, err := mapping.NewIndexMappingFromFile(filepath.Join(curDir, "../examples/example_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}
	raftServer, err := NewRaftServer("node1", raftAddress, dir, indexMapping, true, logger)
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer func() {
		if err := raftServer.Stop(); err != nil {
			t.Fatalf("%v", err)
		}
	}()
	if err := raftServer.Start(); err != nil {
		t.Fatalf("%v", err)
	}

	// gRPC service
	certificateFile := ""
	commonName := ""

	grpcService, err := NewGRPCService(raftServer, certificateFile, commonName, logger)
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer func() {
		if err := grpcService.Stop(); err != nil {
			t.Fatalf("%v", err)
		}
	}()

	if err := grpcService.Start(); err != nil {
		t.Fatalf("%v", err)
	}

	time.Sleep(3 * time.Second)

	grpcAddress := fmt.Sprintf(":%d", util.TmpPort())
	httpAddress := fmt.Sprintf(":%d", util.TmpPort())

	ctx := context.Background()
	req := &protobuf.JoinRequest{
		Id: "node1",
		Node: &protobuf.Node{
			RaftAddress: raftAddress,
			Metadata: &protobuf.Metadata{
				GrpcAddress: grpcAddress,
				HttpAddress: httpAddress,
			},
		},
	}

	_, err = grpcService.Join(ctx, req)
	if err != nil {
		t.Fatalf("%v", err)
	}

	resp, err := grpcService.Node(ctx, &empty.Empty{})
	if err != nil {
		t.Fatalf("%v", err)
	}

	if raftAddress != resp.Node.RaftAddress {
		t.Fatalf("expected content to see %v, saw %v", raftAddress, resp.Node.RaftAddress)
	}

	if grpcAddress != resp.Node.Metadata.GrpcAddress {
		t.Fatalf("expected content to see %v, saw %v", grpcAddress, resp.Node.Metadata.GrpcAddress)
	}

	if httpAddress != resp.Node.Metadata.HttpAddress {
		t.Fatalf("expected content to see %v, saw %v", grpcAddress, resp.Node.Metadata.HttpAddress)
	}

	if raft.Leader.String() != resp.Node.State {
		t.Fatalf("expected content to see %v, saw %v", raft.Leader.String(), resp.Node.State)
	}
}

//func Test_GRPCService_Leave(t *testing.T) {
//	curDir, err := os.Getwd()
//	if err != nil {
//		t.Fatalf("%v", err)
//	}
//
//	tmpDir := util.TmpDir()
//	defer func() {
//		_ = os.RemoveAll(tmpDir)
//	}()
//
//	logger := log.NewLogger("WARN", "", 500, 3, 30, false)
//
//	certificateFile := ""
//	commonName := ""
//
//	raftAddress1 := fmt.Sprintf(":%d", util.TmpPort())
//	grpcAddress1 := fmt.Sprintf(":%d", util.TmpPort())
//	httpAddress1 := fmt.Sprintf(":%d", util.TmpPort())
//
//	raftAddress2 := fmt.Sprintf(":%d", util.TmpPort())
//	grpcAddress2 := fmt.Sprintf(":%d", util.TmpPort())
//	httpAddress2 := fmt.Sprintf(":%d", util.TmpPort())
//
//	raftAddress3 := fmt.Sprintf(":%d", util.TmpPort())
//	grpcAddress3 := fmt.Sprintf(":%d", util.TmpPort())
//	httpAddress3 := fmt.Sprintf(":%d", util.TmpPort())
//
//	dir1 := util.TmpDir()
//	defer func() {
//		_ = os.RemoveAll(dir1)
//	}()
//	dir2 := util.TmpDir()
//	defer func() {
//		_ = os.RemoveAll(dir2)
//	}()
//	dir3 := util.TmpDir()
//	defer func() {
//		_ = os.RemoveAll(dir3)
//	}()
//
//	indexMapping, err := mapping.NewIndexMappingFromFile(filepath.Join(curDir, "../examples/example_mapping.json"))
//	if err != nil {
//		t.Fatalf("%v", err)
//	}
//
//	// Raft server
//	raftServer1, err := NewRaftServer("node1", raftAddress1, dir1, indexMapping, true, logger)
//	if err != nil {
//		t.Fatalf("%v", err)
//	}
//	defer func() {
//		if err := raftServer1.Stop(); err != nil {
//			t.Fatalf("%v", err)
//		}
//	}()
//	if err := raftServer1.Start(); err != nil {
//		t.Fatalf("%v", err)
//	}
//
//	raftServer2, err := NewRaftServer("node2", raftAddress2, dir2, indexMapping, false, logger)
//	if err != nil {
//		t.Fatalf("%v", err)
//	}
//	defer func() {
//		if err := raftServer2.Stop(); err != nil {
//			t.Fatalf("%v", err)
//		}
//	}()
//	if err := raftServer2.Start(); err != nil {
//		t.Fatalf("%v", err)
//	}
//
//	raftServer3, err := NewRaftServer("node3", raftAddress3, dir3, indexMapping, false, logger)
//	if err != nil {
//		t.Fatalf("%v", err)
//	}
//	defer func() {
//		if err := raftServer3.Stop(); err != nil {
//			t.Fatalf("%v", err)
//		}
//	}()
//	if err := raftServer3.Start(); err != nil {
//		t.Fatalf("%v", err)
//	}
//
//	// gRPC service
//	grpcService1, err := NewGRPCService(raftServer1, certificateFile, commonName, logger)
//	if err != nil {
//		t.Fatalf("%v", err)
//	}
//	defer func() {
//		if err := grpcService1.Stop(); err != nil {
//			t.Fatalf("%v", err)
//		}
//	}()
//	if err := grpcService1.Start(); err != nil {
//		t.Fatalf("%v", err)
//	}
//
//	grpcService2, err := NewGRPCService(raftServer2, certificateFile, commonName, logger)
//	if err != nil {
//		t.Fatalf("%v", err)
//	}
//	defer func() {
//		if err := grpcService2.Stop(); err != nil {
//			t.Fatalf("%v", err)
//		}
//	}()
//	if err := grpcService2.Start(); err != nil {
//		t.Fatalf("%v", err)
//	}
//
//	grpcService3, err := NewGRPCService(raftServer3, certificateFile, commonName, logger)
//	if err != nil {
//		t.Fatalf("%v", err)
//	}
//	defer func() {
//		if err := grpcService3.Stop(); err != nil {
//			t.Fatalf("%v", err)
//		}
//	}()
//	if err := grpcService3.Start(); err != nil {
//		t.Fatalf("%v", err)
//	}
//
//	time.Sleep(3 * time.Second)
//
//	ctx := context.Background()
//	req1 := &protobuf.JoinRequest{
//		Id: "node1",
//		Node: &protobuf.Node{
//			RaftAddress: raftAddress1,
//			Metadata: &protobuf.Metadata{
//				GrpcAddress: grpcAddress1,
//				HttpAddress: httpAddress1,
//			},
//		},
//	}
//	_, err = grpcService1.Join(ctx, req1)
//	if err != nil {
//		t.Fatalf("%v", err)
//	}
//
//	req2 := &protobuf.JoinRequest{
//		Id: "node2",
//		Node: &protobuf.Node{
//			RaftAddress: raftAddress2,
//			Metadata: &protobuf.Metadata{
//				GrpcAddress: grpcAddress2,
//				HttpAddress: httpAddress2,
//			},
//		},
//	}
//	_, err = grpcService1.Join(ctx, req2)
//	if err != nil {
//		t.Fatalf("%v", err)
//	}
//
//	req3 := &protobuf.JoinRequest{
//		Id: "node3",
//		Node: &protobuf.Node{
//			RaftAddress: raftAddress3,
//			Metadata: &protobuf.Metadata{
//				GrpcAddress: grpcAddress3,
//				HttpAddress: httpAddress3,
//			},
//		},
//	}
//	_, err = grpcService1.Join(ctx, req3)
//	if err != nil {
//		t.Fatalf("%v", err)
//	}
//
//	resp, err := grpcService1.Cluster(ctx, &empty.Empty{})
//	if err != nil {
//		t.Fatalf("%v", err)
//	}
//
//	if "node1" != resp.Cluster.Leader {
//		t.Fatalf("expected content to see %v, saw %v", "node1", resp.Cluster.Leader)
//	}
//
//	//if raftAddress1 != resp..RaftAddress {
//	//	t.Fatalf("expected content to see %v, saw %v", raftAddress1, resp.Node.RaftAddress)
//	//}
//	//
//	//if grpcAddress1 != resp.Node.Metadata.GrpcAddress {
//	//	t.Fatalf("expected content to see %v, saw %v", grpcAddress1, resp.Node.Metadata.GrpcAddress)
//	//}
//	//
//	//if httpAddress1 != resp.Node.Metadata.HttpAddress {
//	//	t.Fatalf("expected content to see %v, saw %v", grpcAddress1, resp.Node.Metadata.HttpAddress)
//	//}
//	//
//	//if raft.Leader.String() != resp.Node.State {
//	//	t.Fatalf("expected content to see %v, saw %v", raft.Leader.String(), resp.Node.State)
//	//}
//}

//func Test_GRPCService_Cluster(t *testing.T) {
//	curDir, err := os.Getwd()
//	if err != nil {
//		t.Fatalf("%v", err)
//	}
//
//	tmpDir := util.TmpDir()
//	defer func() {
//		_ = os.RemoveAll(tmpDir)
//	}()
//
//	// Raft server
//	raftAddress1 := fmt.Sprintf(":%d", util.TmpPort())
//	dir1 := util.TmpDir()
//	defer func() {
//		_ = os.RemoveAll(dir1)
//	}()
//	indexMapping1, err := mapping.NewIndexMappingFromFile(filepath.Join(curDir, "../examples/example_mapping.json"))
//	if err != nil {
//		t.Fatalf("%v", err)
//	}
//	logger1 := log.NewLogger("WARN", "", 500, 3, 30, false)
//	raftServer1, err := NewRaftServer("node1", raftAddress1, dir1, indexMapping1, true, logger1)
//	if err != nil {
//		t.Fatalf("%v", err)
//	}
//	defer func() {
//		if err := raftServer1.Stop(); err != nil {
//			t.Fatalf("%v", err)
//		}
//	}()
//	if err := raftServer1.Start(); err != nil {
//		t.Fatalf("%v", err)
//	}
//
//	// gRPC service
//	certificateFile1 := ""
//	commonName1 := ""
//	grpcService1, err := NewGRPCService(raftServer1, certificateFile1, commonName1, logger1)
//	if err != nil {
//		t.Fatalf("%v", err)
//	}
//	defer func() {
//		if err := grpcService1.Stop(); err != nil {
//			t.Fatalf("%v", err)
//		}
//	}()
//
//	if err := grpcService1.Start(); err != nil {
//		t.Fatalf("%v", err)
//	}
//
//	time.Sleep(3 * time.Second)
//
//	grpcAddress1 := fmt.Sprintf(":%d", util.TmpPort())
//	httpAddress1 := fmt.Sprintf(":%d", util.TmpPort())
//
//	ctx1 := context.Background()
//	joinReq1 := &protobuf.JoinRequest{
//		Id: "node1",
//		Node: &protobuf.Node{
//			RaftAddress: raftAddress1,
//			Metadata: &protobuf.Metadata{
//				GrpcAddress: grpcAddress1,
//				HttpAddress: httpAddress1,
//			},
//		},
//	}
//	_, err = grpcService1.Join(ctx1, joinReq1)
//	if err != nil {
//		t.Fatalf("%v", err)
//	}
//
//	// Raft server
//	raftAddress2 := fmt.Sprintf(":%d", util.TmpPort())
//	dir2 := util.TmpDir()
//	defer func() {
//		_ = os.RemoveAll(dir2)
//	}()
//	indexMapping2, err := mapping.NewIndexMappingFromFile(filepath.Join(curDir, "../examples/example_mapping.json"))
//	if err != nil {
//		t.Fatalf("%v", err)
//	}
//	logger2 := log.NewLogger("WARN", "", 500, 3, 30, false)
//	raftServer2, err := NewRaftServer("node2", raftAddress2, dir2, indexMapping2, false, logger2)
//	if err != nil {
//		t.Fatalf("%v", err)
//	}
//	defer func() {
//		if err := raftServer2.Stop(); err != nil {
//			t.Fatalf("%v", err)
//		}
//	}()
//	if err := raftServer2.Start(); err != nil {
//		t.Fatalf("%v", err)
//	}
//
//	// gRPC service
//	certificateFile2 := ""
//	commonName2 := ""
//	grpcService2, err := NewGRPCService(raftServer2, certificateFile2, commonName2, logger2)
//	if err != nil {
//		t.Fatalf("%v", err)
//	}
//	defer func() {
//		if err := grpcService2.Stop(); err != nil {
//			t.Fatalf("%v", err)
//		}
//	}()
//
//	if err := grpcService2.Start(); err != nil {
//		t.Fatalf("%v", err)
//	}
//
//	time.Sleep(3 * time.Second)
//
//	grpcAddress2 := fmt.Sprintf(":%d", util.TmpPort())
//	httpAddress2 := fmt.Sprintf(":%d", util.TmpPort())
//
//	ctx2 := context.Background()
//	joinReq2 := &protobuf.JoinRequest{
//		Id: "node2",
//		Node: &protobuf.Node{
//			RaftAddress: raftAddress2,
//			Metadata: &protobuf.Metadata{
//				GrpcAddress: grpcAddress2,
//				HttpAddress: httpAddress2,
//			},
//		},
//	}
//	_, err = grpcService1.Join(ctx2, joinReq2)
//	if err != nil {
//		t.Fatalf("%v", err)
//	}
//
//	// Raft server
//	raftAddress3 := fmt.Sprintf(":%d", util.TmpPort())
//	dir3 := util.TmpDir()
//	defer func() {
//		_ = os.RemoveAll(dir3)
//	}()
//	indexMapping3, err := mapping.NewIndexMappingFromFile(filepath.Join(curDir, "../examples/example_mapping.json"))
//	if err != nil {
//		t.Fatalf("%v", err)
//	}
//	logger3 := log.NewLogger("WARN", "", 500, 3, 30, false)
//	raftServer3, err := NewRaftServer("node3", raftAddress3, dir3, indexMapping3, false, logger3)
//	if err != nil {
//		t.Fatalf("%v", err)
//	}
//	defer func() {
//		if err := raftServer3.Stop(); err != nil {
//			t.Fatalf("%v", err)
//		}
//	}()
//	if err := raftServer3.Start(); err != nil {
//		t.Fatalf("%v", err)
//	}
//
//	// gRPC service
//	certificateFile3 := ""
//	commonName3 := ""
//	grpcService3, err := NewGRPCService(raftServer3, certificateFile3, commonName3, logger3)
//	if err != nil {
//		t.Fatalf("%v", err)
//	}
//	defer func() {
//		if err := grpcService3.Stop(); err != nil {
//			t.Fatalf("%v", err)
//		}
//	}()
//
//	if err := grpcService3.Start(); err != nil {
//		t.Fatalf("%v", err)
//	}
//
//	time.Sleep(3 * time.Second)
//
//	grpcAddress3 := fmt.Sprintf(":%d", util.TmpPort())
//	httpAddress3 := fmt.Sprintf(":%d", util.TmpPort())
//
//	ctx3 := context.Background()
//	joinReq3 := &protobuf.JoinRequest{
//		Id: "node3",
//		Node: &protobuf.Node{
//			RaftAddress: raftAddress3,
//			Metadata: &protobuf.Metadata{
//				GrpcAddress: grpcAddress3,
//				HttpAddress: httpAddress3,
//			},
//		},
//	}
//	_, err = grpcService1.Join(ctx3, joinReq3)
//	if err != nil {
//		t.Fatalf("%v", err)
//	}
//
//	respCluster1, err := grpcService1.Cluster(ctx1, &empty.Empty{})
//	if err != nil {
//		t.Fatalf("%v", err)
//	}
//	if 3 != len(respCluster1.Cluster.Nodes) {
//		t.Fatalf("expected content to see %v, saw %v", 3, len(respCluster1.Cluster.Nodes))
//	}
//
//	respCluster2, err := grpcService2.Cluster(ctx2, &empty.Empty{})
//	if err != nil {
//		t.Fatalf("%v", err)
//	}
//	if 3 != len(respCluster2.Cluster.Nodes) {
//		t.Fatalf("expected content to see %v, saw %v", 3, len(respCluster2.Cluster.Nodes))
//	}
//
//	respCluster3, err := grpcService2.Cluster(ctx3, &empty.Empty{})
//	if err != nil {
//		t.Fatalf("%v", err)
//	}
//	if 3 != len(respCluster3.Cluster.Nodes) {
//		t.Fatalf("expected content to see %v, saw %v", 3, len(respCluster3.Cluster.Nodes))
//	}
//}
