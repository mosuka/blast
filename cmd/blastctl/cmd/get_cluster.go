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

type GetClusterCommandOptions struct {
	etcdServers    []string
	requestTimeout int
	name           string
}

var getClusterCmdOpts = GetClusterCommandOptions{
	etcdServers:    []string{"localhost:2379"},
	requestTimeout: 15000,
	name:           "",
}

var getClusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "gets the cluster info",
	Long:  `The get cluster command gets the cluster info.`,
	RunE:  runEGetClusterCmd,
}

func runEGetClusterCmd(cmd *cobra.Command, args []string) error {
	// check id
	if getClusterCmdOpts.name == "" {
		return fmt.Errorf("required flag: --%s", cmd.Flag("name").Name)
	}

	// create client
	cw, err := client.NewEtcdClientWrapper(getClusterCmdOpts.etcdServers, getClusterCmdOpts.requestTimeout)
	if err != nil {
		return err
	}
	defer cw.Close()

	// get cluster
	resp, err := cw.GetCluster(getClusterCmdOpts.name)
	if err != nil {
		return err
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
	getClusterCmd.Flags().SortFlags = false

	getClusterCmd.Flags().StringSliceVar(&getClusterCmdOpts.etcdServers, "etcd-server", getClusterCmdOpts.etcdServers, "etcd server to connect to")
	getClusterCmd.Flags().IntVar(&getClusterCmdOpts.requestTimeout, "request-timeout", getClusterCmdOpts.requestTimeout, "request timeout")
	getClusterCmd.Flags().StringVar(&getClusterCmdOpts.name, "name", getClusterCmdOpts.name, "cluster name")

	getCmd.AddCommand(getClusterCmd)
}
