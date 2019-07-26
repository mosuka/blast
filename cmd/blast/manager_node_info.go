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

	"github.com/mosuka/blast/manager"
	"github.com/urfave/cli"
)

func managerNodeInfo(c *cli.Context) error {
	peerGrpcAddr := c.String("peer-grpc-address")
	nodeId := c.String("node-id")
	grpcAddr := c.String("grpc-address")

	if peerGrpcAddr != "" {

	}

	client, err := manager.NewGRPCClient(grpcAddr)
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
