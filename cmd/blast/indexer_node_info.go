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
	"encoding/json"
	"fmt"
	"os"

	"github.com/mosuka/blast/indexer"
	"github.com/urfave/cli"
)

func indexerNodeInfo(c *cli.Context) error {
	clusterGrpcAddr := c.String("cluster-grpc-address")
	shardId := c.String("shard-id")
	peerGrpcAddr := c.String("peer-grpc-address")

	if clusterGrpcAddr != "" && shardId != "" {

	} else if peerGrpcAddr != "" {

	}

	grpcAddr := c.String("grpc-address")

	nodeId := c.Args().Get(0)

	client, err := indexer.NewGRPCClient(grpcAddr)
	if err != nil {
		return err
	}
	defer func() {
		err := client.Close()
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
		}
	}()

	metadata, err := client.GetNode(nodeId)
	if err != nil {
		return err
	}

	metadataBytes, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return err
	}

	_, _ = fmt.Fprintln(os.Stdout, fmt.Sprintf("%v", string(metadataBytes)))

	return nil
}
