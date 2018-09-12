//  Copyright (c) 2018 Minoru Osuka
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

package main

import (
	"log"
	"os"
	"path"

	"github.com/mosuka/blast/version"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = path.Base(os.Args[0])
	app.Usage = "The lightweight distributed indexing and search server"
	app.Version = version.Version

	app.Authors = []cli.Author{
		{
			Name:  "mosuka",
			Email: "minoru.osuka@gmail.com",
		},
	}

	cli.HelpFlag = cli.BoolFlag{
		Name:  "help, h",
		Usage: "Show this message",
	}

	cli.VersionFlag = cli.BoolFlag{
		Name:  "version, v",
		Usage: "Print the version",
	}

	app.Commands = commands

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
