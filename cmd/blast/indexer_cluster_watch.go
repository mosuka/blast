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
	"io"
	"log"
	"os"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/mosuka/blast/indexer"
	"github.com/mosuka/blast/protobuf/index"
	"github.com/urfave/cli"
)

func indexerClusterWatch(c *cli.Context) error {
	grpcAddr := c.String("grpc-address")

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

	marshaler := indexer.JsonMarshaler{}

	req := &empty.Empty{}
	clusterInfo, err := client.ClusterInfo(req)
	if err != nil {
		return err
	}
	resp := &index.ClusterWatchResponse{
		Event:   0,
		Node:    nil,
		Cluster: clusterInfo.Cluster,
	}
	respBytes, err := marshaler.Marshal(resp)
	if err != nil {
		return err
	}
	_, _ = fmt.Fprintln(os.Stdout, fmt.Sprintf("%v", string(respBytes)))

	clusterWatchClient, err := client.ClusterWatch(req)
	if err != nil {
		return err
	}

	for {
		resp, err := clusterWatchClient.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Println(err.Error())
			break
		}
		respBytes, err = marshaler.Marshal(resp)
		if err != nil {
			return err
		}
		_, _ = fmt.Fprintln(os.Stdout, fmt.Sprintf("%v", string(respBytes)))
	}

	return nil
}
