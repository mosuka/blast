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

type GetClusterCommandOptions struct {
	etcdEndpoints      []string
	etcdDialTimeout    int
	etcdRequestTimeout int
	collection         string
	indexMapping       bool
	indexType          bool
	kvstore            bool
	kvconfig           bool
}

var getClusterCmdOpts = GetClusterCommandOptions{
	etcdEndpoints:      []string{"localhost:2379"},
	etcdDialTimeout:    5000,
	etcdRequestTimeout: 5000,
	collection:         "",
	indexMapping:       false,
	indexType:          false,
	kvstore:            false,
	kvconfig:           false,
}

var getClusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "gets the cluster info",
	Long:  `The get cluster command gets the cluster info.`,
	RunE:  runEGetClusterCmd,
}

func runEGetClusterCmd(cmd *cobra.Command, args []string) error {
	// check cluster name
	if getClusterCmdOpts.collection == "" {
		return fmt.Errorf("required flag: --%s", cmd.Flag("collection").Name)
	}

	if !getClusterCmdOpts.indexMapping && !getClusterCmdOpts.indexType && !getClusterCmdOpts.kvstore && !getClusterCmdOpts.kvconfig {
		getClusterCmdOpts.indexMapping = true
		getClusterCmdOpts.indexType = true
		getClusterCmdOpts.kvstore = true
		getClusterCmdOpts.kvconfig = true
	}

	c, err := cluster.NewBlastCluster(putClusterCmdOpts.etcdEndpoints, putClusterCmdOpts.etcdDialTimeout)
	if err != nil {
		return err
	}
	defer c.Close()

	resp := struct {
		IndexMapping *mapping.IndexMappingImpl `json:"index_mapping,omitempty"`
		IndexType    string                    `json:"index_type,omitempty"`
		Kvstore      string                    `json:"kvstore,omitempty"`
		Kvconfig     map[string]interface{}    `json:"kvconfig,omitempty"`
	}{}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(getClusterCmdOpts.etcdRequestTimeout)*time.Millisecond)
	defer cancel()

	if getClusterCmdOpts.indexMapping == true {
		indexMapping, err := c.GetIndexMapping(ctx, getClusterCmdOpts.collection)
		if err != nil {
			return err
		}

		resp.IndexMapping = indexMapping
	}

	if getClusterCmdOpts.indexType == true {
		indexType, err := c.GetIndexType(ctx, getClusterCmdOpts.collection)
		if err != nil {
			return err
		}

		resp.IndexType = indexType
	}

	if getClusterCmdOpts.kvstore == true {
		kvstore, err := c.GetKvstore(ctx, getClusterCmdOpts.collection)
		if err != nil {
			return err
		}

		resp.Kvstore = kvstore
	}

	if getClusterCmdOpts.kvconfig == true {
		kvconfig, err := c.GetKvconfig(ctx, getClusterCmdOpts.collection)
		if err != nil {
			return err
		}

		resp.Kvconfig = kvconfig
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
	getClusterCmd.Flags().SortFlags = false

	getClusterCmd.Flags().StringSliceVar(&getClusterCmdOpts.etcdEndpoints, "etcd-endpoint", getClusterCmdOpts.etcdEndpoints, "etcd endpoint")
	getClusterCmd.Flags().IntVar(&getClusterCmdOpts.etcdDialTimeout, "etcd-dial-timeout", getClusterCmdOpts.etcdDialTimeout, "etcd dial timeout")
	getClusterCmd.Flags().IntVar(&getClusterCmdOpts.etcdRequestTimeout, "etcd-request-timeout", getClusterCmdOpts.etcdRequestTimeout, "etcd request timeout")
	getClusterCmd.Flags().StringVar(&getClusterCmdOpts.collection, "collection", getClusterCmdOpts.collection, "collection name")
	getClusterCmd.Flags().BoolVar(&getClusterCmdOpts.indexMapping, "index-mapping", getClusterCmdOpts.indexMapping, "include index mapping")
	getClusterCmd.Flags().BoolVar(&getClusterCmdOpts.indexType, "index-type", getClusterCmdOpts.indexType, "include index type")
	getClusterCmd.Flags().BoolVar(&getClusterCmdOpts.kvstore, "kvstore", getClusterCmdOpts.kvstore, "include kvstore")
	getClusterCmd.Flags().BoolVar(&getClusterCmdOpts.kvconfig, "kvconfig", getClusterCmdOpts.kvconfig, "include kvconfig")

	getCmd.AddCommand(getClusterCmd)
}
