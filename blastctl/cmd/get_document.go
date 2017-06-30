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

type GetDocumentCommandOptions struct {
	id string
}

var getDocumentCmdOpts GetDocumentCommandOptions

var getDocumentCmd = &cobra.Command{
	Use:   "document",
	Short: "gets the document from the Blast Server",
	Long:  `The get document command gets the document from the Blast Server.`,
	RunE:  runEGetDocumentCmd,
}

func runEGetDocumentCmd(cmd *cobra.Command, args []string) error {
	// check id
	if getDocumentCmdOpts.id == "" {
		return fmt.Errorf("required flag: --%s", cmd.Flag("id").Name)
	}

	// create client
	cw, err := client.NewBlastClientWrapper(getCmdOpts.server)
	if err != nil {
		return err
	}
	defer cw.Close()

	// request
	resp, err := cw.GetDocument(context.Background(), getDocumentCmdOpts.id)
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
	getDocumentCmd.Flags().StringVar(&getDocumentCmdOpts.id, "id", DefaultId, "document id")

	getCmd.AddCommand(getDocumentCmd)
}
