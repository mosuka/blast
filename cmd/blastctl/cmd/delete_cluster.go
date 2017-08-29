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
	"github.com/mosuka/blast/cluster"
	"github.com/spf13/cobra"
	"time"
)

type DeleteClusterCommandOptions struct {
	etcdEndpoints      []string
	etcdDialTimeout    int
	etcdRequestTimeout int
	collection         string
}

var deleteClusterCmdOpts = DeleteClusterCommandOptions{
	etcdEndpoints:      []string{"localhost:2379"},
	etcdDialTimeout:    5000,
	etcdRequestTimeout: 5000,
	collection:         "",
}

var deleteClusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "deletes the cluster information",
	Long:  `The delete cluster command deletes the cluster information.`,
	RunE:  runEDeleteClusterCmd,
}

func runEDeleteClusterCmd(cmd *cobra.Command, args []string) error {
	// check cluster name
	if deleteClusterCmdOpts.collection == "" {
		return fmt.Errorf("required flag: --%s", cmd.Flag("collection").Name)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(deleteClusterCmdOpts.etcdRequestTimeout)*time.Millisecond)
	defer cancel()

	c, err := cluster.NewBlastCluster(deleteClusterCmdOpts.etcdEndpoints, deleteClusterCmdOpts.etcdDialTimeout)
	if err != nil {
		return err
	}
	defer c.Close()

	resp := struct {
		Succeeded bool `json:"succeeded"`
	}{
		Succeeded: true,
	}

	err = c.DeleteCollection(ctx, deleteClusterCmdOpts.collection)
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
	deleteClusterCmd.Flags().SortFlags = false

	deleteClusterCmd.Flags().StringSliceVar(&deleteClusterCmdOpts.etcdEndpoints, "etcd-endpoint", deleteClusterCmdOpts.etcdEndpoints, "etcd endpoint")
	deleteClusterCmd.Flags().IntVar(&deleteClusterCmdOpts.etcdDialTimeout, "etcd-dial-timeout", deleteClusterCmdOpts.etcdDialTimeout, "etcd dial timeout")
	deleteClusterCmd.Flags().IntVar(&deleteClusterCmdOpts.etcdRequestTimeout, "etcd-request-timeout", deleteClusterCmdOpts.etcdRequestTimeout, "etcd request timeout")
	deleteClusterCmd.Flags().StringVar(&deleteClusterCmdOpts.collection, "collection", deleteClusterCmdOpts.collection, "collection name")
	//deleteClusterCmd.Flags().BoolVar(&deleteClusterCmdOpts.indexMapping, "index-mapping", deleteClusterCmdOpts.indexMapping, "include index mapping")
	//deleteClusterCmd.Flags().BoolVar(&deleteClusterCmdOpts.indexType, "index-type", deleteClusterCmdOpts.indexType, "include index type")
	//deleteClusterCmd.Flags().BoolVar(&deleteClusterCmdOpts.kvstore, "kvstore", deleteClusterCmdOpts.kvstore, "include kvstore")
	//deleteClusterCmd.Flags().BoolVar(&deleteClusterCmdOpts.kvconfig, "kvconfig", deleteClusterCmdOpts.kvconfig, "include kvconfig")

	deleteCmd.AddCommand(deleteClusterCmd)
}
