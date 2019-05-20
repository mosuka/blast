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
	"io"
	"log"
	"os"

	"github.com/mosuka/blast/manager"
	"github.com/mosuka/blast/protobuf"
	"github.com/urfave/cli"
)

func execWatchCluster(c *cli.Context) error {
	grpcAddr := c.String("grpc-addr")

	client, err := manager.NewGRPCClient(grpcAddr)
	if err != nil {
		return err
	}
	defer func() {
		err := client.Close()
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
		}
	}()

	watchClusterClient, err := client.WatchCluster()
	if err != nil {
		return err
	}

	for {
		resp, err := watchClusterClient.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Println(err.Error())
			break
		}

		value, err := protobuf.MarshalAny(resp.Cluster)
		if err != nil {
			return err
		}
		if value == nil {
			return errors.New("nil")
		}

		var valueBytes []byte
		valueMap := *value.(*map[string]interface{})
		valueBytes, err = json.MarshalIndent(valueMap, "", "  ")
		if err != nil {
			return err
		}
		_, _ = fmt.Fprintln(os.Stdout, fmt.Sprintf("%v", string(valueBytes)))
	}

	return nil
}
