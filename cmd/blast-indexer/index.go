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
	"github.com/urfave/cli"
)

func execIndex(c *cli.Context) error {
	grpcAddr := c.String("grpc-addr")

	// create documents
	docs := make([]map[string]interface{}, 0)

	if c.NArg() == 1 {
		// documents
		docsStr := c.Args().Get(0)

		err := json.Unmarshal([]byte(docsStr), &docs)
		if err != nil {
			return err
		}
	} else if c.NArg() == 2 {
		// document
		id := c.Args().Get(0)
		fieldsSrc := c.Args().Get(1)

		// string -> map[string]interface{}
		var fields map[string]interface{}
		err := json.Unmarshal([]byte(fieldsSrc), &fields)
		if err != nil {
			return err
		}

		// create document
		doc := map[string]interface{}{
			"id":     id,
			"fields": fields,
		}

		docs = append(docs, doc)
	} else {
		return errors.New("argument error")
	}

	// create gRPC client
	client, err := indexer.NewGRPCClient(grpcAddr)
	if err != nil {
		return err
	}
	defer func() {
		err := client.Close()
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
		}
	}()

	// index documents in bulk
	count, err := client.IndexDocument(docs)
	if err != nil {
		return err
	}

	resultBytes, err := json.MarshalIndent(count, "", "  ")
	if err != nil {
		return err
	}

	_, _ = fmt.Fprintln(os.Stdout, fmt.Sprintf("%v", string(resultBytes)))

	return nil
}
