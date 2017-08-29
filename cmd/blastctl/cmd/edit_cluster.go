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

type EditClusterCommandOptions struct {
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

var editClusterCmdOpts = EditClusterCommandOptions{
	etcdEndpoints:      []string{"localhost:2379"},
	etcdDialTimeout:    5000,
	etcdRequestTimeout: 5000,
	collection:         "",
	indexMapping:       "",
	indexType:          "",
	kvstore:            "",
	kvconfig:           "",
	numberOfShards:     0,
}

var editClusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "edits the cluster information",
	Long:  `The edit cluster command edits the cluster information.`,
	RunE:  runEEditClusterCmd,
}

func runEEditClusterCmd(cmd *cobra.Command, args []string) error {
	// check cluster name
	if editClusterCmdOpts.collection == "" {
		return fmt.Errorf("required flag: --%s", cmd.Flag("collection").Name)
	}

	// IndexMapping
	indexMapping := mapping.NewIndexMapping()
	if editClusterCmdOpts.indexMapping != "" {
		file, err := os.Open(editClusterCmdOpts.indexMapping)
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
	if editClusterCmdOpts.kvconfig != "" {
		file, err := os.Open(editClusterCmdOpts.kvconfig)
		if err != nil {
			return err
		}
		defer file.Close()

		kvconfig, err = util.NewKvconfig(file)
		if err != nil {
			return err
		}
	}

	c, err := cluster.NewBlastCluster(editClusterCmdOpts.etcdEndpoints, editClusterCmdOpts.etcdDialTimeout)
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

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(editClusterCmdOpts.etcdRequestTimeout)*time.Millisecond)
	defer cancel()

	if cmd.Flag("index-mapping").Changed {
		err := c.PutIndexMapping(ctx, editClusterCmdOpts.collection, indexMapping)
		if err == nil {
			resp.IndexMapping = indexMapping
		}

	}

	if cmd.Flag("index-type").Changed {
		err := c.PutIndexType(ctx, editClusterCmdOpts.collection, editClusterCmdOpts.indexType)
		if err == nil {
			resp.IndexType = editClusterCmdOpts.indexType
		}
	}

	if cmd.Flag("kvstore").Changed {
		err := c.PutKvstore(ctx, editClusterCmdOpts.collection, editClusterCmdOpts.kvstore)
		if err == nil {
			resp.Kvstore = editClusterCmdOpts.kvstore
		}
	}

	if cmd.Flag("kvconfig").Changed {
		err := c.PutKvconfig(ctx, editClusterCmdOpts.collection, kvconfig)
		if err == nil {
			resp.Kvconfig = kvconfig
		}
	}

	if cmd.Flag("number-of-shards").Changed {
		err := c.PutNumberOfShards(ctx, editClusterCmdOpts.collection, editClusterCmdOpts.numberOfShards)
		if err == nil {
			resp.NumberOfShards = editClusterCmdOpts.numberOfShards
		}
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
	editClusterCmd.Flags().SortFlags = false

	editClusterCmd.Flags().StringSliceVar(&editClusterCmdOpts.etcdEndpoints, "etcd-endpoint", editClusterCmdOpts.etcdEndpoints, "etcd eendpoint")
	editClusterCmd.Flags().IntVar(&editClusterCmdOpts.etcdDialTimeout, "etcd-dial-timeout", editClusterCmdOpts.etcdDialTimeout, "etcd dial timeout")
	editClusterCmd.Flags().IntVar(&editClusterCmdOpts.etcdRequestTimeout, "etcd-request-timeout", editClusterCmdOpts.etcdRequestTimeout, "etcd request timeout")
	editClusterCmd.Flags().StringVar(&editClusterCmdOpts.collection, "collection", editClusterCmdOpts.collection, "cluster name")
	editClusterCmd.Flags().StringVar(&editClusterCmdOpts.indexMapping, "index-mapping", editClusterCmdOpts.indexMapping, "index mapping")
	editClusterCmd.Flags().StringVar(&editClusterCmdOpts.indexType, "index-type", editClusterCmdOpts.indexType, "index type")
	editClusterCmd.Flags().StringVar(&editClusterCmdOpts.kvstore, "kvstore", editClusterCmdOpts.kvstore, "kvstore")
	editClusterCmd.Flags().StringVar(&editClusterCmdOpts.kvconfig, "kvconfig", editClusterCmdOpts.kvconfig, "kvconfig")
	editClusterCmd.Flags().IntVar(&editClusterCmdOpts.numberOfShards, "number-of-shards", editClusterCmdOpts.numberOfShards, "number of shards")

	editCmd.AddCommand(editClusterCmd)
}
