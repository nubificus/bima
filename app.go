package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/go-containerregistry/pkg/crane"
	unikernel "github.com/nubificus/bima/pkg/unikernel"
	"github.com/nubificus/bima/pkg/utils"
	"github.com/urfave/cli/v2"
)

var Flags = []cli.Flag{
	&cli.StringFlag{
		Name:     "unikernel",
		Aliases:  []string{"u"},
		Usage:    "The unikernel binary you want to package.",
		Required: true,
	},
	&cli.StringFlag{
		Name:     "image",
		Aliases:  []string{"i"},
		Usage:    "The name of the image you want to create.",
		Required: true,
	},
	&cli.StringFlag{
		Name:     "type",
		Aliases:  []string{"t"},
		Usage:    "The unikernel type.",
		Required: true,
	},
	&cli.StringFlag{
		Name:     "architecture",
		Aliases:  []string{"a", "arch"},
		Usage:    "The image's architecture.",
		Required: false,
		Value:    "",
	},
	&cli.StringFlag{
		Name:     "extra",
		Aliases:  []string{"e"},
		Usage:    "Path of the extra file or directory you want to include in the container image.",
		Required: false,
	},
	&cli.StringFlag{
		Name:     "cmdline",
		Aliases:  []string{"c"},
		Usage:    "The cmdline you want to include in the container image annotations.",
		Required: false,
		Value:    "",
	},
}

func action(ctx *cli.Context) (err error) {
	defer func() {
		if err != nil {
			fmt.Println("ERR: " + err.Error())
			os.Exit(1)
		}
	}()

	// Get parameters
	unikernelBinary := ctx.String("unikernel")
	imageName := ctx.String("image")
	unikernelType := ctx.String("type")
	extraFile := ctx.String("extra")
	cmdLine := ctx.String("cmdline")
	cpuArch := ctx.String("architecture")

	// check image name
	if !utils.ValidImageName(imageName) {
		err = fmt.Errorf("given image name is not compatible with containerd")
		return err
	}

	// get absolute path for unikernel binary and check it is a valid
	unikernelBinary, err = utils.ValidAbsoluteFilePath(unikernelBinary)
	if err != nil {
		return err
	}

	// get absolute path for extra file, if given
	if extraFile != "" {
		extraFile, err = utils.ValidAbsolutePath(extraFile)
		if err != nil {
			return err
		}
	}

	// if cpu architecture is not provided, try to determine by reading ELF header
	if cpuArch == "" {
		cpuArch, err = utils.GetBinaryArchitecture(unikernelBinary)
		if err != nil {
			return err
		}
	}
	imageConfig := unikernel.UnikernelImageConfig{
		Name:      imageName,
		Type:      unikernelType,
		Unikernel: unikernelBinary,
		Extra:     extraFile,
		Arch:      cpuArch,
		CmdLine:   cmdLine,
	}

	image, err := unikernel.CreateImage(imageConfig)
	if err != nil {
		return err
	}
	name, tag := utils.SplitImageName(imageName)
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	target := filepath.Join(cwd, name) + ".tar"
	err = crane.Save(image, tag, target)
	if err != nil {
		return err
	}
	// TODO: Import image to ctr (?)
	return nil

}
func Bima() *cli.App {
	app := cli.NewApp()
	app.Name = "bima"
	app.Version = "0.0.2"

	// TODO: add better description
	app.Description = "bima is a simple CLI tool to build OCI images for urunc runtime."
	app.Action = action
	return app
}
