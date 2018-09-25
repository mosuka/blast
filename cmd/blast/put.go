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

	"github.com/golang/protobuf/ptypes/any"
	"github.com/mosuka/blast/node/data/client"
	"github.com/mosuka/blast/node/data/protobuf"
	"github.com/urfave/cli"
)

func put(c *cli.Context) error {
	grpcAddr := c.String("grpc-addr")

	id := c.Args().Get(0)
	fieldsBytes := []byte(c.Args().Get(1))

	var err error

	// Check id
	if id == "" {
		err := errors.New("id argument must be set")
		return err
	}

	// Check field length
	if len(fieldsBytes) <= 0 {
		err := errors.New("fields argument must be set")
		return err
	}

	var fieldsMap map[string]interface{}
	if fieldsBytes != nil {
		if err = json.Unmarshal(fieldsBytes, &fieldsMap); err != nil {
			return err
		}
	}

	fieldsAny := &any.Any{}
	if err = protobuf.UnmarshalAny(fieldsMap, fieldsAny); err != nil {
		return err
	}

	dataClient, err := client.NewGRPCClient(grpcAddr)
	if err != nil {
		return err
	}
	defer dataClient.Close()

	// Create request message
	req := &protobuf.PutDocumentRequest{
		Id:     id,
		Fields: fieldsAny,
	}

	_, err = dataClient.PutDocument(req)
	if err != nil {
		return err
	}

	return nil
}
