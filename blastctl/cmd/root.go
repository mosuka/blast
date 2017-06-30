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
	"fmt"
	ver "github.com/mosuka/blast/version"
	"github.com/spf13/cobra"
	"os"
)

type RootCommandOptions struct {
	outputFormat string
	versionFlag  bool
}

var rootCmdOpts RootCommandOptions

var RootCmd = &cobra.Command{
	Use:               "blastctl",
	Short:             "Blast Server controller",
	Long:              `The Command Line Interface for controlling the Blast Server.`,
	PersistentPreRunE: persistentPreRunERootCmd,
	RunE:              runERootCmd,
}

func persistentPreRunERootCmd(cmd *cobra.Command, args []string) error {
	if rootCmdOpts.versionFlag {
		fmt.Printf("%s\n", ver.Version)
		os.Exit(0)
	}

	return nil
}

func runERootCmd(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return cmd.Help()
	}

	return nil
}

func init() {
	RootCmd.PersistentFlags().StringVar(&rootCmdOpts.outputFormat, "output-format", DefaultOutputFormat, "output format")
	RootCmd.PersistentFlags().BoolVar(&rootCmdOpts.versionFlag, "version", DefaultVersionFlag, "show version number")
}
