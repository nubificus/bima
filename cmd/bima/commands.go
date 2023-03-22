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
	"github.com/urfave/cli/v2"
)

func extraBimaCommands() []*cli.Command {
	return []*cli.Command{
		{
			Name:  "version",
			Usage: "show bima version and exit",
			Action: func(ctx *cli.Context) error {
				cli.ShowVersion(ctx)
				return nil
			},
		},
		{
			Name:  "build",
			Usage: "build a container image",
			Before: func(ctx *cli.Context) error {
				if ctx.Args().Len() == 0 {
					return cli.Exit("ERROR: \"bima build\" requires requires exactly 1 argument (the context for the build process).", 1)
				}
				return nil
			},
			Flags:  buildFlags(),
			Action: bimaBuild,
		}}
}
