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

type CreateCollectionCommandOptions struct {
	etcdEndpoints      []string
	etcdDialTimeout    int
	etcdRequestTimeout int
	name               string
	indexMapping       string
	indexType          string
	kvstore            string
	kvconfig           string
	numberOfShards     int
}

var createCollectionCmdOpts = CreateCollectionCommandOptions{
	etcdEndpoints:      []string{"localhost:2379"},
	etcdDialTimeout:    5000,
	etcdRequestTimeout: 5000,
	name:               "",
	indexMapping:       "",
	indexType:          "upside_down",
	kvstore:            "boltdb",
	kvconfig:           "",
	numberOfShards:     1,
}

var createCollectionCmd = &cobra.Command{
	Use:   "collection",
	Short: "creates the collection",
	Long:  `The create collection command creates the collection.`,
	RunE:  runECollectionClusterCmd,
}

func runECollectionClusterCmd(cmd *cobra.Command, args []string) error {
	// check cluster name
	if createCollectionCmdOpts.name == "" {
		return fmt.Errorf("required flag: --%s", cmd.Flag("name").Name)
	}

	// IndexMapping
	indexMapping := mapping.NewIndexMapping()
	if createCollectionCmdOpts.indexMapping != "" {
		file, err := os.Open(createCollectionCmdOpts.indexMapping)
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
	if createCollectionCmdOpts.kvconfig != "" {
		file, err := os.Open(createCollectionCmdOpts.kvconfig)
		if err != nil {
			return err
		}
		defer file.Close()

		kvconfig, err = util.NewKvconfig(file)
		if err != nil {
			return err
		}
	}

	c, err := cluster.NewBlastCluster(createCollectionCmdOpts.etcdEndpoints, createCollectionCmdOpts.etcdDialTimeout)
	if err != nil {
		return err
	}
	defer c.Close()

	resp := struct {
		Succeeded bool `json:"succeeded"`
	}{
		Succeeded: true,
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(createCollectionCmdOpts.etcdRequestTimeout)*time.Millisecond)
	defer cancel()

	err = c.CreateCollection(ctx, createCollectionCmdOpts.name, indexMapping, createCollectionCmdOpts.indexType, createCollectionCmdOpts.kvstore, kvconfig, createCollectionCmdOpts.numberOfShards)
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
	createCollectionCmd.Flags().SortFlags = false

	createCollectionCmd.Flags().StringSliceVar(&createCollectionCmdOpts.etcdEndpoints, "etcd-endpoint", createCollectionCmdOpts.etcdEndpoints, "etcd eendpoint")
	createCollectionCmd.Flags().IntVar(&createCollectionCmdOpts.etcdDialTimeout, "etcd-dial-timeout", createCollectionCmdOpts.etcdDialTimeout, "etcd dial timeout")
	createCollectionCmd.Flags().IntVar(&createCollectionCmdOpts.etcdRequestTimeout, "etcd-request-timeout", createCollectionCmdOpts.etcdRequestTimeout, "etcd request timeout")
	createCollectionCmd.Flags().StringVar(&createCollectionCmdOpts.name, "name", createCollectionCmdOpts.name, "collection name")
	createCollectionCmd.Flags().StringVar(&createCollectionCmdOpts.indexMapping, "index-mapping", createCollectionCmdOpts.indexMapping, "index mapping")
	createCollectionCmd.Flags().StringVar(&createCollectionCmdOpts.indexType, "index-type", createCollectionCmdOpts.indexType, "index type")
	createCollectionCmd.Flags().StringVar(&createCollectionCmdOpts.kvstore, "kvstore", createCollectionCmdOpts.kvstore, "kvstore")
	createCollectionCmd.Flags().StringVar(&createCollectionCmdOpts.kvconfig, "kvconfig", createCollectionCmdOpts.kvconfig, "kvconfig")
	createCollectionCmd.Flags().IntVar(&createCollectionCmdOpts.numberOfShards, "number-of-shards", createCollectionCmdOpts.numberOfShards, "number of shards")

	createCmd.AddCommand(createCollectionCmd)
}
