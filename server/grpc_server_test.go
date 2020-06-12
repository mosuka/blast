package server

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mosuka/blast/log"
	"github.com/mosuka/blast/mapping"
	"github.com/mosuka/blast/util"
)

func Test_GRPCServer_Start_Stop(t *testing.T) {
	curDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("%v", err)
	}

	tmpDir := util.TmpDir()
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	logger := log.NewLogger("WARN", "", 500, 3, 30, false)

	raftAddress := fmt.Sprintf(":%d", util.TmpPort())
	grpcAddress := fmt.Sprintf(":%d", util.TmpPort())

	dir := util.TmpDir()
	defer func() {
		_ = os.RemoveAll(dir)
	}()

	indexMapping, err := mapping.NewIndexMappingFromFile(filepath.Join(curDir, "../examples/example_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}

	// Raft server
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

	// gRPC server
	grpcServer, err := NewGRPCServer(grpcAddress, raftServer, logger)
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer func() {
		if err := grpcServer.Stop(); err != nil {
			t.Fatalf("%v", err)
		}
	}()

	if err := grpcServer.Start(); err != nil {
		t.Fatalf("%v", err)
	}

	time.Sleep(3 * time.Second)
}
