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
	"io/ioutil"
	"os"

	"github.com/blevesearch/bleve"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/mosuka/blast/indexer"
	"github.com/mosuka/blast/protobuf"
	"github.com/mosuka/blast/protobuf/index"
	"github.com/urfave/cli"
)

func indexerSearch(c *cli.Context) error {
	grpcAddr := c.String("grpc-address")
	filePath := c.String("file")

	searchRequest := bleve.NewSearchRequest(nil)

	if filePath != "" {
		_, err := os.Stat(filePath)
		if err != nil {
			if os.IsNotExist(err) {
				// does not exist
				return err
			}
			// other error
			return err
		}

		// open file
		file, err := os.Open(filePath)
		if err != nil {
			return err
		}
		defer func() {
			_ = file.Close()
		}()

		// read file
		fileBytes, err := ioutil.ReadAll(file)
		if err != nil {
			return err
		}

		// create search request
		if fileBytes != nil {
			var tmpValue map[string]interface{}
			err = json.Unmarshal(fileBytes, &tmpValue)
			if err != nil {
				return err
			}
			searchRequestMap, ok := tmpValue["search_request"]
			if !ok {
				return errors.New("search_request does not exist")
			}
			searchRequestBytes, err := json.Marshal(searchRequestMap)
			if err != nil {
				return err
			}
			err = json.Unmarshal(searchRequestBytes, &searchRequest)
			if err != nil {
				return err
			}
		}
	}

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

	searchRequestAny := &any.Any{}
	err = protobuf.UnmarshalAny(searchRequest, searchRequestAny)
	if err != nil {
		return err
	}

	req := &index.SearchRequest{SearchRequest: searchRequestAny}

	res, err := client.Search(req)
	if err != nil {
		return err
	}

	marshaler := indexer.JsonMarshaler{}
	resBytes, err := marshaler.Marshal(res)
	if err != nil {
		return err
	}

	_, _ = fmt.Fprintln(os.Stdout, fmt.Sprintf("%v", string(resBytes)))

	return nil
}
