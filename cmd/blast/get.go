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
	"errors"
	"fmt"
	"os"

	"github.com/mosuka/blast/node/data/client"
	"github.com/mosuka/blast/node/data/protobuf"
	"github.com/urfave/cli"
)

func get(c *cli.Context) error {
	grpcAddr := c.String("grpc-addr")

	id := c.Args().Get(0)

	if id == "" {
		err := errors.New("id argument must be set")
		return err
	}

	dataClient, err := client.NewGRPCClient(grpcAddr)
	if err != nil {
		return err
	}
	defer dataClient.Close()

	req := &protobuf.GetDocumentRequest{
		Id: id,
	}

	resp, err := dataClient.GetDocument(req)
	if err != nil {
		return err
	}

	// Document does not exist
	if resp.Document == nil {
		return nil
	}

	fieldsInstance, err := protobuf.MarshalAny(resp.Document.Fields)
	if err != nil {
		return err
	}

	document := struct {
		Id     string                 `json:"id,omitempty"`
		Fields map[string]interface{} `json:"fields,omitempty"`
	}{
		Id:     resp.Document.Id,
		Fields: *fieldsInstance.(*map[string]interface{}),
	}

	jsonBytes, err := json.MarshalIndent(&document, "", "  ")
	if err != nil {
		return err
	}

	fmt.Fprintln(os.Stdout, fmt.Sprintf("%v\n", string(jsonBytes)))

	return nil
}
