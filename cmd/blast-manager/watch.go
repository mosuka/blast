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
	"io"
	"log"
	"os"

	"github.com/mosuka/blast/manager"
	"github.com/mosuka/blast/protobuf"
	pbfederation "github.com/mosuka/blast/protobuf/management"
	"github.com/urfave/cli"
)

func execWatch(c *cli.Context) error {
	grpcAddr := c.String("grpc-addr")

	key := c.String("key")
	if key == "" {
		err := errors.New("key argument must be set")
		return err
	}

	req := &pbfederation.KeyValuePair{
		Key: key,
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

	watchClient, err := client.Watch(req)
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

		value, err := protobuf.MarshalAny(resp.Value)
		if err != nil {
			return err
		}
		if value == nil {
			return errors.New("nil")
		}

		var valueBytes []byte
		switch value.(type) {
		case *map[string]interface{}:
			valueMap := *value.(*map[string]interface{})
			valueBytes, err = json.MarshalIndent(valueMap, "", "  ")
			if err != nil {
				return err
			}
			fmt.Fprintln(os.Stdout, fmt.Sprintf("%v\n", string(valueBytes)))
		case *string:
			valueStr := *value.(*string)
			fmt.Fprintln(os.Stdout, fmt.Sprintf("%s\n", valueStr))
		default:
			fmt.Fprintln(os.Stdout, fmt.Sprintf("%v\n", &value))
		}
	}

	return nil
}
