package server

import (
	"fmt"
	"testing"
	"time"

	"github.com/mosuka/blast/log"
	"github.com/mosuka/blast/util"
)

func Test_GRPCGateway_Start_Stop(t *testing.T) {
	httpAddress := fmt.Sprintf(":%d", util.TmpPort())
	grpcAddress := fmt.Sprintf(":%d", util.TmpPort())
	certificateFile := ""
	KeyFile := ""
	commonName := ""
	corsAllowedMethods := make([]string, 0)
	corsAllowedOrigins := make([]string, 0)
	corsAllowedHeaders := make([]string, 0)
	logger := log.NewLogger("WARN", "", 500, 3, 30, false)

	grpcGateway, err := NewGRPCGateway(httpAddress, grpcAddress, certificateFile, KeyFile, commonName, corsAllowedMethods, corsAllowedOrigins, corsAllowedHeaders, logger)
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer func() {
		if err := grpcGateway.Stop(); err != nil {
			t.Fatalf("%v", err)
		}
	}()

	if err := grpcGateway.Start(); err != nil {
		t.Fatalf("%v", err)
	}

	time.Sleep(3 * time.Second)
}
