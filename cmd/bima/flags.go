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
	"github.com/containerd/containerd/defaults"
	"github.com/urfave/cli/v2"
)

func buildFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:    "namespace",
			Aliases: []string{"n"},
			Usage:   "`NAMESPACE` to use when importing image to containerd",
			Value:   "default",
			EnvVars: []string{"CONTAINERD_NAMESPACE"},
		},
		&cli.StringFlag{
			Name:    "address",
			Aliases: []string{"a"},
			Usage:   "`ADDRESS` for containerd's GRPC server to use when importing image to containerd",
			Value:   defaults.DefaultAddress,
			EnvVars: []string{"CONTAINERD_ADDRESS"},
		},
		&cli.StringFlag{
			Name:     "snapshotter",
			Usage:    "[Optional] `SNAPSHOTTER` name. Empty value stands for the default value. Used when importing the produced image to containerd",
			Required: false,
			EnvVars:  []string{"CONTAINERD_SNAPSHOTTER"},
			Value:    "",
		},
		&cli.StringFlag{
			Name:     "output",
			Aliases:  []string{"out", "o"},
			Usage:    "[Optional] `OUTPUT` format for the produced images. Possible values: [\"ctr\", \"tar\"]",
			Required: false,
			Value:    "ctr",
		},
		&cli.BoolFlag{
			Name:     "tar",
			Usage:    "[Optional] Shorthand version of --output=tar",
			Required: false,
		},
		&cli.StringFlag{
			Name:     "tag",
			Aliases:  []string{"t"},
			Usage:    "Image `NAME` and optionally a tag (format: \"name:tag\")",
			Required: true,
		},
		&cli.StringFlag{
			Name:    "file",
			Aliases: []string{"f"},
			Usage:   "Name of the `CONTAINERFILE` ",
			Value:   "./Containerfile",
		},
	}
}
