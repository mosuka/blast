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
	"github.com/mosuka/blast/protobuf"
	"github.com/urfave/cli"
)

func execIndexConfig(c *cli.Context) error {
	grpcAddr := c.String("grpc-addr")

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

	resp, err := client.GetIndexConfig()
	if err != nil {
		return err
	}

	// Any -> IndexMappingImpl
	indexMappingInstance, err := protobuf.MarshalAny(resp.IndexMapping)
	if err != nil {
		return err
	}
	if indexMappingInstance == nil {
		return errors.New("nil")
	}
	indexMapping := *indexMappingInstance.(*map[string]interface{})

	indexConfig := map[string]interface{}{
		"index_mapping":      indexMapping,
		"index_type":         resp.IndexType,
		"index_storage_type": resp.IndexStorageType,
	}

	// map[string]interface -> []byte
	fieldsBytes, err := json.MarshalIndent(indexConfig, "", "  ")
	if err != nil {
		return err
	}

	fmt.Fprintln(os.Stdout, fmt.Sprintf("%v\n", string(fieldsBytes)))

	return nil
}
