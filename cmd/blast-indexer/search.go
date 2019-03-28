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

	"github.com/blevesearch/bleve"
	"github.com/mosuka/blast/indexer"
	"github.com/urfave/cli"
)

func execSearch(c *cli.Context) error {
	grpcAddr := c.String("grpc-addr")

	searchRequestStr := c.Args().Get(0)
	if searchRequestStr == "" {
		err := errors.New("key argument must be set")
		return err
	}

	// string -> bleve.SearchRequest
	searchRequest := bleve.NewSearchRequest(nil)
	if searchRequestStr != "" {
		err := json.Unmarshal([]byte(searchRequestStr), searchRequest)
		if err != nil {
			return err
		}
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

	searchResult, err := client.Search(searchRequest)
	if err != nil {
		return err
	}

	jsonBytes, err := json.MarshalIndent(&searchResult, "", "  ")
	if err != nil {
		return err
	}

	fmt.Fprintln(os.Stdout, fmt.Sprintf("%v\n", string(jsonBytes)))

	return nil
}
