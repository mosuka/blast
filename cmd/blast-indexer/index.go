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
	"github.com/mosuka/blast/indexer"
	"github.com/mosuka/blast/protobuf"
	pbindex "github.com/mosuka/blast/protobuf/index"
	"github.com/urfave/cli"
)

func execIndex(c *cli.Context) error {
	grpcAddr := c.String("grpc-addr")
	id := c.String("id")

	if c.NArg() == 0 {
		err := errors.New("arguments are not correct")
		return err
	}

	// create documents
	docs := make([]*pbindex.Document, 0)

	if id == "" {
		// documents
		docsStr := c.Args().Get(0)

		var docMaps []map[string]interface{}
		err := json.Unmarshal([]byte(docsStr), &docMaps)
		if err != nil {
			return err
		}

		for _, docMap := range docMaps {
			// map[string]interface{} -> Any
			fieldsAny := &any.Any{}
			err = protobuf.UnmarshalAny(docMap["fields"], fieldsAny)
			if err != nil {
				return err
			}

			// create document
			doc := &pbindex.Document{
				Id:     docMap["id"].(string),
				Fields: fieldsAny,
			}

			docs = append(docs, doc)
		}
	} else {
		// document
		fields := c.Args().Get(0)

		// string -> map[string]interface{}
		var fieldsMap map[string]interface{}
		err := json.Unmarshal([]byte(fields), &fieldsMap)
		if err != nil {
			return err
		}

		// map[string]interface{} -> Any
		fieldsAny := &any.Any{}
		err = protobuf.UnmarshalAny(fieldsMap, fieldsAny)
		if err != nil {
			return err
		}

		// create document
		doc := &pbindex.Document{
			Id:     id,
			Fields: fieldsAny,
		}

		docs = append(docs, doc)
	}

	// create gRPC client
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

	// index documents in bulk
	result, err := client.Index(docs)
	if err != nil {
		return err
	}

	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	fmt.Fprintln(os.Stdout, fmt.Sprintf("%v\n", string(resultBytes)))

	return nil
}
