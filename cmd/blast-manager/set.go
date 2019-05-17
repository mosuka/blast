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

	"github.com/mosuka/blast/manager"
	"github.com/urfave/cli"
)

func execSet(c *cli.Context) error {
	grpcAddr := c.String("grpc-addr")

	key := c.Args().Get(0)
	if key == "" {
		err := errors.New("key argument must be set")
		return err
	}

	valueStr := c.Args().Get(1)
	if valueStr == "" {
		err := errors.New("value argument must be set")
		return err
	}

	var value interface{}
	err := json.Unmarshal([]byte(valueStr), &value)
	if err != nil {
		switch err.(type) {
		case *json.SyntaxError:
			value = valueStr
		default:
			return err
		}
	}

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

	err = client.Set(key, value)
	if err != nil {
		return err
	}

	return nil
}
