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
	"github.com/mosuka/blast/raft"
	"github.com/urfave/cli"
)

var (
	flBindAddr = cli.StringFlag{
		Name:   "bind-addr",
		Value:  "127.0.0.1:10000",
		Usage:  "Address to listen on for peer traffic",
		EnvVar: "BLAST_BIND_ADDR",
	}
	flGRPCAddr = cli.StringFlag{
		Name:   "grpc-addr",
		Value:  "127.0.0.1:10001",
		Usage:  "Address to listen on for client traffic via gRPC",
		EnvVar: "BLAST_GRPC_ADDR",
	}
	flHTTPAddr = cli.StringFlag{
		Name:   "http-addr",
		Value:  "127.0.0.1:10002",
		Usage:  "Address to listen on for client traffic via HTTP",
		EnvVar: "BLAST_HTTP_ADDR",
	}

	flRaftNodeID = cli.StringFlag{
		Name:   "raft-node-id",
		Value:  raft.DefaultNodeID,
		Usage:  "Node ID",
		EnvVar: "BLAST_RAFT_NODE_ID",
	}

	flPeerGRPCAddr = cli.StringFlag{
		Name:   "peer-grpc-addr",
		Usage:  "Peer gRPC address to connect on for join the cluster",
		EnvVar: "BLAST_PEER_GRPC_ADDR",
	}

	flBatchSize = cli.IntFlag{
		Name:   "batch-size",
		Value:  1000,
		Usage:  "Batch size for bulk update",
		EnvVar: "BLAST_BATCH_SIZE",
	}

	flPrettyPrint = cli.BoolFlag{
		Name:   "pretty-print",
		Usage:  "Pretty print JSON",
		EnvVar: "BLAST_PRETTY_PRINT",
	}
)
