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

package config

import (
	"github.com/mosuka/blast/maputils"
)

type ClusterConfig struct {
	maputils.Map
}

func DefaultClusterConfig() ClusterConfig {
	return ClusterConfig{Map: maputils.New()}
}

func (c ClusterConfig) SetManagerAddr(managerAddr string) error {
	err := c.Set("/manager_addr", managerAddr)
	if err != nil {
		return err
	}
	return nil
}

func (c ClusterConfig) GetManagerAddr() (string, error) {
	managerAddr, err := c.Get("/manager_addr")
	if err != nil {
		return "", err
	}
	return managerAddr.(string), nil
}

func (c ClusterConfig) SetClusterId(clusterId string) error {
	err := c.Set("/cluster_id", clusterId)
	if err != nil {
		return err
	}
	return nil
}

func (c ClusterConfig) GetClusterId() (string, error) {
	clusterId, err := c.Get("/cluster_id")
	if err != nil {
		return "", err
	}
	return clusterId.(string), nil
}

func (c ClusterConfig) SetPeerAddr(peerAddr string) error {
	err := c.Set("/peer_addr", peerAddr)
	if err != nil {
		return err
	}
	return nil
}

func (c ClusterConfig) GetPeerAddr() (string, error) {
	peerAddr, err := c.Get("/peer_addr")
	if err != nil {
		return "", err
	}
	return peerAddr.(string), nil
}
