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

	// gRPC server
	grpcAddress := fmt.Sprintf(":%d", util.TmpPort())
	certificateFile := ""
	keyFile := ""
	commonName := ""

	grpcServer, err := NewGRPCServer(grpcAddress, raftServer, certificateFile, keyFile, commonName, logger)
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
