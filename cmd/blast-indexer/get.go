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

	"github.com/mosuka/blast/indexer"
	"github.com/mosuka/blast/protobuf"
	pbindex "github.com/mosuka/blast/protobuf/index"
	"github.com/urfave/cli"
)

func execGet(c *cli.Context) error {
	grpcAddr := c.String("grpc-addr")
	id := c.String("id")
	if id == "" {
		err := errors.New("arguments are not correct")
		return err
	}

	doc := &pbindex.Document{
		Id: id,
	}

	client, err := indexer.NewGRPCClient(grpcAddr)
	if err != nil {
		return err
	}
	defer func() {
		err := client.Close()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}()

	resp, err := client.Get(doc)
	if err != nil {
		return err
	}

	// Any -> map[string]interface{}
	var fieldsMap *map[string]interface{}
	fieldsInstance, err := protobuf.MarshalAny(resp.Fields)
	if err != nil {
		return err
	}
	if fieldsInstance == nil {
		return errors.New("nil")
	}
	fieldsMap = fieldsInstance.(*map[string]interface{})

	// map[string]interface -> []byte
	fieldsBytes, err := json.MarshalIndent(fieldsMap, "", "  ")
	if err != nil {
		return err
	}

	fmt.Fprintln(os.Stdout, fmt.Sprintf("%v\n", string(fieldsBytes)))

	return nil
}
