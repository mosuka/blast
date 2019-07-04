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
	app.Usage = "blastd"
	app.Version = version.Version
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "log-level",
			Value: "INFO",
			Usage: "Log level",
		},
		cli.StringFlag{
			Name:  "log-file",
			Value: os.Stderr.Name(),
			Usage: "Log file",
		},
		cli.IntFlag{
			Name:  "log-max-size",
			Value: 500,
			Usage: "Max size of a log file (megabytes)",
		},
		cli.IntFlag{
			Name:  "log-max-backups",
			Value: 3,
			Usage: "Max backup count of log files",
		},
		cli.IntFlag{
			Name:  "log-max-age",
			Value: 30,
			Usage: "Max age of a log file (days)",
		},
		cli.BoolFlag{
			Name:  "log-compress",
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
	}
	app.Authors = []cli.Author{
		{
			Name:  "mosuka",
			Email: "minoru.osuka@gmail.com",
		},
	}
	app.Commands = []cli.Command{
		{
			Name:  "indexer",
			Usage: "Start indexer",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "manager-addr",
					Value: "",
					Usage: "Manager address",
				},
				cli.StringFlag{
					Name:  "cluster-id",
					Value: "default",
					Usage: "Cluster ID",
				},
				cli.StringFlag{
					Name:  "node-id",
					Value: "indexer1",
					Usage: "Node ID",
				},
				cli.StringFlag{
					Name:  "bind-addr",
					Value: ":5000",
					Usage: "Raft bind address",
				},
				cli.StringFlag{
					Name:  "grpc-addr",
					Value: ":5001",
					Usage: "gRPC Server listen address",
				},
				cli.StringFlag{
					Name:  "http-addr",
					Value: ":5002",
					Usage: "HTTP server listen address",
				},
				cli.StringFlag{
					Name:  "data-dir",
					Value: "/tmp/blast-index",
					Usage: "Data directory",
				},
				cli.StringFlag{
					Name:  "raft-storage-type",
					Value: "boltdb",
					Usage: "Raft log storage type to use",
				},
				cli.StringFlag{
					Name:  "peer-addr",
					Value: "",
					Usage: "Existing gRPC server listen address to join to the cluster",
				},
				cli.StringFlag{
					Name:  "index-mapping-file",
					Value: "",
					Usage: "Path to a file containing a JSON representation of an index mapping to use",
				},
				cli.StringFlag{
					Name:  "index-type",
					Value: bleve.Config.DefaultIndexType,
					Usage: "Index storage type to use",
				},
				cli.StringFlag{
					Name:  "index-storage-type",
					Value: bleve.Config.DefaultKVStore,
					Usage: "Index storage type to use",
				},
			},
			Action: startIndexer,
		},
		{
			Name:  "manager",
			Usage: "Start manager",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "node-id",
					Value: "",
					Usage: "Node ID",
				},
				cli.StringFlag{
					Name:  "bind-addr",
					Value: ":15000",
					Usage: "Raft bind address",
				},
				cli.StringFlag{
					Name:  "grpc-addr",
					Value: ":15001",
					Usage: "gRPC Server listen address",
				},
				cli.StringFlag{
					Name:  "http-addr",
					Value: ":15002",
					Usage: "HTTP server listen address",
				},
				cli.StringFlag{
					Name:  "data-dir",
					Value: "./",
					Usage: "Data directory",
				},
				cli.StringFlag{
					Name:  "raft-storage-type",
					Value: "boltdb",
					Usage: "Raft log storage type to use",
				},
				cli.StringFlag{
					Name:  "peer-addr",
					Value: "",
					Usage: "Existing gRPC server listen address to join to the cluster",
				},
				cli.StringFlag{
					Name:  "index-mapping-file",
					Value: "",
					Usage: "Path to a file containing a JSON representation of an index mapping to use",
				},
				cli.StringFlag{
					Name:  "index-type",
					Value: bleve.Config.DefaultIndexType,
					Usage: "Index storage type to use",
				},
				cli.StringFlag{
					Name:  "index-storage-type",
					Value: bleve.Config.DefaultKVStore,
					Usage: "Index storage type to use",
				},
			},
			Action: startManager,
		},
		{
			Name:  "dispatcher",
			Usage: "Start dispatcher",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "manager-addr",
					Value: ":15001",
					Usage: "Manager address",
				},
				cli.StringFlag{
					Name:  "grpc-addr",
					Value: ":25001",
					Usage: "gRPC Server listen address",
				},
				cli.StringFlag{
					Name:  "http-addr",
					Value: ":25002",
					Usage: "HTTP server listen address",
				},
			},
			Action: startDispatcher,
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
