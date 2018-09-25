// Copyright (c) 2018 Minoru Osuka
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
	"os"

	"errors"
	"github.com/blevesearch/bleve"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/mosuka/blast/node/data/client"
	"github.com/mosuka/blast/node/data/protobuf"
	"github.com/urfave/cli"
)

func search(c *cli.Context) error {
	grpcAddr := c.String("grpc-addr")

	var err error

	searchRequestBytes := []byte(c.Args().Get(0))

	if len(searchRequestBytes) <= 0 {
		err := errors.New("search request must be set")
		return err
	}

	searchRequest := bleve.NewSearchRequest(nil)
	if searchRequestBytes != nil {
		err = json.Unmarshal(searchRequestBytes, searchRequest)
		if err != nil {
			return err
		}
	}

	searchRequestAny := &any.Any{}
	err = protobuf.UnmarshalAny(searchRequest, searchRequestAny)
	if err != nil {
		return err
	}

	dataClient, err := client.NewGRPCClient(grpcAddr)
	if err != nil {
		return err
	}
	defer dataClient.Close()

	// Create request message
	req := &protobuf.SearchDocumentsRequest{
		SearchRequest: searchRequestAny,
	}

	resp, err := dataClient.SearchDocuments(req)
	if err != nil {
		return err
	}

	searchResultInstance, err := protobuf.MarshalAny(resp.SearchResult)
	if err != nil {
		return err
	}

	jsonBytes, err := json.MarshalIndent(searchResultInstance, "", "  ")
	if err != nil {
		return err
	}

	fmt.Fprintln(os.Stdout, fmt.Sprintf("%s", string(jsonBytes)))

	return nil
}
