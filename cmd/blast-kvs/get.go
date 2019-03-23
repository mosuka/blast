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
	"errors"
	"fmt"
	"os"

	"github.com/mosuka/blast/kvs"
	pbkvs "github.com/mosuka/blast/protobuf/kvs"
	"github.com/urfave/cli"
)

func execGet(c *cli.Context) error {
	grpcAddr := c.String("grpc-addr")

	key := c.String("key")
	if key == "" {
		err := errors.New("key argument must be set")
		return err
	}

	req := &pbkvs.KeyValuePair{
		Key: []byte(key),
	}

	client, err := kvs.NewGRPCClient(grpcAddr)
	if err != nil {
		return err
	}
	defer func() {
		err := client.Close()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}()

	resp, err := client.Get(req)
	if err != nil {
		return err
	}

	// key does not exist
	if resp.Value == nil {
		return nil
	}

	fmt.Fprintln(os.Stdout, fmt.Sprintf("%v", string(resp.Value)))

	return nil
}
