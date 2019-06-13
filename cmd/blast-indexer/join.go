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

	"github.com/mosuka/blast/grpc"
	"github.com/urfave/cli"
)

func execJoin(c *cli.Context) error {
	grpcAddr := c.String("grpc-addr")
	peerAddr := c.String("peer-addr")

	// TODO: set target metadeta from CLI

	targetClient, err := grpc.NewClient(grpcAddr)
	if err != nil {
		return err
	}
	defer func() {
		err := targetClient.Close()
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
		}
	}()

	metadata, err := targetClient.GetNode("")
	if err != nil {
		return err
	}

	peerClient, err := grpc.NewClient(peerAddr)
	if err != nil {
		return err
	}
	defer func() {
		err := targetClient.Close()
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
		}
	}()

	err = peerClient.SetNode("", metadata)
	if err != nil {
		return err
	}

	return nil
}
