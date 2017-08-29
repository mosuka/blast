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
	"github.com/mosuka/blast/util"
	"github.com/spf13/cobra"
	"os"
	"time"
)

type CreateClusterCommandOptions struct {
	etcdEndpoints      []string
	etcdDialTimeout    int
	etcdRequestTimeout int
	collection         string
	indexMapping       string
	indexType          string
	kvstore            string
	kvconfig           string
	numberOfShards     int
}

var createClusterCmdOpts = CreateClusterCommandOptions{
	etcdEndpoints:      []string{"localhost:2379"},
	etcdDialTimeout:    5000,
	etcdRequestTimeout: 5000,
	collection:         "",
	indexMapping:       "",
	indexType:          "upside_down",
	kvstore:            "boltdb",
	kvconfig:           "",
	numberOfShards:     1,
}

var createClusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "creates the cluster information",
	Long:  `The create cluster command creates the cluster information.`,
	RunE:  runECreateClusterCmd,
}

func runECreateClusterCmd(cmd *cobra.Command, args []string) error {
	// check cluster name
	if createClusterCmdOpts.collection == "" {
		return fmt.Errorf("required flag: --%s", cmd.Flag("collection").Name)
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

	c, err := cluster.NewBlastCluster(createClusterCmdOpts.etcdEndpoints, createClusterCmdOpts.etcdDialTimeout)
	if err != nil {
		return err
	}
	defer c.Close()

	resp := struct {
		Succeeded bool `json:"succeeded"`
	}{
		Succeeded: true,
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(createClusterCmdOpts.etcdRequestTimeout)*time.Millisecond)
	defer cancel()

	err = c.CreateCollection(ctx, createClusterCmdOpts.collection, indexMapping, createClusterCmdOpts.indexType, createClusterCmdOpts.kvstore, kvconfig, createClusterCmdOpts.numberOfShards)
	if err != nil {
		resp.Succeeded = false
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
	createClusterCmd.Flags().SortFlags = false

	createClusterCmd.Flags().StringSliceVar(&createClusterCmdOpts.etcdEndpoints, "etcd-endpoint", createClusterCmdOpts.etcdEndpoints, "etcd eendpoint")
	createClusterCmd.Flags().IntVar(&createClusterCmdOpts.etcdDialTimeout, "etcd-dial-timeout", createClusterCmdOpts.etcdDialTimeout, "etcd dial timeout")
	createClusterCmd.Flags().IntVar(&createClusterCmdOpts.etcdRequestTimeout, "etcd-request-timeout", createClusterCmdOpts.etcdRequestTimeout, "etcd request timeout")
	createClusterCmd.Flags().StringVar(&createClusterCmdOpts.collection, "collection", createClusterCmdOpts.collection, "cluster name")
	createClusterCmd.Flags().StringVar(&createClusterCmdOpts.indexMapping, "index-mapping", createClusterCmdOpts.indexMapping, "index mapping")
	createClusterCmd.Flags().StringVar(&createClusterCmdOpts.indexType, "index-type", createClusterCmdOpts.indexType, "index type")
	createClusterCmd.Flags().StringVar(&createClusterCmdOpts.kvstore, "kvstore", createClusterCmdOpts.kvstore, "kvstore")
	createClusterCmd.Flags().StringVar(&createClusterCmdOpts.kvconfig, "kvconfig", createClusterCmdOpts.kvconfig, "kvconfig")
	createClusterCmd.Flags().IntVar(&createClusterCmdOpts.numberOfShards, "number-of-shards", createClusterCmdOpts.numberOfShards, "number of shards")

	createCmd.AddCommand(createClusterCmd)
}
