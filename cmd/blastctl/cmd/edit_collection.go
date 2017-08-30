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

type EditCollectionCommandOptions struct {
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

var editCollectionCmdOpts = EditCollectionCommandOptions{
	etcdEndpoints:      []string{"localhost:2379"},
	etcdDialTimeout:    5000,
	etcdRequestTimeout: 5000,
	name:               "",
	indexMapping:       "",
	indexType:          "",
	kvstore:            "",
	kvconfig:           "",
	numberOfShards:     0,
}

var editCollectionCmd = &cobra.Command{
	Use:   "collection",
	Short: "edits the collection",
	Long:  `The edit collection command edits the collection.`,
	RunE:  runEEditCollectionCmd,
}

func runEEditCollectionCmd(cmd *cobra.Command, args []string) error {
	// check cluster name
	if editCollectionCmdOpts.name == "" {
		return fmt.Errorf("required flag: --%s", cmd.Flag("name").Name)
	}

	// IndexMapping
	indexMapping := mapping.NewIndexMapping()
	if editCollectionCmdOpts.indexMapping != "" {
		file, err := os.Open(editCollectionCmdOpts.indexMapping)
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
	if editCollectionCmdOpts.kvconfig != "" {
		file, err := os.Open(editCollectionCmdOpts.kvconfig)
		if err != nil {
			return err
		}
		defer file.Close()

		kvconfig, err = util.NewKvconfig(file)
		if err != nil {
			return err
		}
	}

	c, err := cluster.NewBlastCluster(editCollectionCmdOpts.etcdEndpoints, editCollectionCmdOpts.etcdDialTimeout)
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

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(editCollectionCmdOpts.etcdRequestTimeout)*time.Millisecond)
	defer cancel()

	if cmd.Flag("index-mapping").Changed {
		err := c.PutIndexMapping(ctx, editCollectionCmdOpts.name, indexMapping)
		if err == nil {
			resp.IndexMapping = indexMapping
		}

	}

	if cmd.Flag("index-type").Changed {
		err := c.PutIndexType(ctx, editCollectionCmdOpts.name, editCollectionCmdOpts.indexType)
		if err == nil {
			resp.IndexType = editCollectionCmdOpts.indexType
		}
	}

	if cmd.Flag("kvstore").Changed {
		err := c.PutKvstore(ctx, editCollectionCmdOpts.name, editCollectionCmdOpts.kvstore)
		if err == nil {
			resp.Kvstore = editCollectionCmdOpts.kvstore
		}
	}

	if cmd.Flag("kvconfig").Changed {
		err := c.PutKvconfig(ctx, editCollectionCmdOpts.name, kvconfig)
		if err == nil {
			resp.Kvconfig = kvconfig
		}
	}

	if cmd.Flag("number-of-shards").Changed {
		err := c.PutNumberOfShards(ctx, editCollectionCmdOpts.name, editCollectionCmdOpts.numberOfShards)
		if err == nil {
			resp.NumberOfShards = editCollectionCmdOpts.numberOfShards
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
	editCollectionCmd.Flags().SortFlags = false

	editCollectionCmd.Flags().StringSliceVar(&editCollectionCmdOpts.etcdEndpoints, "etcd-endpoint", editCollectionCmdOpts.etcdEndpoints, "etcd eendpoint")
	editCollectionCmd.Flags().IntVar(&editCollectionCmdOpts.etcdDialTimeout, "etcd-dial-timeout", editCollectionCmdOpts.etcdDialTimeout, "etcd dial timeout")
	editCollectionCmd.Flags().IntVar(&editCollectionCmdOpts.etcdRequestTimeout, "etcd-request-timeout", editCollectionCmdOpts.etcdRequestTimeout, "etcd request timeout")
	editCollectionCmd.Flags().StringVar(&editCollectionCmdOpts.name, "name", editCollectionCmdOpts.name, "collection name")
	editCollectionCmd.Flags().StringVar(&editCollectionCmdOpts.indexMapping, "index-mapping", editCollectionCmdOpts.indexMapping, "index mapping")
	editCollectionCmd.Flags().StringVar(&editCollectionCmdOpts.indexType, "index-type", editCollectionCmdOpts.indexType, "index type")
	editCollectionCmd.Flags().StringVar(&editCollectionCmdOpts.kvstore, "kvstore", editCollectionCmdOpts.kvstore, "kvstore")
	editCollectionCmd.Flags().StringVar(&editCollectionCmdOpts.kvconfig, "kvconfig", editCollectionCmdOpts.kvconfig, "kvconfig")
	editCollectionCmd.Flags().IntVar(&editCollectionCmdOpts.numberOfShards, "number-of-shards", editCollectionCmdOpts.numberOfShards, "number of shards")

	editCmd.AddCommand(editCollectionCmd)
}
