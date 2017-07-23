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
	"github.com/buger/jsonparser"
	"github.com/mosuka/blast/client"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
)

type PutDocumentCommandOptions struct {
	server         string
	requestTimeout int
	id             string
	fields         string
	request        string
}

var putDocumentCmdOpts = PutDocumentCommandOptions{
	server:         "localhost:20884",
	requestTimeout: 15000,
	id:             "",
	fields:         "",
	request:        "",
}

var putDocumentCmd = &cobra.Command{
	Use:   "document",
	Short: "puts the document to the Blast Server",
	Long:  `The index document command puts the document to the Blast Server.`,
	RunE:  runEPutDocumentCmd,
}

func runEPutDocumentCmd(cmd *cobra.Command, args []string) error {
	// read request
	var data []byte
	var err error
	if cmd.Flag("request").Changed {
		if putDocumentCmdOpts.request == "-" {
			data, err = ioutil.ReadAll(os.Stdin)
		} else {
			file, err := os.Open(putDocumentCmdOpts.request)
			if err != nil {
				return err
			}
			defer file.Close()
			data, err = ioutil.ReadAll(file)
			if err != nil {
				return err
			}
		}
	}

	// get id
	id, err := jsonparser.GetString(data, "document", "id")
	if err != nil {
		return err
	}

	// get fields
	fieldsBytes, _, _, err := jsonparser.Get(data, "document", "fields")
	if err != nil {
		return err
	}
	var fields map[string]interface{}
	err = json.Unmarshal(fieldsBytes, &fields)
	if err != nil {
		return err
	}

	// overwrite id
	if cmd.Flag("id").Changed {
		id = putDocumentCmdOpts.id
	}

	// overwrite fields
	if cmd.Flag("fields").Changed {
		err = json.Unmarshal([]byte(putDocumentCmdOpts.fields), &fields)
		if err != nil {
			return err
		}
	}

	// create client
	cw, err := client.NewBlastClientWrapper(putDocumentCmdOpts.server, putDocumentCmdOpts.requestTimeout)
	if err != nil {
		return err
	}
	defer cw.Close()

	// request
	resp, err := cw.PutDocument(id, fields)
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
	putDocumentCmd.Flags().SortFlags = false

	putDocumentCmd.Flags().StringVar(&putDocumentCmdOpts.server, "server", putDocumentCmdOpts.server, "server to connect to")
	putDocumentCmd.Flags().IntVar(&putDocumentCmdOpts.requestTimeout, "request-timeout", putDocumentCmdOpts.requestTimeout, "request timeout")
	putDocumentCmd.Flags().StringVar(&putDocumentCmdOpts.id, "id", putDocumentCmdOpts.id, "document id")
	putDocumentCmd.Flags().StringVar(&putDocumentCmdOpts.fields, "fields", putDocumentCmdOpts.fields, "document fields")
	putDocumentCmd.Flags().StringVar(&putDocumentCmdOpts.request, "request", putDocumentCmdOpts.request, "request file")

	putCmd.AddCommand(putDocumentCmd)
}
