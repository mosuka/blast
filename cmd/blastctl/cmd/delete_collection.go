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

type DeleteCollectionCommandOptions struct {
	etcdEndpoints      []string
	etcdDialTimeout    int
	etcdRequestTimeout int
	name               string
}

var deleteCollectionCmdOpts = DeleteCollectionCommandOptions{
	etcdEndpoints:      []string{"localhost:2379"},
	etcdDialTimeout:    5000,
	etcdRequestTimeout: 5000,
	name:               "",
}

var deleteCollectionCmd = &cobra.Command{
	Use:   "collection",
	Short: "deletes the collection",
	Long:  `The delete collection command deletes the collection.`,
	RunE:  runEDeleteCollectionCmd,
}

func runEDeleteCollectionCmd(cmd *cobra.Command, args []string) error {
	// check cluster name
	if deleteCollectionCmdOpts.name == "" {
		return fmt.Errorf("required flag: --%s", cmd.Flag("name").Name)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(deleteCollectionCmdOpts.etcdRequestTimeout)*time.Millisecond)
	defer cancel()

	c, err := cluster.NewBlastCluster(deleteCollectionCmdOpts.etcdEndpoints, deleteCollectionCmdOpts.etcdDialTimeout)
	if err != nil {
		return err
	}
	defer c.Close()

	resp := struct {
		Succeeded bool `json:"succeeded"`
	}{
		Succeeded: true,
	}

	err = c.DeleteCollection(ctx, deleteCollectionCmdOpts.name)
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
	deleteCollectionCmd.Flags().SortFlags = false

	deleteCollectionCmd.Flags().StringSliceVar(&deleteCollectionCmdOpts.etcdEndpoints, "etcd-endpoint", deleteCollectionCmdOpts.etcdEndpoints, "etcd endpoint")
	deleteCollectionCmd.Flags().IntVar(&deleteCollectionCmdOpts.etcdDialTimeout, "etcd-dial-timeout", deleteCollectionCmdOpts.etcdDialTimeout, "etcd dial timeout")
	deleteCollectionCmd.Flags().IntVar(&deleteCollectionCmdOpts.etcdRequestTimeout, "etcd-request-timeout", deleteCollectionCmdOpts.etcdRequestTimeout, "etcd request timeout")
	deleteCollectionCmd.Flags().StringVar(&deleteCollectionCmdOpts.name, "name", deleteCollectionCmdOpts.name, "collection name")

	deleteCmd.AddCommand(deleteCollectionCmd)
}
