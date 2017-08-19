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
	"fmt"
	"google.golang.org/grpc"
)

type BlastMultiClient struct {
	clients        map[string]*BlastClient
	requestTimeout int
}

func NewBlastMultiClient(servers []string, dialTimeout int, requestTimeout int) (*BlastMultiClient, error) {

	clients := make(map[string]*BlastClient, 0)

	for _, server := range servers {
		client, err := NewBlastClient(server, dialTimeout, requestTimeout)
		if err != nil {
			fmt.Errorf(err.Error())
		}
		clients[server] = client
	}

	return &BlastMultiClient{
		clients:        clients,
		requestTimeout: requestTimeout,
	}, nil
}

func (c *BlastMultiClient) GetDocument(id string, opts ...grpc.CallOption) (*GetDocumentResponse, error) {

	resp, err := c.clients["hoge"].GetDocument(id, opts...)

	return resp, err
}
