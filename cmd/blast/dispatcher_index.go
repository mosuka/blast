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
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/golang/protobuf/ptypes/any"
	"github.com/mosuka/blast/dispatcher"
	"github.com/mosuka/blast/protobuf"
	"github.com/mosuka/blast/protobuf/distribute"
	"github.com/mosuka/blast/protobuf/index"
	"github.com/urfave/cli"
)

func dispatcherIndex(c *cli.Context) error {
	grpcAddr := c.String("grpc-address")
	filePath := c.String("file")
	bulk := c.Bool("bulk")

	// create gRPC client
	client, err := dispatcher.NewGRPCClient(grpcAddr)
	if err != nil {
		return err
	}
	defer func() {
		err := client.Close()
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
		}
	}()

	marshaler := dispatcher.JsonMarshaler{}

	if c.NArg() >= 2 {
		// index document by specifying ID and fields via standard input
		id := c.Args().Get(0)
		fieldsSrc := c.Args().Get(1)

		var fieldsMap map[string]interface{}
		err := json.Unmarshal([]byte(fieldsSrc), &fieldsMap)
		if err != nil {
			return err
		}

		fieldsAny := &any.Any{}
		err = protobuf.UnmarshalAny(fieldsMap, fieldsAny)
		if err != nil {
			return err
		}

		req := &distribute.IndexRequest{
			Id:     id,
			Fields: fieldsAny,
		}

		res, err := client.Index(req)
		if err != nil {
			return err
		}

		resBytes, err := marshaler.Marshal(res)
		if err != nil {
			return err
		}

		_, _ = fmt.Fprintln(os.Stdout, fmt.Sprintf("%v", string(resBytes)))
	} else if c.NArg() == 1 {
		// index document by specifying document(s) via standard input
		docSrc := c.Args().Get(0)

		if bulk {
			// jsonl
			docs := make([]*index.Document, 0)
			reader := bufio.NewReader(bytes.NewReader([]byte(docSrc)))
			for {
				docBytes, err := reader.ReadBytes('\n')
				if err != nil {
					if err == io.EOF || err == io.ErrClosedPipe {
						if len(docBytes) > 0 {
							doc := &index.Document{}
							err = index.UnmarshalDocument(docBytes, doc)
							if err != nil {
								return err
							}
							docs = append(docs, doc)
						}
						break
					}
				}

				if len(docBytes) > 0 {
					doc := &index.Document{}
					err = index.UnmarshalDocument(docBytes, doc)
					if err != nil {
						return err
					}
					docs = append(docs, doc)
				}
			}

			req := &distribute.BulkIndexRequest{
				Documents: docs,
			}
			res, err := client.BulkIndex(req)
			if err != nil {
				return err
			}

			resBytes, err := marshaler.Marshal(res)
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintln(os.Stdout, fmt.Sprintf("%v", string(resBytes)))
		} else {
			// json
			var docMap map[string]interface{}
			err := json.Unmarshal([]byte(docSrc), &docMap)
			if err != nil {
				return err
			}

			fieldsAny := &any.Any{}
			err = protobuf.UnmarshalAny(docMap["fields"].(map[string]interface{}), fieldsAny)
			if err != nil {
				return err
			}

			req := &distribute.IndexRequest{
				Id:     docMap["id"].(string),
				Fields: fieldsAny,
			}

			res, err := client.Index(req)
			if err != nil {
				return err
			}

			resBytes, err := marshaler.Marshal(res)
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintln(os.Stdout, fmt.Sprintf("%v", string(resBytes)))
		}
	} else {
		// index document by specifying document(s) via file
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

			if bulk {
				// jsonl
				docs := make([]*index.Document, 0)
				reader := bufio.NewReader(file)
				for {
					docBytes, err := reader.ReadBytes('\n')
					if err != nil {
						if err == io.EOF || err == io.ErrClosedPipe {
							if len(docBytes) > 0 {
								doc := &index.Document{}
								err = index.UnmarshalDocument(docBytes, doc)
								if err != nil {
									return err
								}
								docs = append(docs, doc)
							}
							break
						}
					}

					if len(docBytes) > 0 {
						doc := &index.Document{}
						err = index.UnmarshalDocument(docBytes, doc)
						if err != nil {
							return err
						}
						docs = append(docs, doc)
					}
				}

				req := &distribute.BulkIndexRequest{
					Documents: docs,
				}
				res, err := client.BulkIndex(req)
				if err != nil {
					return err
				}

				resBytes, err := marshaler.Marshal(res)
				if err != nil {
					return err
				}

				_, _ = fmt.Fprintln(os.Stdout, fmt.Sprintf("%v", string(resBytes)))
			} else {
				// json
				docBytes, err := ioutil.ReadAll(file)
				if err != nil {
					return err
				}
				var docMap map[string]interface{}
				err = json.Unmarshal(docBytes, &docMap)
				if err != nil {
					return err
				}

				fieldsAny := &any.Any{}
				err = protobuf.UnmarshalAny(docMap["fields"].(map[string]interface{}), fieldsAny)
				if err != nil {
					return err
				}

				req := &distribute.IndexRequest{
					Id:     docMap["id"].(string),
					Fields: fieldsAny,
				}

				res, err := client.Index(req)
				if err != nil {
					return err
				}

				resBytes, err := marshaler.Marshal(res)
				if err != nil {
					return err
				}

				_, _ = fmt.Fprintln(os.Stdout, fmt.Sprintf("%v", string(resBytes)))
			}
		} else {
			return errors.New("argument error")
		}
	}

	return nil
}
