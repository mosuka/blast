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
	"fmt"
	"io/ioutil"
	"os"

	"github.com/blevesearch/bleve"
	"github.com/mosuka/blast/grpc"
	"github.com/urfave/cli"
)

func indexerSearch(c *cli.Context) error {
	grpcAddr := c.String("grpc-address")
	searchRequestPath := c.String("file")

	searchRequest := bleve.NewSearchRequest(nil)

	if searchRequestPath != "" {
		_, err := os.Stat(searchRequestPath)
		if err != nil {
			if os.IsNotExist(err) {
				// does not exist
				return err
			}
			// other error
			return err
		}

		// open file
		searchRequestFile, err := os.Open(searchRequestPath)
		if err != nil {
			return err
		}
		defer func() {
			_ = searchRequestFile.Close()
		}()

		// read file
		searchRequestBytes, err := ioutil.ReadAll(searchRequestFile)
		if err != nil {
			return err
		}

		// create search request
		if searchRequestBytes != nil {
			err := json.Unmarshal(searchRequestBytes, searchRequest)
			if err != nil {
				return err
			}
		}
	}

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

	searchResult, err := client.Search(searchRequest)
	if err != nil {
		return err
	}

	jsonBytes, err := json.MarshalIndent(&searchResult, "", "  ")
	if err != nil {
		return err
	}

	_, _ = fmt.Fprintln(os.Stdout, fmt.Sprintf("%v", string(jsonBytes)))

	return nil
}
