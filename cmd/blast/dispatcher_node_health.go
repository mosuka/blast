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

	"github.com/mosuka/blast/dispatcher"
	"github.com/mosuka/blast/protobuf/distribute"
	"github.com/urfave/cli"
)

func dispatcherNodeHealth(c *cli.Context) error {
	grpcAddr := c.String("grpc-address")
	healthiness := c.Bool("healthiness")
	liveness := c.Bool("liveness")
	readiness := c.Bool("readiness")

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

	var res *distribute.NodeHealthCheckResponse
	if healthiness {
		req := &distribute.NodeHealthCheckRequest{Probe: distribute.NodeHealthCheckRequest_HEALTHINESS}
		res, err = client.NodeHealthCheck(req)
		if err != nil {
			res = &distribute.NodeHealthCheckResponse{State: distribute.NodeHealthCheckResponse_UNHEALTHY}
		}
	} else if liveness {
		req := &distribute.NodeHealthCheckRequest{Probe: distribute.NodeHealthCheckRequest_LIVENESS}
		res, err = client.NodeHealthCheck(req)
		if err != nil {
			res = &distribute.NodeHealthCheckResponse{State: distribute.NodeHealthCheckResponse_DEAD}
		}
	} else if readiness {
		req := &distribute.NodeHealthCheckRequest{Probe: distribute.NodeHealthCheckRequest_READINESS}
		res, err = client.NodeHealthCheck(req)
		if err != nil {
			res = &distribute.NodeHealthCheckResponse{State: distribute.NodeHealthCheckResponse_NOT_READY}
		}
	} else {
		req := &distribute.NodeHealthCheckRequest{Probe: distribute.NodeHealthCheckRequest_HEALTHINESS}
		res, err = client.NodeHealthCheck(req)
		if err != nil {
			res = &distribute.NodeHealthCheckResponse{State: distribute.NodeHealthCheckResponse_UNHEALTHY}
		}
	}

	marshaler := dispatcher.JsonMarshaler{}
	resBytes, err := marshaler.Marshal(res)
	if err != nil {
		return err
	}

	_, _ = fmt.Fprintln(os.Stdout, fmt.Sprintf("%v", string(resBytes)))

	return nil
}
