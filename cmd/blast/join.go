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
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"github.com/mosuka/blast/node/data/client"
	"github.com/mosuka/blast/node/data/protobuf"
	"github.com/urfave/cli"
)

func join(c *cli.Context) {
	bindAddr := c.String("bind-addr")
	grpcAddr := c.String("grpc-addr")
	httpAddr := c.String("http-addr")
	nodeID := c.String("raft-node-id")
	peerGRPCAddr := c.String("peer-grpc-addr")
	prettyPrint := c.Bool("pretty-print")

	var err error

	var dataClient *client.GRPCClient
	if dataClient, err = client.NewGRPCClient(peerGRPCAddr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	defer dataClient.Close()

	req := &protobuf.JoinRequest{
		NodeId:  nodeID,
		Address: bindAddr,
		Metadata: &protobuf.Metadata{
			GrpcAddress: grpcAddr,
			HttpAddress: httpAddr,
		},
	}

	var resp *protobuf.JoinResponse
	if resp, err = dataClient.Join(req); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	var jsonBytes []byte
	if jsonBytes, err = resp.GetBytes(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	if prettyPrint {
		var buff bytes.Buffer
		if err = json.Indent(&buff, jsonBytes, "", "  "); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}
		jsonBytes = buff.Bytes()
	}

	fmt.Fprintln(os.Stdout, fmt.Sprintf("%s", string(jsonBytes)))
	return
}
