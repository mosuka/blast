//  Copyright (c) 2017 Minoru Osuka
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

package client

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type BlastClient struct {
	Index

	cfg    Config
	ctx    context.Context
	cancel context.CancelFunc
	conn   *grpc.ClientConn
}

func NewBlastClient(config *Config) (*BlastClient, error) {
	baseCtx := context.TODO()
	if config.Context != nil {
		baseCtx = config.Context
	}

	ctx, cancel := context.WithCancel(baseCtx)

	conn, err := grpc.DialContext(ctx, config.Server, grpc.WithInsecure())
	if err != nil {
		cancel()
		return nil, err
	}

	c := &BlastClient{
		cfg:    *config,
		ctx:    ctx,
		cancel: cancel,
		conn:   conn,
	}

	c.Index = NewIndex(c)

	return c, nil
}

func (c *BlastClient) Close() error {
	c.cancel()
	if c.conn != nil {
		return c.conn.Close()
	}
	return c.ctx.Err()
}
