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
	"fmt"
	"os"

	"github.com/mosuka/blast/indexer"
	"github.com/urfave/cli"
)

func indexerNodeHealth(c *cli.Context) error {
	grpcAddr := c.String("grpc-address")
	liveness := c.Bool("liveness")
	readiness := c.Bool("readiness")

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

	if !liveness && !readiness {
		LivenessState, err := client.LivenessProbe()
		if err != nil {
			return err
		}
		_, _ = fmt.Fprintln(os.Stdout, fmt.Sprintf("%v", LivenessState))

		readinessState, err := client.ReadinessProbe()
		if err != nil {
			return err
		}
		_, _ = fmt.Fprintln(os.Stdout, fmt.Sprintf("%v", readinessState))
	} else {
		if liveness {
			state, err := client.LivenessProbe()
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintln(os.Stdout, fmt.Sprintf("%v", state))
		}
		if readiness {
			state, err := client.ReadinessProbe()
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintln(os.Stdout, fmt.Sprintf("%v", state))
		}
	}

	return nil
}
