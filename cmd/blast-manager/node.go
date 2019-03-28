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

func execNode(c *cli.Context) error {
	grpcAddr := c.String("grpc-addr")

	client, err := manager.NewGRPCClient(grpcAddr)
	if err != nil {
		return err
	}
	defer func() {
		err := client.Close()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}()

	nodes, err := client.GetNode()
	if err != nil {
		return err
	}

	nodesBytes, err := json.MarshalIndent(nodes, "", "  ")
	if err != nil {
		return err
	}

	fmt.Fprintln(os.Stdout, fmt.Sprintf("%v\n", string(nodesBytes)))

	return nil
}
