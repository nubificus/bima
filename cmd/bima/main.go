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
