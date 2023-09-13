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
