// Copyright (c) 2018 Minoru Osuka
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// 		http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"github.com/urfave/cli"
)

var (
	flGRPCAddr = cli.StringFlag{
		Name:   "grpc-addr",
		Value:  ":10001",
		Usage:  "gRPC address to connect to",
		EnvVar: "BLAST_GRPC_ADDR",
	}

	flTargetRaftNodeID = cli.StringFlag{
		Name:  "target-raft-node-id",
		Usage: "Target node ID",
	}
	flTargetRaftAddr = cli.StringFlag{
		Name:  "target-raft-addr",
		Usage: "Target raft address",
	}
	flTargetGRPCAddr = cli.StringFlag{
		Name:  "target-grpc-addr",
		Usage: "Target gRPC address",
	}
	flTargetHTTPAddr = cli.StringFlag{
		Name:  "target-http-addr",
		Usage: "Target HTTP address",
	}

	flBatchSize = cli.IntFlag{
		Name:   "batch-size",
		Value:  1000,
		Usage:  "Batch size for bulk update",
		EnvVar: "BLAST_BATCH_SIZE",
	}
)
