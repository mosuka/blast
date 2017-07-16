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

type PutClusterCommandOptions struct {
	etcdServers    []string
	requestTimeout int
	name           string
	shards         int
	indexPath      string
	indexMapping   string
	indexType      string
	kvstore        string
	kvconfig       string
}

var putClusterCmdOpts = PutClusterCommandOptions{
	etcdServers:    []string{"localhost:2379"},
	requestTimeout: 15000,
	name:           "",
	shards:         1,
	indexPath:      "./data/index",
	indexMapping:   "",
	indexType:      "upside_down",
	kvstore:        "boltdb",
	kvconfig:       "",
}

var putClusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "puts the cluster info",
	Long:  `The put cluster command puts the cluster info.`,
	RunE:  runEPutClusterCmd,
}

func runEPutClusterCmd(cmd *cobra.Command, args []string) error {
	// check id
	if putClusterCmdOpts.name == "" {
		return fmt.Errorf("required flag: --%s", cmd.Flag("name").Name)
	}

	// IndexMapping
	indexMapping := mapping.NewIndexMapping()
	if putClusterCmdOpts.indexMapping != "" {
		file, err := os.Open(putClusterCmdOpts.indexMapping)
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
	if putClusterCmdOpts.kvconfig != "" {
		file, err := os.Open(putClusterCmdOpts.kvconfig)
		if err != nil {
			return err
		}
		defer file.Close()

		kvconfig, err = util.NewKvconfig(file)
		if err != nil {
			return err
		}
	}

	// create client
	cw, err := client.NewEtcdClientWrapper(putClusterCmdOpts.etcdServers, getClusterCmdOpts.requestTimeout)
	if err != nil {
		return err
	}
	defer cw.Close()

	// create cluster
	err = cw.PutCluster(putClusterCmdOpts.name,
		putClusterCmdOpts.shards,
		putClusterCmdOpts.indexPath,
		indexMapping,
		putClusterCmdOpts.indexType,
		putClusterCmdOpts.kvstore,
		kvconfig)
	if err != nil {
		return err
	}

	resp := make(map[string]interface{})

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
	putClusterCmd.Flags().SortFlags = false

	putClusterCmd.Flags().StringSliceVar(&putClusterCmdOpts.etcdServers, "etcd-server", putClusterCmdOpts.etcdServers, "etcd server to connect to")
	putClusterCmd.Flags().IntVar(&putClusterCmdOpts.requestTimeout, "request-timeout", putClusterCmdOpts.requestTimeout, "request timeout")
	putClusterCmd.Flags().StringVar(&putClusterCmdOpts.name, "name", putClusterCmdOpts.name, "cluster name")
	putClusterCmd.Flags().IntVar(&putClusterCmdOpts.shards, "shards", putClusterCmdOpts.shards, "number of shards")
	putClusterCmd.Flags().StringVar(&putClusterCmdOpts.indexPath, "index-path", putClusterCmdOpts.indexPath, "index directory path")
	putClusterCmd.Flags().StringVar(&putClusterCmdOpts.indexMapping, "index-mapping", putClusterCmdOpts.indexMapping, "index mapping")
	putClusterCmd.Flags().StringVar(&putClusterCmdOpts.indexType, "index-type", putClusterCmdOpts.indexType, "index type")
	putClusterCmd.Flags().StringVar(&putClusterCmdOpts.kvstore, "kvstore", putClusterCmdOpts.kvstore, "kvstore")
	putClusterCmd.Flags().StringVar(&putClusterCmdOpts.kvconfig, "kvconfig", putClusterCmdOpts.kvconfig, "kvconfig")

	putCmd.AddCommand(putClusterCmd)
}
