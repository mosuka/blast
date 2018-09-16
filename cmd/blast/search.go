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
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"github.com/mosuka/blast/node/data/client"
	"github.com/mosuka/blast/node/data/protobuf"
	"github.com/urfave/cli"
)

func search(c *cli.Context) {
	grpcAddr := c.String("grpc-addr")
	prettyPrint := c.Bool("pretty-print")

	searchRequest := []byte(c.Args().Get(0))

	var err error

	if len(searchRequest) <= 0 {
		fmt.Fprintln(os.Stderr, "search request must be set")
		return
	}

	var dataClient *client.GRPCClient
	if dataClient, err = client.NewGRPCClient(grpcAddr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	defer dataClient.Close()

	req := &protobuf.SearchRequest{}
	if err = req.SetBytes(searchRequest); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	var resp *protobuf.SearchResponse
	if resp, err = dataClient.Search(req); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	var jsonBytes []byte
	if jsonBytes, err = resp.GetBytes(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	if prettyPrint {
		var buff bytes.Buffer
		if err = json.Indent(&buff, jsonBytes, "", "  "); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}
		jsonBytes = buff.Bytes()
	}

	fmt.Fprintln(os.Stdout, fmt.Sprintf("%s", string(jsonBytes)))
	return
}
