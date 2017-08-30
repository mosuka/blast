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
	"context"
	"encoding/json"
	"fmt"
	"github.com/blevesearch/bleve/mapping"
	"github.com/mosuka/blast/cluster"
	"github.com/spf13/cobra"
	"time"
)

type GetCollectionCommandOptions struct {
	etcdEndpoints      []string
	etcdDialTimeout    int
	etcdRequestTimeout int
	name               string
	indexMapping       bool
	indexType          bool
	kvstore            bool
	kvconfig           bool
	numberOfShards     bool
}

var getCollectionCmdOpts = GetCollectionCommandOptions{
	etcdEndpoints:      []string{"localhost:2379"},
	etcdDialTimeout:    5000,
	etcdRequestTimeout: 5000,
	name:               "",
	indexMapping:       false,
	indexType:          false,
	kvstore:            false,
	kvconfig:           false,
	numberOfShards:     false,
}

var getCollectionCmd = &cobra.Command{
	Use:   "collection",
	Short: "gets the collection",
	Long:  `The get cluster command gets the collection.`,
	RunE:  runEGetCollectionCmd,
}

func runEGetCollectionCmd(cmd *cobra.Command, args []string) error {
	// check cluster name
	if getCollectionCmdOpts.name == "" {
		return fmt.Errorf("required flag: --%s", cmd.Flag("name").Name)
	}

	if !getCollectionCmdOpts.indexMapping && !getCollectionCmdOpts.indexType && !getCollectionCmdOpts.kvstore &&
		!getCollectionCmdOpts.kvconfig && !getCollectionCmdOpts.numberOfShards {
		getCollectionCmdOpts.indexMapping = true
		getCollectionCmdOpts.indexType = true
		getCollectionCmdOpts.kvstore = true
		getCollectionCmdOpts.kvconfig = true
		getCollectionCmdOpts.numberOfShards = true
	}

	c, err := cluster.NewBlastCluster(getCollectionCmdOpts.etcdEndpoints, getCollectionCmdOpts.etcdDialTimeout)
	if err != nil {
		return err
	}
	defer c.Close()

	resp := struct {
		IndexMapping   *mapping.IndexMappingImpl `json:"index_mapping,omitempty"`
		IndexType      string                    `json:"index_type,omitempty"`
		Kvstore        string                    `json:"kvstore,omitempty"`
		Kvconfig       map[string]interface{}    `json:"kvconfig,omitempty"`
		NumberOfShards int                       `json:"number_of_shards,omitempty"`
	}{}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(getCollectionCmdOpts.etcdRequestTimeout)*time.Millisecond)
	defer cancel()

	if getCollectionCmdOpts.indexMapping == true {
		indexMapping, _ := c.GetIndexMapping(ctx, getCollectionCmdOpts.name)
		resp.IndexMapping = indexMapping
	}

	if getCollectionCmdOpts.indexType == true {
		indexType, _ := c.GetIndexType(ctx, getCollectionCmdOpts.name)
		resp.IndexType = indexType
	}

	if getCollectionCmdOpts.kvstore == true {
		kvstore, _ := c.GetKvstore(ctx, getCollectionCmdOpts.name)
		resp.Kvstore = kvstore
	}

	if getCollectionCmdOpts.kvconfig == true {
		kvconfig, _ := c.GetKvconfig(ctx, getCollectionCmdOpts.name)
		resp.Kvconfig = kvconfig
	}

	if getCollectionCmdOpts.numberOfShards == true {
		numberOfShards, _ := c.GetNumberOfShards(ctx, getCollectionCmdOpts.name)
		resp.NumberOfShards = numberOfShards
	}

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
	getCollectionCmd.Flags().SortFlags = false

	getCollectionCmd.Flags().StringSliceVar(&getCollectionCmdOpts.etcdEndpoints, "etcd-endpoint", getCollectionCmdOpts.etcdEndpoints, "etcd endpoint")
	getCollectionCmd.Flags().IntVar(&getCollectionCmdOpts.etcdDialTimeout, "etcd-dial-timeout", getCollectionCmdOpts.etcdDialTimeout, "etcd dial timeout")
	getCollectionCmd.Flags().IntVar(&getCollectionCmdOpts.etcdRequestTimeout, "etcd-request-timeout", getCollectionCmdOpts.etcdRequestTimeout, "etcd request timeout")
	getCollectionCmd.Flags().StringVar(&getCollectionCmdOpts.name, "name", getCollectionCmdOpts.name, "collection name")
	getCollectionCmd.Flags().BoolVar(&getCollectionCmdOpts.indexMapping, "index-mapping", getCollectionCmdOpts.indexMapping, "include index mapping")
	getCollectionCmd.Flags().BoolVar(&getCollectionCmdOpts.indexType, "index-type", getCollectionCmdOpts.indexType, "include index type")
	getCollectionCmd.Flags().BoolVar(&getCollectionCmdOpts.kvstore, "kvstore", getCollectionCmdOpts.kvstore, "include kvstore")
	getCollectionCmd.Flags().BoolVar(&getCollectionCmdOpts.kvconfig, "kvconfig", getCollectionCmdOpts.kvconfig, "include kvconfig")

	getCmd.AddCommand(getCollectionCmd)
}
