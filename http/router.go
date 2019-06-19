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

package http

import (
	"log"

	"github.com/gorilla/mux"
	"github.com/mosuka/blast/grpc"
)

type Router struct {
	mux.Router

	GRPCClient *grpc.Client
	logger     *log.Logger
}

func NewRouter(grpcAddr string, logger *log.Logger) (*Router, error) {
	grpcClient, err := grpc.NewClient(grpcAddr)
	if err != nil {
		return nil, err
	}

	return &Router{
		GRPCClient: grpcClient,
		logger:     logger,
	}, nil
}

func (r *Router) Close() error {
	r.GRPCClient.Cancel()

	err := r.GRPCClient.Close()
	if err != nil {
		return err
	}

	return nil
}
