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
	"time"
)

type GetDocumentCommandOptions struct {
	server         string
	dialTimeout    int
	requestTimeout int
	id             string
}

var getDocumentCmdOpts = GetDocumentCommandOptions{
	server:         "localhost:5000",
	dialTimeout:    5000,
	requestTimeout: 5000,
	id:             "",
}

var getDocumentCmd = &cobra.Command{
	Use:   "document",
	Short: "gets the document",
	Long:  `The get document command gets the document.`,
	RunE:  runEGetDocumentCmd,
}

func runEGetDocumentCmd(cmd *cobra.Command, args []string) error {
	// check id
	if getDocumentCmdOpts.id == "" {
		return fmt.Errorf("required flag: --%s", cmd.Flag("id").Name)
	}

	// create client config
	cfg := client.Config{
		Server:      getDocumentCmdOpts.server,
		DialTimeout: time.Duration(getDocumentCmdOpts.dialTimeout) * time.Millisecond,
	}

	// create client
	cl, err := client.NewClient(&cfg)
	if err != nil {
		return err
	}
	defer cl.Close()

	// create index
	idx := client.NewIndex(cl)

	// create context
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(getDocumentCmdOpts.requestTimeout)*time.Millisecond)
	defer cancel()

	// get document from index
	resp, _ := idx.GetDocument(ctx, getDocumentCmdOpts.id)

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
	getDocumentCmd.Flags().SortFlags = false

	getDocumentCmd.Flags().StringVar(&getDocumentCmdOpts.server, "server", getDocumentCmdOpts.server, "server to connect to")
	getDocumentCmd.Flags().IntVar(&getDocumentCmdOpts.dialTimeout, "dial-timeout", getDocumentCmdOpts.dialTimeout, "dial timeout")
	getDocumentCmd.Flags().IntVar(&getDocumentCmdOpts.requestTimeout, "request-timeout", getDocumentCmdOpts.requestTimeout, "request timeout")
	getDocumentCmd.Flags().StringVar(&getDocumentCmdOpts.id, "id", getDocumentCmdOpts.id, "document id")

	getCmd.AddCommand(getDocumentCmd)
}
