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

	"github.com/mosuka/blast/index"
	"github.com/mosuka/blast/protobuf"
	"github.com/urfave/cli"
)

func execStats(c *cli.Context) error {
	grpcAddr := c.String("grpc-addr")

	client, err := index.NewGRPCClient(grpcAddr)
	if err != nil {
		return err
	}
	defer func() {
		err := client.Close()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}()

	resp, err := client.GetIndexStats()
	if err != nil {
		return err
	}

	// Any -> map[string]interface{}
	var statsMap *map[string]interface{}
	statsInstance, err := protobuf.MarshalAny(resp.Stats)
	if err != nil {
		return err
	}
	if statsInstance == nil {
		return errors.New("nil")
	}
	statsMap = statsInstance.(*map[string]interface{})

	// map[string]interface -> []byte
	fieldsBytes, err := json.MarshalIndent(statsMap, "", "  ")
	if err != nil {
		return err
	}

	fmt.Fprintln(os.Stdout, fmt.Sprintf("%v\n", string(fieldsBytes)))

	return nil
}
