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

type DeleteDocumentCommandOptions struct {
	server         string
	requestTimeout int
	id             string
}

var deleteDocumentCmdOpts = DeleteDocumentCommandOptions{
	server:         "localhost:10000",
	requestTimeout: 15000,
	id:             "",
}

var deleteDocumentCmd = &cobra.Command{
	Use:   "document",
	Short: "deletes the document from the Blast Server",
	Long:  `The delete document command deletes the document from the Blast Server.`,
	RunE:  runEDeleteDocumentCmd,
}

func runEDeleteDocumentCmd(cmd *cobra.Command, args []string) error {
	// check id
	if deleteDocumentCmdOpts.id == "" {
		return fmt.Errorf("required flag: --%s", cmd.Flag("id").Name)
	}

	// create client
	cw, err := client.NewBlastClientWrapper(deleteDocumentCmdOpts.server, deleteDocumentCmdOpts.requestTimeout)
	if err != nil {
		return err
	}
	defer cw.Close()

	// request
	resp, err := cw.DeleteDocument(deleteDocumentCmdOpts.id)
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
	deleteDocumentCmd.Flags().SortFlags = false

	deleteDocumentCmd.Flags().StringVar(&deleteDocumentCmdOpts.server, "server", deleteDocumentCmdOpts.server, "server to connect to")
	deleteDocumentCmd.Flags().IntVar(&deleteDocumentCmdOpts.requestTimeout, "request-timeout", deleteDocumentCmdOpts.requestTimeout, "request timeout")
	deleteDocumentCmd.Flags().StringVar(&deleteDocumentCmdOpts.id, "id", deleteDocumentCmdOpts.id, "document id")

	deleteCmd.AddCommand(deleteDocumentCmd)
}
