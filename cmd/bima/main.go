// Copyright 2023 Nubificus LTD.

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"os"

	l "github.com/nubificus/bima/internal/log"

	"github.com/urfave/cli/v2"
)

var (
	version string
	log     = l.Logger()
)

func main() {
	app := &cli.App{
		Name:    "bima",
		Usage:   "Create OCI compatible images for non-container deployments",
		Version: version,
		Before: func(ctx *cli.Context) error {
			if ctx.Args().Len() == 0 {
				err := cli.ShowAppHelp(ctx)
				if err != nil {
					return err
				}
				return cli.Exit("", 0)
			}
			return nil
		},
		Commands: extraBimaCommands(),
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
