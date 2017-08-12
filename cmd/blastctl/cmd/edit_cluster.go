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

type EditClusterCommandOptions struct {
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

var editClusterCmdOpts = EditClusterCommandOptions{
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

var editClusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "edits the cluster information",
	Long:  `The edit cluster command edits the cluster information.`,
	RunE:  runEEditClusterCmd,
}

func runEEditClusterCmd(cmd *cobra.Command, args []string) error {
	// check cluster name
	if editClusterCmdOpts.clusterName == "" {
		return fmt.Errorf("required flag: --%s", cmd.Flag("cluster-name").Name)
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

	resp := struct {
		NumberOfShards int                       `json:"number_of_shards,omitempty"`
		IndexMapping   *mapping.IndexMappingImpl `json:"index_mapping,omitempty"`
		IndexType      string                    `json:"index_type,omitempty"`
		Kvstore        string                    `json:"kvstore,omitempty"`
		Kvconfig       map[string]interface{}    `json:"kvconfig,omitempty"`
	}{}

	// create client
	cw, err := client.NewEtcdClientWrapper(editClusterCmdOpts.etcdServers, getClusterCmdOpts.etcdDialTimeout, getClusterCmdOpts.etcdRequestTimeout)
	if err != nil {
		return err
	}
	defer cw.Close()

	if cmd.Flag("number-of-shards").Changed {
		err = cw.PutNumberOfShards(editClusterCmdOpts.clusterName, editClusterCmdOpts.numberOfShards, false)
		if err != nil {
			return err
		}
		resp.NumberOfShards = editClusterCmdOpts.numberOfShards
	}

	if cmd.Flag("index-mapping").Changed {
		err = cw.PutIndexMapping(editClusterCmdOpts.clusterName, indexMapping, false)
		if err != nil {
			return err
		}
		resp.IndexMapping = indexMapping
	}

	if cmd.Flag("index-type").Changed {
		err = cw.PutIndexType(editClusterCmdOpts.clusterName, editClusterCmdOpts.indexType, false)
		if err != nil {
			return err
		}
		resp.IndexType = editClusterCmdOpts.indexType
	}

	if cmd.Flag("kvstore").Changed {
		err = cw.PutKvstore(editClusterCmdOpts.clusterName, editClusterCmdOpts.kvstore, false)
		if err != nil {
			return err
		}
		resp.Kvstore = editClusterCmdOpts.kvstore
	}

	if cmd.Flag("kvconfig").Changed {
		err = cw.PutKvconfig(editClusterCmdOpts.clusterName, kvconfig, false)
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
	editClusterCmd.Flags().SortFlags = false

	editClusterCmd.Flags().StringSliceVar(&editClusterCmdOpts.etcdServers, "etcd-server", editClusterCmdOpts.etcdServers, "etcd server to connect to")
	editClusterCmd.Flags().IntVar(&editClusterCmdOpts.etcdDialTimeout, "etcd-dial-timeout", editClusterCmdOpts.etcdDialTimeout, "etcd dial timeout")
	editClusterCmd.Flags().IntVar(&editClusterCmdOpts.etcdRequestTimeout, "etcd-request-timeout", editClusterCmdOpts.etcdRequestTimeout, "etcd request timeout")
	editClusterCmd.Flags().StringVar(&editClusterCmdOpts.clusterName, "cluster-name", editClusterCmdOpts.clusterName, "cluster name")
	editClusterCmd.Flags().IntVar(&editClusterCmdOpts.numberOfShards, "number-of-shards", editClusterCmdOpts.numberOfShards, "number of shards")
	editClusterCmd.Flags().StringVar(&editClusterCmdOpts.indexMapping, "index-mapping", editClusterCmdOpts.indexMapping, "index mapping")
	editClusterCmd.Flags().StringVar(&editClusterCmdOpts.indexType, "index-type", editClusterCmdOpts.indexType, "index type")
	editClusterCmd.Flags().StringVar(&editClusterCmdOpts.kvstore, "kvstore", editClusterCmdOpts.kvstore, "kvstore")
	editClusterCmd.Flags().StringVar(&editClusterCmdOpts.kvconfig, "kvconfig", editClusterCmdOpts.kvconfig, "kvconfig")

	editCmd.AddCommand(editClusterCmd)
}
