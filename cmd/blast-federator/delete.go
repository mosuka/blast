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

	"github.com/mosuka/blast/grpc"
	"github.com/urfave/cli"
)

func execDelete(c *cli.Context) error {
	grpcAddr := c.String("grpc-addr")
	id := c.Args().Get(0)

	// create documents
	ids := make([]string, 0)

	if id == "" {
		if c.NArg() == 0 {
			err := errors.New("arguments are not correct")
			return err
		}

		// documents
		docsStr := c.Args().Get(0)

		err := json.Unmarshal([]byte(docsStr), &ids)
		if err != nil {
			return err
		}
	} else {
		ids = append(ids, id)
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
