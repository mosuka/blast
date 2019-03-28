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

	"github.com/golang/protobuf/ptypes/any"
	"github.com/mosuka/blast/manager"
	"github.com/mosuka/blast/protobuf"
	pbfederation "github.com/mosuka/blast/protobuf/management"
	"github.com/urfave/cli"
)

func execSet(c *cli.Context) error {
	grpcAddr := c.String("grpc-addr")

	key := c.String("key")
	if key == "" {
		err := errors.New("key argument must be set")
		return err
	}

	value := c.Args().Get(0)
	if value == "" {
		err := errors.New("value argument must be set")
		return err
	}

	// string -> map[string]interface{}
	var valueMap map[string]interface{}
	err := json.Unmarshal([]byte(value), &valueMap)
	if err != nil {
		return err
	}

	// map[string]interface{} -> Any
	valueAny := &any.Any{}
	err = protobuf.UnmarshalAny(valueMap, valueAny)
	if err != nil {
		return err
	}

	// create PutRequest
	req := &pbfederation.KeyValuePair{
		Key:   key,
		Value: valueAny,
	}

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

	err = client.Set(req)
	if err != nil {
		return err
	}

	return nil
}
