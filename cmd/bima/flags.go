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
