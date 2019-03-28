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

	"github.com/mosuka/blast/version"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = path.Base(os.Args[0])
	app.Usage = "Blast indexer"
	app.Version = version.Version
	app.Authors = []cli.Author{
		{
			Name:  "mosuka",
			Email: "minoru.osuka@gmail.com",
		},
	}
	app.Commands = []cli.Command{
		{
			Name:  "start",
			Usage: "Start index server",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "node-id, n",
					Value: "",
					Usage: "Node ID",
				},
				cli.StringFlag{
					Name:  "bind-addr, b",
					Value: ":6060",
					Usage: "Raft bind address",
				},
				cli.StringFlag{
					Name:  "grpc-addr, g",
					Value: ":5050",
					Usage: "gRPC Server listen address",
				},
				cli.StringFlag{
					Name:  "http-addr, H",
					Value: ":8080",
					Usage: "HTTP server listen address",
				},
				cli.StringFlag{
					Name:  "data-dir, d",
					Value: "/tmp/blast-index",
					Usage: "Data directory",
				},
				cli.StringFlag{
					Name:  "join-addr, j",
					Value: "",
					Usage: "Existing gRPC server listen address to join to the cluster",
				},
				cli.StringFlag{
					Name:  "index-mapping-file, m",
					Value: "",
					Usage: "Path to a file containing a JSON representation of an index mapping to use",
				},
				cli.StringFlag{
					Name:  "index-storage-type, s",
					Value: "boltdb",
					Usage: "Index storage type to use",
				},
				cli.StringFlag{
					Name:  "log-level, L",
					Value: "INFO",
					Usage: "Log level",
				},
				cli.StringFlag{
					Name:  "log-file, F",
					Value: os.Stderr.Name(),
					Usage: "Log file",
				},
				cli.IntFlag{
					Name:  "log-max-size, S",
					Value: 500,
					Usage: "Max size of a log file (megabytes)",
				},
				cli.IntFlag{
					Name:  "log-max-backups, B",
					Value: 3,
					Usage: "Max backup count of log files",
				},
				cli.IntFlag{
					Name:  "log-max-age, A",
					Value: 30,
					Usage: "Max age of a log file (days)",
				},
				cli.BoolFlag{
					Name:  "log-compress, C",
					Usage: "Compress a log file",
				},
				cli.StringFlag{
					Name:  "http-access-log-file",
					Value: os.Stderr.Name(),
					Usage: "HTTP access log file",
				},
				cli.IntFlag{
					Name:  "http-access-log-max-size",
					Value: 500,
					Usage: "Max size of a HTTP access log file (megabytes)",
				},
				cli.IntFlag{
					Name:  "http-access-log-max-backups",
					Value: 3,
					Usage: "Max backup count of HTTP access log files",
				},
				cli.IntFlag{
					Name:  "http-access-log-max-age",
					Value: 30,
					Usage: "Max age of a HTTP access log file (days)",
				},
				cli.BoolFlag{
					Name:  "http-access-log-compress",
					Usage: "Compress a HTTP access log",
				},
			},
			Action: execStart,
		},
		{
			Name:  "join",
			Usage: "Join a node to the cluster",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "grpc-addr, g",
					Value: ":5050",
					Usage: "gRPC address to connect to",
				},
			},
			ArgsUsage: "[id] [addr]",
			Action:    execJoin,
		},
		{
			Name:  "leave",
			Usage: "Leave a node from the cluster",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "grpc-addr, g",
					Value: ":5050",
					Usage: "gRPC address to connect to",
				},
			},
			ArgsUsage: "[id]",
			Action:    execLeave,
		},
		{
			Name:  "node",
			Usage: "Get node",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "grpc-addr, g",
					Value: ":5050",
					Usage: "gRPC address to connect to",
				},
			},
			ArgsUsage: "[id]",
			Action:    execNode,
		},
		{
			Name:  "cluster",
			Usage: "Get cluster",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "grpc-addr, g",
					Value: ":5050",
					Usage: "gRPC address to connect to",
				},
			},
			Action: execCluster,
		},
		{
			Name:  "snapshot",
			Usage: "Create snapshot manually",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "grpc-addr, g",
					Value: ":5050",
					Usage: "gRPC address to connect to",
				},
			},
			Action: execSnapshot,
		},
		{
			Name:  "get",
			Usage: "get document",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "grpc-addr, g",
					Value: ":5050",
					Usage: "gRPC address to connect to",
				},
				cli.StringFlag{
					Name:  "id, i",
					Value: "",
					Usage: "document id",
				},
			},
			Action: execGet,
		},
		{
			Name:  "index",
			Usage: "Index documents in bulk",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "grpc-addr, g",
					Value: ":5050",
					Usage: "gRPC address to connect to",
				},
				cli.StringFlag{
					Name:  "id, i",
					Value: "",
					Usage: "document id",
				},
			},
			ArgsUsage: "[documents | fields]",
			Action:    execIndex,
		},
		{
			Name:  "delete",
			Usage: "Delete documents in bulk",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "grpc-addr, g",
					Value: ":5050",
					Usage: "address to connect to",
				},
				cli.StringFlag{
					Name:  "id, i",
					Value: "",
					Usage: "document id",
				},
			},
			ArgsUsage: "[documents]",
			Action:    execDelete,
		},
		{
			Name:  "search",
			Usage: "Search documents",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "grpc-addr, g",
					Value: ":5050",
					Usage: "gRPC address to connect to",
				},
			},
			ArgsUsage: "[search request]",
			Action:    execSearch,
		},
		{
			Name:  "stats",
			Usage: "Get a index stats",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "grpc-addr, g",
					Value: ":5050",
					Usage: "address to connect to",
				},
			},
			Action: execStats,
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
		fmt.Fprintln(os.Stderr, err)
	}
}
