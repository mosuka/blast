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

type GetIndexCommandOptions struct {
	server         string
	requestTimeout int
	indexMapping   bool
	indexType      bool
	kvstore        bool
	kvconfig       bool
}

var getIndexCmdOpts = GetIndexCommandOptions{
	server:         "localhost:20884",
	requestTimeout: 15000,
	indexMapping:   false,
	indexType:      false,
	kvstore:        false,
	kvconfig:       false,
}

var getIndexCmd = &cobra.Command{
	Use:   "index",
	Short: "gets the index information",
	Long:  `The get index command gets the index information.`,
	RunE:  runEGetIndexCmd,
}

func runEGetIndexCmd(cmd *cobra.Command, args []string) error {
	if !getIndexCmdOpts.indexMapping && !getIndexCmdOpts.indexType && !getIndexCmdOpts.kvstore && !getIndexCmdOpts.kvconfig {
		getIndexCmdOpts.indexMapping = true
		getIndexCmdOpts.indexType = true
		getIndexCmdOpts.kvstore = true
		getIndexCmdOpts.kvconfig = true
	}

	// create client
	cw, err := client.NewBlastClient(getIndexCmdOpts.server, getIndexCmdOpts.requestTimeout)
	if err != nil {
		return err
	}
	defer cw.Close()

	// request
	resp, err := cw.GetIndex(getIndexCmdOpts.indexMapping, getIndexCmdOpts.indexType, getIndexCmdOpts.kvstore, getIndexCmdOpts.kvconfig)
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
	getIndexCmd.Flags().SortFlags = false

	getIndexCmd.Flags().StringVar(&getIndexCmdOpts.server, "server", getIndexCmdOpts.server, "server to connect to")
	getIndexCmd.Flags().IntVar(&getIndexCmdOpts.requestTimeout, "request-timeout", getIndexCmdOpts.requestTimeout, "request timeout")
	getIndexCmd.Flags().BoolVar(&getIndexCmdOpts.indexMapping, "index-mapping", getIndexCmdOpts.indexMapping, "include index mapping")
	getIndexCmd.Flags().BoolVar(&getIndexCmdOpts.indexType, "index-type", getIndexCmdOpts.indexType, "include index type")
	getIndexCmd.Flags().BoolVar(&getIndexCmdOpts.kvstore, "kvstore", getIndexCmdOpts.kvstore, "include kvstore")
	getIndexCmd.Flags().BoolVar(&getIndexCmdOpts.kvconfig, "kvconfig", getIndexCmdOpts.kvconfig, "include kvconfig")

	getCmd.AddCommand(getIndexCmd)
}
