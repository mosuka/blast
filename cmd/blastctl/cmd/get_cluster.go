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
	"github.com/spf13/cobra"
)

type GetClusterCommandOptions struct {
	etcdEndpoints         []string
	etcdDialTimeout       int
	etcdRequestTimeout    int
	clusterName           string
	includeNumberOfShards bool
	includeIndexMapping   bool
	includeIndexType      bool
	includeKvstore        bool
	includeKvconfig       bool
}

var getClusterCmdOpts = GetClusterCommandOptions{
	etcdEndpoints:         []string{"localhost:2379"},
	etcdDialTimeout:       5000,
	etcdRequestTimeout:    5000,
	clusterName:           "",
	includeNumberOfShards: false,
	includeIndexMapping:   false,
	includeIndexType:      false,
	includeKvstore:        false,
	includeKvconfig:       false,
}

var getClusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "gets the cluster info",
	Long:  `The get cluster command gets the cluster info.`,
	RunE:  runEGetClusterCmd,
}

func runEGetClusterCmd(cmd *cobra.Command, args []string) error {
	// check cluster name
	if getClusterCmdOpts.clusterName == "" {
		return fmt.Errorf("required flag: --%s", cmd.Flag("cluster-name").Name)
	}

	if !getClusterCmdOpts.includeNumberOfShards && !getClusterCmdOpts.includeIndexMapping && !getClusterCmdOpts.includeIndexType && !getClusterCmdOpts.includeKvstore && !getClusterCmdOpts.includeKvconfig {
		getClusterCmdOpts.includeNumberOfShards = true
		getClusterCmdOpts.includeIndexMapping = true
		getClusterCmdOpts.includeIndexType = true
		getClusterCmdOpts.includeKvstore = true
		getClusterCmdOpts.includeKvconfig = true
	}

	// create client
	cw, err := client.NewEtcdClient(createClusterCmdOpts.etcdEndpoints, createClusterCmdOpts.etcdDialTimeout, createClusterCmdOpts.etcdRequestTimeout)
	if err != nil {
		return err
	}
	defer cw.Close()

	resp := struct {
		NumberOfShards int                       `json:"number_of_shards,omitempty"`
		IndexMapping   *mapping.IndexMappingImpl `json:"index_mapping,omitempty"`
		IndexType      string                    `json:"index_type,omitempty"`
		Kvstore        string                    `json:"kvstore,omitempty"`
		Kvconfig       map[string]interface{}    `json:"kvconfig,omitempty"`
	}{}

	if getClusterCmdOpts.includeNumberOfShards == true {
		numberOfShards, err := cw.GetNumberOfShards(getClusterCmdOpts.clusterName)
		if err != nil {
			return err
		}

		resp.NumberOfShards = numberOfShards
	}

	if getClusterCmdOpts.includeIndexMapping == true {
		indexMapping, err := cw.GetIndexMapping(getClusterCmdOpts.clusterName)
		if err != nil {
			return err
		}

		resp.IndexMapping = indexMapping
	}

	if getClusterCmdOpts.includeIndexType == true {
		indexType, err := cw.GetIndexType(getClusterCmdOpts.clusterName)
		if err != nil {
			return err
		}

		resp.IndexType = indexType
	}

	if getClusterCmdOpts.includeKvstore == true {
		kvstore, err := cw.GetKvstore(getClusterCmdOpts.clusterName)
		if err != nil {
			return err
		}

		resp.Kvstore = kvstore
	}

	if getClusterCmdOpts.includeKvconfig == true {
		kvconfig, err := cw.GetKvconfig(getClusterCmdOpts.clusterName)
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
	getClusterCmd.Flags().StringVar(&getClusterCmdOpts.clusterName, "cluster-name", getClusterCmdOpts.clusterName, "cluster name")
	getClusterCmd.Flags().BoolVar(&getClusterCmdOpts.includeNumberOfShards, "include-number-of-shards", getClusterCmdOpts.includeNumberOfShards, "include number of shards")
	getClusterCmd.Flags().BoolVar(&getClusterCmdOpts.includeIndexMapping, "include-index-mapping", getClusterCmdOpts.includeIndexMapping, "include index mapping")
	getClusterCmd.Flags().BoolVar(&getClusterCmdOpts.includeIndexType, "include-index-type", getClusterCmdOpts.includeIndexType, "include index type")
	getClusterCmd.Flags().BoolVar(&getClusterCmdOpts.includeKvstore, "include-kvstore", getClusterCmdOpts.includeKvstore, "include kvstore")
	getClusterCmd.Flags().BoolVar(&getClusterCmdOpts.includeKvconfig, "include-kvconfig", getClusterCmdOpts.includeKvconfig, "include kvconfig")

	getCmd.AddCommand(getClusterCmd)
}
