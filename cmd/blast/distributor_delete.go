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
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/mosuka/blast/grpc"
	"github.com/urfave/cli"
)

func distributorDelete(c *cli.Context) error {
	grpcAddr := c.String("grpc-address")
	filePath := c.String("file")
	id := c.Args().Get(0)

	ids := make([]string, 0)

	if id != "" {
		ids = append(ids, id)
	}

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

		// read index mapping file
		file, err := os.Open(filePath)
		if err != nil {
			return err
		}
		defer func() {
			_ = file.Close()
		}()

		reader := bufio.NewReader(file)
		for {
			docId, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF || err == io.ErrClosedPipe {
					if docId != "" {
						ids = append(ids, docId)
					}
					break
				}

				return err
			}

			if docId != "" {
				ids = append(ids, docId)
			}
		}
	}

	// create client
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

	result, err := client.DeleteDocument(ids)
	if err != nil {
		return err
	}

	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	_, _ = fmt.Fprintln(os.Stdout, fmt.Sprintf("%v", string(resultBytes)))

	return nil
}
