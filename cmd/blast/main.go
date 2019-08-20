// Copyright (c) 2019 Minoru Osuka
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
	"fmt"
	"os"
	"path"

	"github.com/blevesearch/bleve"
	"github.com/mosuka/blast/version"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = path.Base(os.Args[0])
	app.Usage = "Command for blast"
	app.Version = version.Version
	app.Authors = []cli.Author{
		{
			Name:  "mosuka",
			Email: "minoru.osuka@gmail.com",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:  "manager",
			Usage: "Command for blast manager",
			Subcommands: []cli.Command{
				{
					Name:  "start",
					Usage: "Start blast manager",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:   "peer-grpc-address",
							Value:  "",
							EnvVar: "BLAST_MANAGER_PEER_GRPC_ADDRESS",
							Usage:  "The gRPC address of the peer node that exists in the cluster to be joined",
						},
						cli.StringFlag{
							Name:   "grpc-address",
							Value:  ":5100",
							EnvVar: "BLAST_MANAGER_GRPC_ADDRESS",
							Usage:  "The gRPC listen address",
						},
						cli.StringFlag{
							Name:   "grpc-gateway-address",
							Value:  ":6100",
							EnvVar: "BLAST_MANAGER_GRPC_GATEWAY_ADDRESS",
							Usage:  "The gRPC gateway listen address",
						},
						cli.StringFlag{
							Name:   "http-address",
							Value:  ":8100",
							EnvVar: "BLAST_MANAGER_HTTP_ADDRESS",
							Usage:  "HTTP listen address",
						},
						cli.StringFlag{
							Name:   "node-id",
							Value:  "",
							EnvVar: "BLAST_MANAGER_NODE_ID",
							Usage:  "Unique ID to identify the node",
						},
						cli.StringFlag{
							Name:   "node-address",
							Value:  ":2100",
							EnvVar: "BLAST_MANAGER_NODE_ADDRESS",
							Usage:  "The address that should be bound to for internal cluster communications",
						},
						cli.StringFlag{
							Name:   "data-dir",
							Value:  "/tmp/blast/manager",
							EnvVar: "BLAST_MANAGER_DATA_DIR",
							Usage:  "A data directory for the node to store state",
						},
						cli.StringFlag{
							Name:   "raft-storage-type",
							Value:  "boltdb",
							EnvVar: "BLAST_MANAGER_RAFT_STORAGE_TYPE",
							Usage:  "Storage type of the database that stores the state",
						},
						cli.StringFlag{
							Name:   "index-mapping-file",
							Value:  "",
							EnvVar: "BLAST_MANAGER_INDEX_MAPPING_FILE",
							Usage:  "An index mapping file to use",
						},
						cli.StringFlag{
							Name:   "index-type",
							Value:  bleve.Config.DefaultIndexType,
							EnvVar: "BLAST_MANAGER_INDEX_TYPE",
							Usage:  "An index type to use",
						},
						cli.StringFlag{
							Name:   "index-storage-type",
							Value:  bleve.Config.DefaultKVStore,
							EnvVar: "BLAST_MANAGER_INDEX_STORAGE_TYPE",
							Usage:  "An index storage type to use",
						},
						cli.StringFlag{
							Name:   "log-level",
							Value:  "INFO",
							EnvVar: "BLAST_MANAGER_LOG_LEVEL",
							Usage:  "Log level",
						},
						cli.StringFlag{
							Name:   "log-file",
							Value:  os.Stderr.Name(),
							EnvVar: "BLAST_MANAGER_LOG_FILE",
							Usage:  "Log file",
						},
						cli.IntFlag{
							Name:   "log-max-size",
							Value:  500,
							EnvVar: "BLAST_MANAGER_LOG_MAX_SIZE",
							Usage:  "Max size of a log file (megabytes)",
						},
						cli.IntFlag{
							Name:   "log-max-backups",
							Value:  3,
							EnvVar: "BLAST_MANAGER_LOG_MAX_BACKUPS",
							Usage:  "Max backup count of log files",
						},
						cli.IntFlag{
							Name:   "log-max-age",
							Value:  30,
							EnvVar: "BLAST_MANAGER_LOG_MAX_AGE",
							Usage:  "Max age of a log file (days)",
						},
						cli.BoolFlag{
							Name:   "log-compress",
							EnvVar: "BLAST_MANAGER_LOG_COMPRESS",
							Usage:  "Compress a log file",
						},
						cli.StringFlag{
							Name:   "grpc-log-level",
							Value:  "WARN",
							EnvVar: "BLAST_MANAGER_GRPC_LOG_LEVEL",
							Usage:  "gRPC log level",
						},
						cli.StringFlag{
							Name:   "grpc-log-file",
							Value:  os.Stderr.Name(),
							EnvVar: "BLAST_MANAGER_GRPC_LOG_FILE",
							Usage:  "gRPC log file",
						},
						cli.IntFlag{
							Name:   "grpc-log-max-size",
							Value:  500,
							EnvVar: "BLAST_MANAGER_GRPC_LOG_MAX_SIZE",
							Usage:  "Max size of a log file (megabytes)",
						},
						cli.IntFlag{
							Name:   "grpc-log-max-backups",
							Value:  3,
							EnvVar: "BLAST_MANAGER_GRPC_LOG_MAX_BACKUPS",
							Usage:  "Max backup count of log files",
						},
						cli.IntFlag{
							Name:   "grpc-log-max-age",
							Value:  30,
							EnvVar: "BLAST_MANAGER_GRPC_LOG_MAX_AGE",
							Usage:  "Max age of a log file (days)",
						},
						cli.BoolFlag{
							Name:   "grpc-log-compress",
							EnvVar: "BLAST_MANAGER_GRPC_LOG_COMPRESS",
							Usage:  "Compress a log file",
						},
						cli.StringFlag{
							Name:   "http-log-file",
							Value:  os.Stderr.Name(),
							EnvVar: "BLAST_MANAGER_HTTP_LOG_FILE",
							Usage:  "HTTP access log file",
						},
						cli.IntFlag{
							Name:   "http-log-max-size",
							Value:  500,
							EnvVar: "BLAST_MANAGER_HTTP_LOG_MAX_SIZE",
							Usage:  "Max size of a HTTP access log file (megabytes)",
						},
						cli.IntFlag{
							Name:   "http-log-max-backups",
							Value:  3,
							EnvVar: "BLAST_MANAGER_HTTP_LOG_MAX_BACKUPS",
							Usage:  "Max backup count of HTTP access log files",
						},
						cli.IntFlag{
							Name:   "http-log-max-age",
							Value:  30,
							EnvVar: "BLAST_MANAGER_HTTP_LOG_MAX_AGE",
							Usage:  "Max age of a HTTP access log file (days)",
						},
						cli.BoolFlag{
							Name:   "http-log-compress",
							EnvVar: "BLAST_MANAGER_HTTP_LOG_COMPRESS",
							Usage:  "Compress a HTTP access log",
						},
					},
					Action: managerStart,
				},
				{
					Name:  "node",
					Usage: "Command for blast manager node",
					Subcommands: []cli.Command{
						{
							Name:  "info",
							Usage: "Get node information",
							Flags: []cli.Flag{
								cli.StringFlag{
									Name:  "grpc-address",
									Value: ":5100",
									Usage: "The gRPC address of the node for which to retrieve the node information",
								},
							},
							Action: managerNodeInfo,
						},
						{
							Name:  "healthcheck",
							Usage: "Health check the node",
							Flags: []cli.Flag{
								cli.StringFlag{
									Name:  "grpc-address",
									Value: ":5100",
									Usage: "The gRPC listen address",
								},
								cli.BoolFlag{
									Name:  "healthiness",
									Usage: "healthiness probe",
								},
								cli.BoolFlag{
									Name:  "liveness",
									Usage: "Liveness probe",
								},
								cli.BoolFlag{
									Name:  "readiness",
									Usage: "Readiness probe",
								},
							},
							Action: managerNodeHealthCheck,
						},
					},
				},
				{
					Name:  "cluster",
					Usage: "Command for blast manager cluster",
					Subcommands: []cli.Command{
						{
							Name:  "info",
							Usage: "Get cluster information",
							Flags: []cli.Flag{
								cli.StringFlag{
									Name:  "grpc-address",
									Value: ":5100",
									Usage: "The gRPC address of the node for which to retrieve the node information",
								},
							},
							Action: managerClusterInfo,
						},
						{
							Name:  "watch",
							Usage: "Watch peers",
							Flags: []cli.Flag{
								cli.StringFlag{
									Name:  "grpc-address",
									Value: ":5100",
									Usage: "The gRPC address of the node for which to retrieve the node information",
								},
							},
							Action: managerClusterWatch,
						},
						{
							Name:  "leave",
							Usage: "Leave the manager from the cluster",
							Flags: []cli.Flag{
								cli.StringFlag{
									Name:  "peer-grpc-address",
									Value: "",
									Usage: "The gRPC address of the peer node that exists in the cluster to be joined",
								},
								cli.StringFlag{
									Name:  "node-id",
									Value: "",
									Usage: "The gRPC listen address",
								},
							},
							Action: managerClusterLeave,
						},
					},
				},
				{
					Name:  "get",
					Usage: "Get data",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "grpc-address",
							Value: ":5100",
							Usage: "The gRPC listen address",
						},
						cli.StringFlag{
							Name:  "format",
							Value: "",
							Usage: "Output format",
						},
					},
					ArgsUsage: "[key]",
					Action:    managerGet,
				},
				{
					Name:  "set",
					Usage: "Set data",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "grpc-address",
							Value: ":5100",
							Usage: "The gRPC listen address",
						},
						cli.StringFlag{
							Name:  "file",
							Value: "",
							Usage: "Value file",
						},
					},
					ArgsUsage: "[key] [value]",
					Action:    managerSet,
				},
				{
					Name:  "delete",
					Usage: "Delete data",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "grpc-address",
							Value: ":5100",
							Usage: "The gRPC listen address",
						},
					},
					ArgsUsage: "[key]",
					Action:    managerDelete,
				},
				{
					Name:  "watch",
					Usage: "Watch data",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "grpc-address",
							Value: ":5100",
							Usage: "The gRPC listen address",
						},
					},
					ArgsUsage: "[key]",
					Action:    managerWatch,
				},
				{
					Name:  "snapshot",
					Usage: "Snapshot the data",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "grpc-address",
							Value: ":5100",
							Usage: "The gRPC listen address",
						},
					},
					Action: managerSnapshot,
				},
			},
		},
		{
			Name:  "indexer",
			Usage: "Command for blast indexer",
			Subcommands: []cli.Command{
				{
					Name:  "start",
					Usage: "Start blast indexer",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:   "manager-grpc-address",
							Value:  "",
							EnvVar: "BLAST_INDEXER_MANAGER_GRPC_ADDRESS",
							Usage:  "The gRPC address of the existing cluster manager to be joined",
						},
						cli.StringFlag{
							Name:   "shard-id",
							Value:  "",
							EnvVar: "BLAST_INDEXER_SHARD_ID",
							Usage:  "Shard ID registered in the existing cluster to be joined",
						},
						cli.StringFlag{
							Name:   "peer-grpc-address",
							Value:  "",
							EnvVar: "BLAST_INDEXER_PEER_GRPC_ADDRESS",
							Usage:  "The gRPC address of the peer node that exists in the cluster to be joined",
						},
						cli.StringFlag{
							Name:   "grpc-address",
							Value:  ":5000",
							EnvVar: "BLAST_INDEXER_GRPC_ADDRESS",
							Usage:  "The gRPC listen address",
						},
						cli.StringFlag{
							Name:   "http-address",
							Value:  ":8000",
							EnvVar: "BLAST_INDEXER_HTTP_ADDRESS",
							Usage:  "HTTP listen address",
						},
						cli.StringFlag{
							Name:   "node-id",
							Value:  "",
							EnvVar: "BLAST_INDEXER_NODE_ID",
							Usage:  "Unique ID to identify the node",
						},
						cli.StringFlag{
							Name:   "node-address",
							Value:  ":2000",
							EnvVar: "BLAST_INDEXER_NODE_ADDRESS",
							Usage:  "The address that should be bound to for internal cluster communications",
						},
						cli.StringFlag{
							Name:   "data-dir",
							Value:  "/tmp/blast/indexer",
							EnvVar: "BLAST_INDEXER_DATA_DIR",
							Usage:  "A data directory for the node to store state",
						},
						cli.StringFlag{
							Name:   "raft-storage-type",
							Value:  "boltdb",
							EnvVar: "BLAST_INDEXER_RAFT_STORAGE_TYPE",
							Usage:  "Storage type of the database that stores the state",
						},
						cli.StringFlag{
							Name:   "index-mapping-file",
							Value:  "",
							EnvVar: "BLAST_INDEXER_INDEX_MAPPING_FILE",
							Usage:  "An index mapping file to use",
						},
						cli.StringFlag{
							Name:   "index-type",
							Value:  bleve.Config.DefaultIndexType,
							EnvVar: "BLAST_INDEXER_INDEX_TYPE",
							Usage:  "An index type to use",
						},
						cli.StringFlag{
							Name:   "index-storage-type",
							Value:  bleve.Config.DefaultKVStore,
							EnvVar: "BLAST_INDEXER_INDEX_STORAGE_TYPE",
							Usage:  "An index storage type to use",
						},
						cli.StringFlag{
							Name:   "log-level",
							Value:  "INFO",
							EnvVar: "BLAST_INDEXER_LOG_LEVEL",
							Usage:  "Log level",
						},
						cli.StringFlag{
							Name:   "log-file",
							Value:  os.Stderr.Name(),
							EnvVar: "BLAST_INDEXER_LOG_FILE",
							Usage:  "Log file",
						},
						cli.IntFlag{
							Name:   "log-max-size",
							Value:  500,
							EnvVar: "BLAST_INDEXER_LOG_MAX_SIZE",
							Usage:  "Max size of a log file (megabytes)",
						},
						cli.IntFlag{
							Name:   "log-max-backups",
							Value:  3,
							EnvVar: "BLAST_INDEXER_LOG_MAX_BACKUPS",
							Usage:  "Max backup count of log files",
						},
						cli.IntFlag{
							Name:   "log-max-age",
							Value:  30,
							EnvVar: "BLAST_INDEXER_LOG_MAX_AGE",
							Usage:  "Max age of a log file (days)",
						},
						cli.BoolFlag{
							Name:   "log-compress",
							EnvVar: "BLAST_INDEXER_LOG_COMPRESS",
							Usage:  "Compress a log file",
						},
						cli.StringFlag{
							Name:   "grpc-log-level",
							Value:  "WARN",
							EnvVar: "BLAST_INDEXER_GRPC_LOG_LEVEL",
							Usage:  "gRPC log level",
						},
						cli.StringFlag{
							Name:   "grpc-log-file",
							Value:  os.Stderr.Name(),
							EnvVar: "BLAST_INDEXER_GRPC_LOG_FILE",
							Usage:  "gRPC log file",
						},
						cli.IntFlag{
							Name:   "grpc-log-max-size",
							Value:  500,
							EnvVar: "BLAST_INDEXER_GRPC_LOG_MAX_SIZE",
							Usage:  "Max size of a log file (megabytes)",
						},
						cli.IntFlag{
							Name:   "grpc-log-max-backups",
							Value:  3,
							EnvVar: "BLAST_INDEXER_GRPC_LOG_MAX_BACKUPS",
							Usage:  "Max backup count of log files",
						},
						cli.IntFlag{
							Name:   "grpc-log-max-age",
							Value:  30,
							EnvVar: "BLAST_INDEXER_GRPC_LOG_MAX_AGE",
							Usage:  "Max age of a log file (days)",
						},
						cli.BoolFlag{
							Name:   "grpc-log-compress",
							EnvVar: "BLAST_INDEXER_GRPC_LOG_COMPRESS",
							Usage:  "Compress a log file",
						},
						cli.StringFlag{
							Name:   "http-log-file",
							Value:  os.Stderr.Name(),
							EnvVar: "BLAST_INDEXER_HTTP_LOG_FILE",
							Usage:  "HTTP access log file",
						},
						cli.IntFlag{
							Name:   "http-log-max-size",
							Value:  500,
							EnvVar: "BLAST_INDEXER_HTTP_LOG_MAX_SIZE",
							Usage:  "Max size of a HTTP access log file (megabytes)",
						},
						cli.IntFlag{
							Name:   "http-log-max-backups",
							Value:  3,
							EnvVar: "BLAST_INDEXER_HTTP_LOG_MAX_BACKUPS",
							Usage:  "Max backup count of HTTP access log files",
						},
						cli.IntFlag{
							Name:   "http-log-max-age",
							Value:  30,
							EnvVar: "BLAST_INDEXER_HTTP_LOG_MAX_AGE",
							Usage:  "Max age of a HTTP access log file (days)",
						},
						cli.BoolFlag{
							Name:   "http-log-compress",
							EnvVar: "BLAST_INDEXER_HTTP_LOG_COMPRESS",
							Usage:  "Compress a HTTP access log",
						},
					},
					Action: indexerStart,
				},
				{
					Name:  "node",
					Usage: "Command for blast indexer node",
					Subcommands: []cli.Command{
						{
							Name:  "info",
							Usage: "Get node information",
							Flags: []cli.Flag{
								cli.StringFlag{
									Name:  "grpc-address",
									Value: ":5000",
									Usage: "The gRPC address of the node for which to retrieve the node information",
								},
							},
							Action: indexerNodeInfo,
						},
						{
							Name:  "healthcheck",
							Usage: "Health check the node",
							Flags: []cli.Flag{
								cli.StringFlag{
									Name:  "grpc-address",
									Value: ":5000",
									Usage: "The gRPC listen address",
								},
								cli.BoolFlag{
									Name:  "healthiness",
									Usage: "healthiness probe",
								},
								cli.BoolFlag{
									Name:  "liveness",
									Usage: "Liveness probe",
								},
								cli.BoolFlag{
									Name:  "readiness",
									Usage: "Readiness probe",
								},
							},
							Action: indexerNodeHealth,
						},
					},
				},
				{
					Name:  "cluster",
					Usage: "Command for blast indexer cluster",
					Subcommands: []cli.Command{
						{
							Name:  "info",
							Usage: "Get cluster information",
							Flags: []cli.Flag{
								cli.StringFlag{
									Name:  "grpc-address",
									Value: ":5000",
									Usage: "The gRPC address of the node for which to retrieve the node information",
								},
							},
							Action: indexerClusterInfo,
						},
						{
							Name:  "watch",
							Usage: "Watch cluster",
							Flags: []cli.Flag{
								cli.StringFlag{
									Name:  "grpc-address",
									Value: ":5000",
									Usage: "The gRPC address of the node for which to retrieve the node information",
								},
							},
							Action: indexerClusterWatch,
						},
						{
							Name:  "leave",
							Usage: "Leave the indexer from the cluster",
							Flags: []cli.Flag{
								cli.StringFlag{
									Name:  "manager-grpc-address",
									Value: "",
									Usage: "The gRPC address of the existing cluster node to be joined",
								},
								cli.StringFlag{
									Name:  "shard-id",
									Value: "",
									Usage: "Shard ID registered in the existing cluster to be joined",
								},
								cli.StringFlag{
									Name:  "peer-grpc-address",
									Value: "",
									Usage: "The gRPC address of the peer node that exists in the cluster to be joined",
								},
								cli.StringFlag{
									Name:  "node-id",
									Value: "",
									Usage: "Node ID to delete",
								},
							},
							Action: indexerClusterLeave,
						},
					},
				},
				{
					Name:  "get",
					Usage: "Get document(s)",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "grpc-address",
							Value: ":5000",
							Usage: "The gRPC listen address",
						},
						cli.StringFlag{
							Name:  "file",
							Value: "",
							Usage: "Document ID list",
						},
					},
					ArgsUsage: "[document ID]",
					Action:    indexerGet,
				},
				{
					Name:  "index",
					Usage: "Index document(s)",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "grpc-address",
							Value: ":5000",
							Usage: "The gRPC listen address",
						},
						cli.StringFlag{
							Name:  "file",
							Value: "",
							Usage: "Document list",
						},
						cli.BoolFlag{
							Name:  "bulk",
							Usage: "Bulk indexing",
						},
					},
					ArgsUsage: "[document ID] [document fields]",
					Action:    indexerIndex,
				},
				{
					Name:  "delete",
					Usage: "Delete document(s)",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "grpc-address",
							Value: ":5000",
							Usage: "The gRPC listen address",
						},
						cli.StringFlag{
							Name:  "file",
							Value: "",
							Usage: "Document ID list",
						},
					},
					ArgsUsage: "[document ID]",
					Action:    indexerDelete,
				},
				{
					Name:  "search",
					Usage: "Search document(s)",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "grpc-address",
							Value: ":5000",
							Usage: "The gRPC listen address",
						},
						cli.StringFlag{
							Name:  "file",
							Value: "",
							Usage: "Search request",
						},
					},
					ArgsUsage: "[search request]",
					Action:    indexerSearch,
				},
				{
					Name:  "snapshot",
					Usage: "Snapshot",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "grpc-address",
							Value: ":5000",
							Usage: "The gRPC listen address",
						},
					},
					Action: indexerSnapshot,
				},
			},
		},
		{
			Name:  "dispatcher",
			Usage: "Command for blast dispatcher",
			Subcommands: []cli.Command{
				{
					Name:  "start",
					Usage: "Start blast dispatcher",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:   "manager-grpc-address",
							Value:  ":5100",
							EnvVar: "BLAST_DISPATCHER_CLUSTER_GRPC_ADDRESS",
							Usage:  "The gRPC address of the existing cluster node to be joined",
						},
						cli.StringFlag{
							Name:   "grpc-address",
							Value:  ":5200",
							EnvVar: "BLAST_DISPATCHER_GRPC_ADDRESS",
							Usage:  "The gRPC listen address",
						},
						cli.StringFlag{
							Name:   "http-address",
							Value:  ":8200",
							EnvVar: "BLAST_DISPATCHER_HTTP_ADDRESS",
							Usage:  "HTTP listen address",
						},
						cli.StringFlag{
							Name:   "log-level",
							Value:  "INFO",
							EnvVar: "BLAST_DISPATCHER_LOG_LEVEL",
							Usage:  "Log level",
						},
						cli.StringFlag{
							Name:   "log-file",
							Value:  os.Stderr.Name(),
							EnvVar: "BLAST_DISPATCHER_LOG_FILE",
							Usage:  "Log file",
						},
						cli.IntFlag{
							Name:   "log-max-size",
							Value:  500,
							EnvVar: "BLAST_DISPATCHER_LOG_MAX_SIZE",
							Usage:  "Max size of a log file (megabytes)",
						},
						cli.IntFlag{
							Name:   "log-max-backups",
							Value:  3,
							EnvVar: "BLAST_DISPATCHER_LOG_MAX_BACKUPS",
							Usage:  "Max backup count of log files",
						},
						cli.IntFlag{
							Name:   "log-max-age",
							Value:  30,
							EnvVar: "BLAST_DISPATCHER_LOG_MAX_AGE",
							Usage:  "Max age of a log file (days)",
						},
						cli.BoolFlag{
							Name:   "log-compress",
							EnvVar: "BLAST_DISPATCHER_LOG_COMPRESS",
							Usage:  "Compress a log file",
						},
						cli.StringFlag{
							Name:   "grpc-log-level",
							Value:  "WARN",
							EnvVar: "BLAST_DISPATCHER_GRPC_LOG_LEVEL",
							Usage:  "gRPC log level",
						},
						cli.StringFlag{
							Name:   "grpc-log-file",
							Value:  os.Stderr.Name(),
							EnvVar: "BLAST_DISPATCHER_GRPC_LOG_FILE",
							Usage:  "gRPC log file",
						},
						cli.IntFlag{
							Name:   "grpc-log-max-size",
							Value:  500,
							EnvVar: "BLAST_DISPATCHER_GRPC_LOG_MAX_SIZE",
							Usage:  "Max size of a log file (megabytes)",
						},
						cli.IntFlag{
							Name:   "grpc-log-max-backups",
							Value:  3,
							EnvVar: "BLAST_DISPATCHER_GRPC_LOG_MAX_BACKUPS",
							Usage:  "Max backup count of log files",
						},
						cli.IntFlag{
							Name:   "grpc-log-max-age",
							Value:  30,
							EnvVar: "BLAST_DISPATCHER_GRPC_LOG_MAX_AGE",
							Usage:  "Max age of a log file (days)",
						},
						cli.BoolFlag{
							Name:   "grpc-log-compress",
							EnvVar: "BLAST_DISPATCHER_GRPC_LOG_COMPRESS",
							Usage:  "Compress a log file",
						},
						cli.StringFlag{
							Name:   "http-log-file",
							Value:  os.Stderr.Name(),
							EnvVar: "BLAST_DISPATCHER_HTTP_LOG_FILE",
							Usage:  "HTTP access log file",
						},
						cli.IntFlag{
							Name:   "http-log-max-size",
							Value:  500,
							EnvVar: "BLAST_DISPATCHER_HTTP_LOG_MAX_SIZE",
							Usage:  "Max size of a HTTP access log file (megabytes)",
						},
						cli.IntFlag{
							Name:   "http-log-max-backups",
							Value:  3,
							EnvVar: "BLAST_DISPATCHER_HTTP_LOG_MAX_BACKUPS",
							Usage:  "Max backup count of HTTP access log files",
						},
						cli.IntFlag{
							Name:   "http-log-max-age",
							Value:  30,
							EnvVar: "BLAST_DISPATCHER_HTTP_LOG_MAX_AGE",
							Usage:  "Max age of a HTTP access log file (days)",
						},
						cli.BoolFlag{
							Name:   "http-log-compress",
							EnvVar: "BLAST_DISPATCHER_HTTP_LOG_COMPRESS",
							Usage:  "Compress a HTTP access log",
						},
					},
					Action: dispatcherStart,
				},
				{
					Name:  "node",
					Usage: "Command for blast dispatcher node",
					Subcommands: []cli.Command{
						{
							Name:  "healthcheck",
							Usage: "Health check the node",
							Flags: []cli.Flag{
								cli.StringFlag{
									Name:  "grpc-address",
									Value: ":5200",
									Usage: "The gRPC listen address",
								},
								cli.BoolFlag{
									Name:  "healthiness",
									Usage: "healthiness probe",
								},
								cli.BoolFlag{
									Name:  "liveness",
									Usage: "Liveness probe",
								},
								cli.BoolFlag{
									Name:  "readiness",
									Usage: "Readiness probe",
								},
							},
							Action: dispatcherNodeHealth,
						},
					},
				},
				{
					Name:  "get",
					Usage: "Get document(s)",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "grpc-address",
							Value: ":5200",
							Usage: "The gRPC listen address",
						},
						cli.StringFlag{
							Name:  "file",
							Value: "",
							Usage: "Document ID list",
						},
					},
					ArgsUsage: "[document IDs]",
					Action:    dispatcherGet,
				},
				{
					Name:  "index",
					Usage: "Index document(s)",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "grpc-address",
							Value: ":5200",
							Usage: "The gRPC listen address",
						},
						cli.StringFlag{
							Name:  "file",
							Value: "",
							Usage: "Document list",
						},
						cli.BoolFlag{
							Name:  "bulk",
							Usage: "Bulk indexing",
						},
					},
					ArgsUsage: "[document ID] [document fields]",
					Action:    dispatcherIndex,
				},
				{
					Name:  "delete",
					Usage: "Delete document(s)",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "grpc-address",
							Value: ":5200",
							Usage: "The gRPC listen address",
						},
						cli.StringFlag{
							Name:  "file",
							Value: "",
							Usage: "Document ID list",
						},
					},
					ArgsUsage: "[document IDs]",
					Action:    dispatcherDelete,
				},
				{
					Name:  "search",
					Usage: "Search document(s)",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "grpc-address",
							Value: ":5200",
							Usage: "The gRPC listen address",
						},
						cli.StringFlag{
							Name:  "file",
							Value: "",
							Usage: "Search request",
						},
					},
					ArgsUsage: "[search request]",
					Action:    dispatcherSearch,
				},
			},
		},
	}

	cli.HelpFlag = cli.BoolFlag{
		Name:  "help, h",
		Usage: "Show this message",
	}
	cli.VersionFlag = cli.BoolFlag{
		Name:  "version, v",
		Usage: "Print the version",
	}

	err := app.Run(os.Args)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
	}
}
