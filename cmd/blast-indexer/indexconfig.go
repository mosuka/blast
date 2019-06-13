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
	"errors"
	"fmt"
	"os"

	"github.com/mosuka/blast/grpc"
	"github.com/mosuka/blast/protobuf"
	"github.com/urfave/cli"
)

func execIndexConfig(c *cli.Context) error {
	grpcAddr := c.String("grpc-addr")

	client, err := grpc.NewClient(grpcAddr)
	if err != nil {
		return err
	}
	defer func() {
		err := client.Close()
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
		}
	}()

	resp, err := client.GetIndexConfig()
	if err != nil {
		return err
	}

	// Any -> IndexMappingImpl
	ins, err := protobuf.MarshalAny(resp.IndexConfig)
	if err != nil {
		return err
	}
	if ins == nil {
		return errors.New("nil")
	}
	indexConfig := *ins.(*map[string]interface{})

	// map[string]interface -> []byte
	indexConfigBytes, err := json.MarshalIndent(indexConfig, "", "  ")
	if err != nil {
		return err
	}

	_, _ = fmt.Fprintln(os.Stdout, fmt.Sprintf("%v", string(indexConfigBytes)))

	return nil
}
