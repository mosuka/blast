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
					Name:  "manager-addr",
					Value: "",
					Usage: "Manager address",
				},
				cli.StringFlag{
					Name:  "grpc-addr",
					Value: ":27070",
					Usage: "gRPC Server listen address",
				},
				cli.StringFlag{
					Name:  "http-addr",
					Value: ":28080",
					Usage: "HTTP server listen address",
				},
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
			},
			Action: execStart,
		},
		{
			Name:  "livenessprobe",
			Usage: "Liveness probe",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "grpc-addr",
					Value: ":27070",
					Usage: "address to connect to",
				},
			},
			Action: execLivenessProbe,
		},
		{
			Name:  "readinessprobe",
			Usage: "Readiness probe",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "grpc-addr",
					Value: ":27070",
					Usage: "address to connect to",
				},
			},
			Action: execReadinessProbe,
		},
		{
			Name:  "get",
			Usage: "get document",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "grpc-addr",
					Value: ":27070",
					Usage: "gRPC address to connect to",
				},
				cli.StringFlag{
					Name:  "id",
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
					Name:  "grpc-addr",
					Value: ":27070",
					Usage: "gRPC address to connect to",
				},
				cli.StringFlag{
					Name:  "id",
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
					Name:  "grpc-addr",
					Value: ":27070",
					Usage: "address to connect to",
				},
				cli.StringFlag{
					Name:  "id",
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
					Name:  "grpc-addr",
					Value: ":27070",
					Usage: "gRPC address to connect to",
				},
			},
			ArgsUsage: "[search request]",
			Action:    execSearch,
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
