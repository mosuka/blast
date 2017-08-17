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
	"github.com/coreos/etcd/clientv3"
	"github.com/spf13/cobra"
	"time"
)

type DeleteClusterCommandOptions struct {
	etcdEndpoints      []string
	etcdDialTimeout    int
	etcdRequestTimeout int
	cluster            string
	indexMapping       bool
	indexType          bool
	kvstore            bool
	kvconfig           bool
}

var deleteClusterCmdOpts = DeleteClusterCommandOptions{
	etcdEndpoints:      []string{"localhost:2379"},
	etcdDialTimeout:    5000,
	etcdRequestTimeout: 5000,
	cluster:            "",
}

var deleteClusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "deletes the cluster information",
	Long:  `The delete cluster command deletes the cluster information.`,
	RunE:  runEDeleteClusterCmd,
}

func runEDeleteClusterCmd(cmd *cobra.Command, args []string) error {
	// check cluster name
	if deleteClusterCmdOpts.cluster == "" {
		return fmt.Errorf("required flag: --%s", cmd.Flag("cluster").Name)
	}

	if !deleteClusterCmdOpts.indexMapping && !deleteClusterCmdOpts.indexType && !deleteClusterCmdOpts.kvstore && !deleteClusterCmdOpts.kvconfig {
		deleteClusterCmdOpts.indexMapping = true
		deleteClusterCmdOpts.indexType = true
		deleteClusterCmdOpts.kvstore = true
		deleteClusterCmdOpts.kvconfig = true
	}

	cfg := clientv3.Config{
		Endpoints:   deleteClusterCmdOpts.etcdEndpoints,
		DialTimeout: time.Duration(deleteClusterCmdOpts.etcdDialTimeout) * time.Millisecond,
		Context:     context.Background(),
	}

	c, err := clientv3.New(cfg)
	if err != nil {
		return err
	}
	defer c.Close()

	var kv clientv3.KV
	if c != nil {
		kv = clientv3.NewKV(c)
	}

	resp := struct {
		IndexMapping string `json:"index_mapping,omitempty"`
		IndexType    string `json:"index_type,omitempty"`
		Kvstore      string `json:"kvstore,omitempty"`
		Kvconfig     string `json:"kvconfig,omitempty"`
	}{}

	if deleteClusterCmdOpts.indexMapping == true {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(deleteClusterCmdOpts.etcdRequestTimeout)*time.Millisecond)
		defer cancel()

		keyIndexMapping := fmt.Sprintf("/blast/clusters/%s/indexMapping", deleteClusterCmdOpts.cluster)

		_, err := kv.Delete(ctx, keyIndexMapping)
		if err != nil {
			return err
		}

		resp.IndexMapping = "deleted"
	}

	if deleteClusterCmdOpts.indexType == true {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(deleteClusterCmdOpts.etcdRequestTimeout)*time.Millisecond)
		defer cancel()

		keyIndexType := fmt.Sprintf("/blast/clusters/%s/indexType", deleteClusterCmdOpts.cluster)

		_, err := kv.Delete(ctx, keyIndexType)
		if err != nil {
			return err
		}

		resp.IndexType = "deleted"
	}

	if deleteClusterCmdOpts.kvstore == true {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(deleteClusterCmdOpts.etcdRequestTimeout)*time.Millisecond)
		defer cancel()

		keyKvstore := fmt.Sprintf("/blast/clusters/%s/kvstore", deleteClusterCmdOpts.cluster)

		_, err := kv.Delete(ctx, keyKvstore)
		if err != nil {
			return err
		}

		resp.Kvstore = "deleted"
	}

	if deleteClusterCmdOpts.kvconfig == true {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(deleteClusterCmdOpts.etcdRequestTimeout)*time.Millisecond)
		defer cancel()

		keyKvconfig := fmt.Sprintf("/blast/clusters/%s/kvconfig", deleteClusterCmdOpts.cluster)

		_, err := kv.Delete(ctx, keyKvconfig)
		if err != nil {
			return err
		}

		resp.Kvconfig = "deleted"
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
	deleteClusterCmd.Flags().SortFlags = false

	deleteClusterCmd.Flags().StringSliceVar(&deleteClusterCmdOpts.etcdEndpoints, "etcd-endpoint", deleteClusterCmdOpts.etcdEndpoints, "etcd endpoint")
	deleteClusterCmd.Flags().IntVar(&deleteClusterCmdOpts.etcdDialTimeout, "etcd-dial-timeout", deleteClusterCmdOpts.etcdDialTimeout, "etcd dial timeout")
	deleteClusterCmd.Flags().IntVar(&deleteClusterCmdOpts.etcdRequestTimeout, "etcd-request-timeout", deleteClusterCmdOpts.etcdRequestTimeout, "etcd request timeout")
	deleteClusterCmd.Flags().StringVar(&deleteClusterCmdOpts.cluster, "cluster", deleteClusterCmdOpts.cluster, "cluster name")
	deleteClusterCmd.Flags().BoolVar(&deleteClusterCmdOpts.indexMapping, "index-mapping", deleteClusterCmdOpts.indexMapping, "include index mapping")
	deleteClusterCmd.Flags().BoolVar(&deleteClusterCmdOpts.indexType, "index-type", deleteClusterCmdOpts.indexType, "include index type")
	deleteClusterCmd.Flags().BoolVar(&deleteClusterCmdOpts.kvstore, "kvstore", deleteClusterCmdOpts.kvstore, "include kvstore")
	deleteClusterCmd.Flags().BoolVar(&deleteClusterCmdOpts.kvconfig, "kvconfig", deleteClusterCmdOpts.kvconfig, "include kvconfig")

	deleteCmd.AddCommand(deleteClusterCmd)
}
