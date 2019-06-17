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
	app.Usage = "blast"
	app.Version = version.Version
	app.Authors = []cli.Author{
		{
			Name:  "mosuka",
			Email: "minoru.osuka@gmail.com",
		},
	}
	app.Commands = []cli.Command{
		{
			Name:  "livenessprobe",
			Usage: "liveness probe",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "grpc-addr",
					Value: ":5001",
					Usage: "address to connect to",
				},
			},
			Action: execLivenessProbe,
		},
		{
			Name:  "readinessprobe",
			Usage: "readiness probe",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "grpc-addr",
					Value: ":5001",
					Usage: "address to connect to",
				},
			},
			Action: execReadinessProbe,
		},
		{
			Name:  "get",
			Usage: "get",
			Subcommands: []cli.Command{
				{
					Name:  "metadata",
					Usage: "get node metadata",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "grpc-addr",
							Value: ":5001",
							Usage: "address to connect to",
						},
					},
					ArgsUsage: "[id]",
					Action:    execGetMetadata,
				},
				{
					Name:  "nodestate",
					Usage: "get node state",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "grpc-addr",
							Value: ":5001",
							Usage: "address to connect to",
						},
					},
					ArgsUsage: "[id]",
					Action:    execGetNodeState,
				},
				{
					Name:  "node",
					Usage: "get node",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "grpc-addr",
							Value: ":5001",
							Usage: "address to connect to",
						},
					},
					ArgsUsage: "[id]",
					Action:    execGetNode,
				},
				{
					Name:  "cluster",
					Usage: "get cluster",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "grpc-addr",
							Value: ":5001",
							Usage: "address to connect to",
						},
					},
					Action: execGetCluster,
				},
				{
					Name:  "state",
					Usage: "get state",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "grpc-addr",
							Value: ":5001",
							Usage: "address to connect to",
						},
					},
					ArgsUsage: "[key]",
					Action:    execGetState,
				},
				{
					Name:  "document",
					Usage: "get document",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "grpc-addr",
							Value: ":5001",
							Usage: "address to connect to",
						},
					},
					ArgsUsage: "[id]",
					Action:    execGetDocument,
				},
			},
		},
		{
			Name:  "set",
			Usage: "set",
			Subcommands: []cli.Command{
				{
					Name:  "node",
					Usage: "set node",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "grpc-addr",
							Value: ":5001",
							Usage: "address to connect to",
						},
					},
					ArgsUsage: "[id] [metadata]",
					Action:    execSetNode,
				},
				{
					Name:  "state",
					Usage: "set state",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "grpc-addr",
							Value: ":5001",
							Usage: "address to connect to",
						},
					},
					ArgsUsage: "[key] [value]",
					Action:    execSetState,
				},
				{
					Name:  "document",
					Usage: "set document",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "grpc-addr",
							Value: ":5001",
							Usage: "address to connect to",
						},
					},
					ArgsUsage: "[documents | [id] [fields]]",
					Action:    execSetDocument,
				},
			},
		},
		{
			Name:  "delete",
			Usage: "delete",
			Subcommands: []cli.Command{
				{
					Name:  "node",
					Usage: "delete node",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "grpc-addr",
							Value: ":5001",
							Usage: "address to connect to",
						},
					},
					ArgsUsage: "[id]",
					Action:    execDeleteNode,
				},
				{
					Name:  "state",
					Usage: "delete state",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "grpc-addr",
							Value: ":5001",
							Usage: "address to connect to",
						},
					},
					ArgsUsage: "[key] [value]",
					Action:    execDeleteState,
				},
				{
					Name:  "document",
					Usage: "delete document",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "grpc-addr",
							Value: ":5001",
							Usage: "address to connect to",
						},
					},
					ArgsUsage: "[id] ...",
					Action:    execDeleteDocument,
				},
			},
		},
		{
			Name:  "watch",
			Usage: "watch",
			Subcommands: []cli.Command{
				{
					Name:  "cluster",
					Usage: "watch cluster",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "grpc-addr",
							Value: ":5001",
							Usage: "address to connect to",
						},
					},
					Action: execWatchCluster,
				},
				{
					Name:  "state",
					Usage: "watch state",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "grpc-addr",
							Value: ":5001",
							Usage: "address to connect to",
						},
					},
					ArgsUsage: "[key]",
					Action:    execWatchState,
				},
			},
		},
		{
			Name:  "snapshot",
			Usage: "snapshot data",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "grpc-addr",
					Value: ":5001",
					Usage: "address to connect to",
				},
			},
			Action: execSnapshot,
		},
		{
			Name:  "search",
			Usage: "search documents",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "grpc-addr",
					Value: ":5001",
					Usage: "address to connect to",
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
		_, _ = fmt.Fprintln(os.Stderr, err)
	}
}
