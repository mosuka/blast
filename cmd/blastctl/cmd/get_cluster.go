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
	etcdServers    []string
	requestTimeout int
	clusterName    string
}

var getClusterCmdOpts = GetClusterCommandOptions{
	etcdServers:    []string{"localhost:2379"},
	requestTimeout: 15000,
	clusterName:    "",
}

var getClusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "gets the cluster info",
	Long:  `The get cluster command gets the cluster info.`,
	RunE:  runEGetClusterCmd,
}

func runEGetClusterCmd(cmd *cobra.Command, args []string) error {
	// check id
	if getClusterCmdOpts.clusterName == "" {
		return fmt.Errorf("required flag: --%s", cmd.Flag("cluster-name").Name)
	}

	// create client
	cw, err := client.NewEtcdClientWrapper(getClusterCmdOpts.etcdServers, getClusterCmdOpts.requestTimeout)
	if err != nil {
		return err
	}
	defer cw.Close()

	shards, err := cw.GetShards(getClusterCmdOpts.clusterName)
	if err != nil {
		return err
	}
	indexMapping, err := cw.GetIndexMapping(getClusterCmdOpts.clusterName)
	if err != nil {
		return err
	}
	indexType, err := cw.GetIndexType(getClusterCmdOpts.clusterName)
	if err != nil {
		return err
	}
	kvstore, err := cw.GetKvstore(getClusterCmdOpts.clusterName)
	if err != nil {
		return err
	}
	kvconfig, err := cw.GetKvconfig(getClusterCmdOpts.clusterName)
	if err != nil {
		return err
	}

	resp := struct {
		Shards       int                       `json:"shards,omitempty"`
		IndexMapping *mapping.IndexMappingImpl `json:"index_mapping,omitempty"`
		IndexType    string                    `json:"index_type,omitempty"`
		Kvstore      string                    `json:"kvstore,omitempty"`
		Kvconfig     map[string]interface{}    `json:"kvconfig,omitempty"`
	}{
		Shards:       shards,
		IndexMapping: indexMapping,
		IndexType:    indexType,
		Kvstore:      kvstore,
		Kvconfig:     kvconfig,
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

	getClusterCmd.Flags().StringSliceVar(&getClusterCmdOpts.etcdServers, "etcd-server", getClusterCmdOpts.etcdServers, "etcd server to connect to")
	getClusterCmd.Flags().IntVar(&getClusterCmdOpts.requestTimeout, "request-timeout", getClusterCmdOpts.requestTimeout, "request timeout")
	getClusterCmd.Flags().StringVar(&getClusterCmdOpts.clusterName, "cluster-name", getClusterCmdOpts.clusterName, "cluster name")

	getCmd.AddCommand(getClusterCmd)
}
