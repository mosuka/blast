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
	etcdServers    []string
	requestTimeout int
	clusterName    string
}

var deleteClusterCmdOpts = DeleteClusterCommandOptions{
	etcdServers:    []string{"localhost:2379"},
	requestTimeout: 15000,
	clusterName:    "",
}

var deleteClusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "deletes the cluster info",
	Long:  `The delete cluster command deletes the cluster info.`,
	RunE:  runEDeleteClusterCmd,
}

func runEDeleteClusterCmd(cmd *cobra.Command, args []string) error {
	// check id
	if deleteClusterCmdOpts.clusterName == "" {
		return fmt.Errorf("required flag: --%s", cmd.Flag("cluster-name").Name)
	}

	// create client
	cw, err := client.NewEtcdClientWrapper(deleteClusterCmdOpts.etcdServers, deleteClusterCmdOpts.requestTimeout)
	if err != nil {
		return err
	}
	defer cw.Close()

	err = cw.DeleteShards(deleteClusterCmdOpts.clusterName)
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

	deleteClusterCmd.Flags().StringSliceVar(&deleteClusterCmdOpts.etcdServers, "etcd-server", deleteClusterCmdOpts.etcdServers, "etcd server to connect to")
	deleteClusterCmd.Flags().IntVar(&deleteClusterCmdOpts.requestTimeout, "request-timeout", deleteClusterCmdOpts.requestTimeout, "request timeout")
	deleteClusterCmd.Flags().StringVar(&deleteClusterCmdOpts.clusterName, "cluster-name", deleteClusterCmdOpts.clusterName, "cluster name")

	deleteCmd.AddCommand(deleteClusterCmd)
}
