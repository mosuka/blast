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

	"github.com/mosuka/blast/indexer"
	"github.com/urfave/cli"
)

func indexerClusterLeave(c *cli.Context) error {
	clusterGrpcAddr := c.String("manager-grpc-address")
	shardId := c.String("shard-id")
	peerGrpcAddr := c.String("peer-grpc-address")

	if clusterGrpcAddr != "" && shardId != "" {
		// get grpc address of leader node
	} else if peerGrpcAddr != "" {
		// get grpc address of leader node
	}

	nodeId := c.String("node-id")

	client, err := indexer.NewGRPCClient(peerGrpcAddr)
	if err != nil {
		return err
	}
	defer func() {
		err := client.Close()
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
		}
	}()

	err = client.ClusterLeave(nodeId)
	if err != nil {
		return err
	}

	return nil
}
