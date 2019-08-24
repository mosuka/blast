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
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/mosuka/blast/indexer"
	"github.com/mosuka/blast/protobuf/index"
	"github.com/urfave/cli"
)

func indexerDelete(c *cli.Context) error {
	grpcAddr := c.String("grpc-address")
	filePath := c.String("file")
	id := c.Args().Get(0)

	// create client
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

	marshaler := indexer.JsonMarshaler{}

	if id != "" {
		req := &index.DeleteRequest{
			Id: id,
		}
		resp, err := client.Delete(req)
		if err != nil {
			return err
		}
		respBytes, err := marshaler.Marshal(resp)
		if err != nil {
			return err
		}
		_, _ = fmt.Fprintln(os.Stdout, fmt.Sprintf("%v", string(respBytes)))
	} else {
		if filePath != "" {
			ids := make([]string, 0)

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
				docIdBytes, _, err := reader.ReadLine()
				if err != nil {
					if err == io.EOF || err == io.ErrClosedPipe {
						docId := string(docIdBytes)
						if docId != "" {
							ids = append(ids, docId)
						}
						break
					}

					return err
				}
				docId := string(docIdBytes)
				if docId != "" {
					ids = append(ids, docId)
				}
			}

			req := &index.BulkDeleteRequest{
				Ids: ids,
			}

			resp, err := client.BulkDelete(req)
			if err != nil {
				return err
			}

			resultBytes, err := marshaler.Marshal(resp)
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintln(os.Stdout, fmt.Sprintf("%v", string(resultBytes)))
		} else {
			return errors.New("argument error")
		}
	}

	return nil
}
