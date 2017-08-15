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
	"github.com/mosuka/blast/client"
	"github.com/spf13/cobra"
)

type DeleteClusterCommandOptions struct {
	etcdEndpoints      []string
	etcdDialTimeout    int
	etcdRequestTimeout int
	clusterName        string
}

var deleteClusterCmdOpts = DeleteClusterCommandOptions{
	etcdEndpoints:      []string{"localhost:2379"},
	etcdDialTimeout:    5000,
	etcdRequestTimeout: 5000,
	clusterName:        "",
}

var deleteClusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "deletes the cluster information",
	Long:  `The delete cluster command deletes the cluster information.`,
	RunE:  runEDeleteClusterCmd,
}

func runEDeleteClusterCmd(cmd *cobra.Command, args []string) error {
	// check cluster name
	if deleteClusterCmdOpts.clusterName == "" {
		return fmt.Errorf("required flag: --%s", cmd.Flag("cluster-name").Name)
	}

	// create client
	cw, err := client.NewEtcdClient(createClusterCmdOpts.etcdEndpoints, createClusterCmdOpts.etcdDialTimeout, createClusterCmdOpts.etcdRequestTimeout)
	if err != nil {
		return err
	}
	defer cw.Close()

	err = cw.DeleteNumberOfShards(deleteClusterCmdOpts.clusterName)
	if err != nil {
		return err
	}
	err = cw.DeleteIndexMapping(deleteClusterCmdOpts.clusterName)
	if err != nil {
		return err
	}
	err = cw.DeleteIndexType(deleteClusterCmdOpts.clusterName)
	if err != nil {
		return err
	}
	err = cw.DeleteKvstore(deleteClusterCmdOpts.clusterName)
	if err != nil {
		return err
	}

	err = cw.DeleteKvconfig(deleteClusterCmdOpts.clusterName)
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
	deleteClusterCmd.Flags().SortFlags = false

	deleteClusterCmd.Flags().StringSliceVar(&deleteClusterCmdOpts.etcdEndpoints, "etcd-endpoint", deleteClusterCmdOpts.etcdEndpoints, "etcd endpoint")
	deleteClusterCmd.Flags().IntVar(&deleteClusterCmdOpts.etcdDialTimeout, "etcd-dial-timeout", deleteClusterCmdOpts.etcdDialTimeout, "etcd dial timeout")
	deleteClusterCmd.Flags().IntVar(&deleteClusterCmdOpts.etcdRequestTimeout, "etcd-request-timeout", deleteClusterCmdOpts.etcdRequestTimeout, "etcd request timeout")
	deleteClusterCmd.Flags().StringVar(&deleteClusterCmdOpts.clusterName, "cluster-name", deleteClusterCmdOpts.clusterName, "cluster name")

	deleteCmd.AddCommand(deleteClusterCmd)
}
