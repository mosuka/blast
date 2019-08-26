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

	"github.com/mosuka/blast/manager"
	"github.com/mosuka/blast/protobuf/management"
	"github.com/urfave/cli"
)

func managerNodeHealthCheck(c *cli.Context) error {
	grpcAddr := c.String("grpc-address")
	healthiness := c.Bool("healthiness")
	liveness := c.Bool("liveness")
	readiness := c.Bool("readiness")

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

	var res *management.NodeHealthCheckResponse
	if healthiness {
		req := &management.NodeHealthCheckRequest{Probe: management.NodeHealthCheckRequest_HEALTHINESS}
		res, err = client.NodeHealthCheck(req)
		if err != nil {
			res = &management.NodeHealthCheckResponse{State: management.NodeHealthCheckResponse_UNHEALTHY}
		}
	} else if liveness {
		req := &management.NodeHealthCheckRequest{Probe: management.NodeHealthCheckRequest_LIVENESS}
		res, err = client.NodeHealthCheck(req)
		if err != nil {
			res = &management.NodeHealthCheckResponse{State: management.NodeHealthCheckResponse_DEAD}
		}
	} else if readiness {
		req := &management.NodeHealthCheckRequest{Probe: management.NodeHealthCheckRequest_READINESS}
		res, err = client.NodeHealthCheck(req)
		if err != nil {
			res = &management.NodeHealthCheckResponse{State: management.NodeHealthCheckResponse_NOT_READY}
		}
	} else {
		req := &management.NodeHealthCheckRequest{Probe: management.NodeHealthCheckRequest_HEALTHINESS}
		res, err = client.NodeHealthCheck(req)
		if err != nil {
			res = &management.NodeHealthCheckResponse{State: management.NodeHealthCheckResponse_UNHEALTHY}
		}
	}

	marshaler := manager.JsonMarshaler{}
	resBytes, err := marshaler.Marshal(res)
	if err != nil {
		return err
	}

	_, _ = fmt.Fprintln(os.Stdout, fmt.Sprintf("%v", string(resBytes)))

	return nil
}
