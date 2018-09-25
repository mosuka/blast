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

import "github.com/urfave/cli"

var (
	commands = []cli.Command{
		{
			Name:  "get",
			Usage: "Get a document",
			Flags: []cli.Flag{
				flGRPCAddr,
			},
			ArgsUsage: "[id]",
			Action:    get,
		},
		{
			Name:  "put",
			Usage: "Put a document",
			Flags: []cli.Flag{
				flGRPCAddr,
			},
			ArgsUsage: "[id] [fields]",
			Action:    put,
		},
		{
			Name:  "delete",
			Usage: "Delete a document",
			Flags: []cli.Flag{
				flGRPCAddr,
			},
			ArgsUsage: "[id]",
			Action:    delete,
		},
		{
			Name:  "bulk",
			Usage: "Update documents in bulk",
			Flags: []cli.Flag{
				flGRPCAddr,
				flBatchSize,
			},
			ArgsUsage: "[update requests]",
			Action:    bulk,
		},
		{
			Name:  "search",
			Usage: "Search documents",
			Flags: []cli.Flag{
				flGRPCAddr,
			},
			ArgsUsage: "[search request]",
			Action:    search,
		},
		{
			Name:  "join",
			Usage: "Join a node to the cluster",
			Flags: []cli.Flag{
				flGRPCAddr,
				flTargetRaftNodeID,
				flTargetRaftAddr,
				flTargetGRPCAddr,
				flTargetHTTPAddr,
			},
			Action: join,
		},
		{
			Name:  "leave",
			Usage: "Leave a node from the cluster",
			Flags: []cli.Flag{
				flGRPCAddr,
				flTargetRaftNodeID,
				flTargetRaftAddr,
			},
			Action: leave,
		},
		{
			Name:  "cluster",
			Usage: "Shows a list of peers in a cluster",
			Flags: []cli.Flag{
				flGRPCAddr,
			},
			Action: cluster,
		},
		{
			Name:  "snapshot",
			Usage: "Create snapshot",
			Flags: []cli.Flag{
				flGRPCAddr,
			},
			Action: snapshot,
		},
	}
)
