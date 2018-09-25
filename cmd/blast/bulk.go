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

	"github.com/golang/protobuf/ptypes/any"
	"github.com/mosuka/blast/node/data/client"
	"github.com/mosuka/blast/node/data/protobuf"
	"github.com/urfave/cli"
)

func bulk(c *cli.Context) error {
	grpcAddr := c.String("grpc-addr")
	batchSize := c.Int("batch-size")

	updatesBytes := []byte(c.Args().Get(0))

	// Check bulk request length
	if len(updatesBytes) <= 0 {
		err := errors.New("update requests argument must be set")
		return err
	}

	updateMapSlice := make([]map[string]interface{}, 0)
	err := json.Unmarshal(updatesBytes, &updateMapSlice)
	if err != nil {
		return err
	}

	updates := make([]*protobuf.BulkUpdateRequest_Update, 0)
	for _, updateMap := range updateMapSlice {
		update := &protobuf.BulkUpdateRequest_Update{}

		update.Command = protobuf.BulkUpdateRequest_Update_Command(protobuf.BulkUpdateRequest_Update_Command_value[updateMap["command"].(string)])

		documentMap, exist := updateMap["document"].(map[string]interface{})
		if !exist {
			err := errors.New("document does not exist")
			return err
		}

		document := &protobuf.Document{}

		documentId, exist := documentMap["id"].(string)
		if !exist {
			err := errors.New("document id does not exist")
			return err
		}
		document.Id = documentId

		fieldsMap, exist := documentMap["fields"].(map[string]interface{})
		if exist {
			fieldsAny := &any.Any{}
			if err = protobuf.UnmarshalAny(fieldsMap, fieldsAny); err != nil {
				return err
			}
			document.Fields = fieldsAny
		}

		update.Document = document

		updates = append(updates, update)
	}

	req := &protobuf.BulkUpdateRequest{
		Updates:   updates,
		BatchSize: int32(batchSize),
	}

	dataClient, err := client.NewGRPCClient(grpcAddr)
	if err != nil {
		return err
	}
	defer dataClient.Close()

	resp, err := dataClient.BulkUpdate(req)
	if err != nil {
		return err
	}

	jsonBytes, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		return err
	}

	fmt.Fprintln(os.Stdout, fmt.Sprintf("%s", string(jsonBytes)))

	return nil
}
