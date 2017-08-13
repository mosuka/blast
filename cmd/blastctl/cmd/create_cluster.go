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

package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/blevesearch/bleve/mapping"
	"github.com/mosuka/blast/client"
	"github.com/mosuka/blast/util"
	"github.com/spf13/cobra"
	"os"
)

type CreateClusterCommandOptions struct {
	etcdServers        []string
	etcdDialTimeout    int
	etcdRequestTimeout int
	clusterName        string
	numberOfShards     int
	indexMapping       string
	indexType          string
	kvstore            string
	kvconfig           string
}

var createClusterCmdOpts = CreateClusterCommandOptions{
	etcdServers:        []string{"localhost:2379"},
	etcdDialTimeout:    5000,
	etcdRequestTimeout: 5000,
	clusterName:        "",
	numberOfShards:     1,
	indexMapping:       "",
	indexType:          "upside_down",
	kvstore:            "boltdb",
	kvconfig:           "",
}

var createClusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "creates the cluster information",
	Long:  `The create cluster command creates the cluster information.`,
	RunE:  runECreateClusterCmd,
}

func runECreateClusterCmd(cmd *cobra.Command, args []string) error {
	// check cluster name
	if createClusterCmdOpts.clusterName == "" {
		return fmt.Errorf("required flag: --%s", cmd.Flag("cluster-name").Name)
	}

	// IndexMapping
	indexMapping := mapping.NewIndexMapping()
	if createClusterCmdOpts.indexMapping != "" {
		file, err := os.Open(createClusterCmdOpts.indexMapping)
		if err != nil {
			return err
		}
		defer file.Close()

		indexMapping, err = util.NewIndexMapping(file)
		if err != nil {
			return err
		}
	}

	// Kvconfig
	kvconfig := make(map[string]interface{})
	if createClusterCmdOpts.kvconfig != "" {
		file, err := os.Open(createClusterCmdOpts.kvconfig)
		if err != nil {
			return err
		}
		defer file.Close()

		kvconfig, err = util.NewKvconfig(file)
		if err != nil {
			return err
		}
	}

	resp := struct {
		NumberOfShards int                       `json:"number_of_shards,omitempty"`
		IndexMapping   *mapping.IndexMappingImpl `json:"index_mapping,omitempty"`
		IndexType      string                    `json:"index_type,omitempty"`
		Kvstore        string                    `json:"kvstore,omitempty"`
		Kvconfig       map[string]interface{}    `json:"kvconfig,omitempty"`
	}{}

	cw, err := client.NewEtcdClient(createClusterCmdOpts.etcdServers, createClusterCmdOpts.etcdDialTimeout, createClusterCmdOpts.etcdRequestTimeout)
	if err != nil {
		return err
	}
	defer cw.Close()

	err = cw.PutNumberOfShards(createClusterCmdOpts.clusterName, createClusterCmdOpts.numberOfShards, true)
	if err != nil {
		return err
	}
	resp.NumberOfShards = createClusterCmdOpts.numberOfShards

	err = cw.PutIndexMapping(createClusterCmdOpts.clusterName, indexMapping, true)
	if err != nil {
		return err
	}
	resp.IndexMapping = indexMapping

	err = cw.PutIndexType(createClusterCmdOpts.clusterName, createClusterCmdOpts.indexType, true)
	if err != nil {
		return err
	}
	resp.IndexType = createClusterCmdOpts.indexType

	err = cw.PutKvstore(createClusterCmdOpts.clusterName, createClusterCmdOpts.kvstore, true)
	if err != nil {
		return err
	}
	resp.Kvstore = createClusterCmdOpts.kvstore

	err = cw.PutKvconfig(createClusterCmdOpts.clusterName, kvconfig, true)
	if err != nil {
		return err
	}
	resp.Kvconfig = kvconfig

	// output response
	switch rootCmdOpts.outputFormat {
	case "text":
		fmt.Printf("%v\n", resp)
	case "json":
		output, err := json.MarshalIndent(resp, "", "  ")
		if err != nil {
			return err
		}
		fmt.Printf("%s\n", output)
	default:
		fmt.Printf("%v\n", resp)
	}

	return nil
}

func init() {
	createClusterCmd.Flags().SortFlags = false

	createClusterCmd.Flags().StringSliceVar(&createClusterCmdOpts.etcdServers, "etcd-server", createClusterCmdOpts.etcdServers, "etcd server to connect to")
	createClusterCmd.Flags().IntVar(&createClusterCmdOpts.etcdDialTimeout, "etcd-dial-timeout", createClusterCmdOpts.etcdDialTimeout, "etcd dial timeout")
	createClusterCmd.Flags().IntVar(&createClusterCmdOpts.etcdRequestTimeout, "etcd-request-timeout", createClusterCmdOpts.etcdRequestTimeout, "etcd request timeout")
	createClusterCmd.Flags().StringVar(&createClusterCmdOpts.clusterName, "cluster-name", createClusterCmdOpts.clusterName, "cluster name")
	createClusterCmd.Flags().IntVar(&createClusterCmdOpts.numberOfShards, "number-of-shards", createClusterCmdOpts.numberOfShards, "number of shards")
	createClusterCmd.Flags().StringVar(&createClusterCmdOpts.indexMapping, "index-mapping", createClusterCmdOpts.indexMapping, "index mapping")
	createClusterCmd.Flags().StringVar(&createClusterCmdOpts.indexType, "index-type", createClusterCmdOpts.indexType, "index type")
	createClusterCmd.Flags().StringVar(&createClusterCmdOpts.kvstore, "kvstore", createClusterCmdOpts.kvstore, "kvstore")
	createClusterCmd.Flags().StringVar(&createClusterCmdOpts.kvconfig, "kvconfig", createClusterCmdOpts.kvconfig, "kvconfig")

	createCmd.AddCommand(createClusterCmd)
}
