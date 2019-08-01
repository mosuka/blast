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
	"io"
	"log"
	"os"

	"github.com/mosuka/blast/manager"
	"github.com/mosuka/blast/protobuf/management"
	"github.com/urfave/cli"
)

func managerClusterWatch(c *cli.Context) error {
	grpcAddr := c.String("grpc-address")

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

	cluster, err := client.ClusterInfo()
	if err != nil {
		return err
	}
	resp := &management.ClusterWatchResponse{
		Event:   0,
		Node:    nil,
		Cluster: cluster,
	}
	clusterBytes, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		return err
	}
	_, _ = fmt.Fprintln(os.Stdout, fmt.Sprintf("%v", string(clusterBytes)))

	watchClient, err := client.ClusterWatch()
	if err != nil {
		return err
	}

	for {
		resp, err := watchClient.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Println(err.Error())
			break
		}

		clusterBytes, err := json.MarshalIndent(resp, "", "  ")
		if err != nil {
			return err
		}
		_, _ = fmt.Fprintln(os.Stdout, fmt.Sprintf("%v", string(clusterBytes)))
	}

	return nil
}
