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
	"golang.org/x/net/context"
)

type GetIndexCommandOptions struct {
	includeIndexMapping bool
	includeIndexType    bool
	includeKvstore      bool
	includeKvconfig     bool
}

var getIndexCmdOpts GetIndexCommandOptions

var getIndexCmd = &cobra.Command{
	Use:   "index",
	Short: "gets the index mapping from the Blast Server",
	Long:  `The get index command gets the index information from the Blast Server.`,
	RunE:  runEGetIndexCmd,
}

func runEGetIndexCmd(cmd *cobra.Command, args []string) error {
	// create client
	cw, err := client.NewBlastClientWrapper(getCmdOpts.server)
	if err != nil {
		return err
	}
	defer cw.Conn.Close()

	// request
	resp, err := cw.GetIndex(context.Background(),
		getIndexCmdOpts.includeIndexMapping,
		getIndexCmdOpts.includeIndexType,
		getIndexCmdOpts.includeKvstore,
		getIndexCmdOpts.includeKvconfig)
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
	getIndexCmd.Flags().BoolVar(&getIndexCmdOpts.includeIndexMapping, "include-index-mapping", DefaultIncludeIndexMapping, "include index mapping")
	getIndexCmd.Flags().BoolVar(&getIndexCmdOpts.includeIndexType, "include-index-type", DefaultIncludeIndexType, "include index type")
	getIndexCmd.Flags().BoolVar(&getIndexCmdOpts.includeKvstore, "include-kvstore", DefaultIncludeKvstore, "include kvstore")
	getIndexCmd.Flags().BoolVar(&getIndexCmdOpts.includeKvconfig, "include-kvconfig", DefaultIncludeKvconfig, "include kvconfig")

	getCmd.AddCommand(getIndexCmd)
}
