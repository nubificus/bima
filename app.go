package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/containerd/containerd/defaults"
	"github.com/google/go-containerregistry/pkg/crane"
	image "github.com/nubificus/bima/internal/image"
	"github.com/nubificus/bima/internal/utils"
	"github.com/urfave/cli/v2"
)

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
	extraFile := ctx.StringSlice("extra")
	cmdLine := ctx.String("cmdline")
	cpuArch := ctx.String("architecture")
	namespace := ctx.String("namespace")
	importFlag := ctx.Bool("import")
	remoteRegistry := ctx.String("remote")

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
	for ind, val := range extraFile {
		extraFile[ind], err = utils.ValidAbsolutePath(val)
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
	imageConfig := image.UnikernelImageConfig{
		Name:      imageName,
		Type:      unikernelType,
		Unikernel: unikernelBinary,
		Extra:     extraFile,
		Arch:      cpuArch,
		CmdLine:   cmdLine,
	}

	image, err := image.CreateUnikernelImage(imageConfig)
	if err != nil {
		return err
	}
	name, tag := utils.SplitImageName(imageName)
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	target := filepath.Join(cwd, name+":"+tag) + ".tar"
	err = crane.Save(image, name, target)
	if err != nil {
		return err
	}

	if importFlag {
		err = utils.CtrImportImage(target, name, ctx.String("address"), namespace, ctx.String("snapshotter"))
		if err != nil {
			return err
		}
	}

	if remoteRegistry != "" {
		err = utils.PushImage(image, remoteRegistry, imageName)
		if err != nil {
			return err
		}
		fmt.Printf("Image %s pushed to remote registry\n", imageName)
	}
	if !importFlag && remoteRegistry == "" {
		fmt.Printf("Image tarball saved at %s\n", target)
	} else {
		os.RemoveAll(target)
	}
	return nil

}
func Bima() *cli.App {
	return &cli.App{
		Name:    "bima",
		Version: "0.0.2",
		Usage:   "Create OCI compatible images for non-container deployments",
		Action:  action,
		Flags: []cli.Flag{
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
		},
		Before: func(ctx *cli.Context) error {
			if ctx.Args().Len() == 0 {
				cli.ShowAppHelp(ctx)
				return cli.Exit("", 0)
			}
			return nil
		},
		Commands: []*cli.Command{
			{
				Name:    "unikernel",
				Aliases: []string{"u"},
				Usage:   "build a unikernel image",
				Action:  action,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "unikernel",
						Aliases:  []string{"u"},
						Usage:    "The `UNIKERNEL` binary you want to package.",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "image",
						Aliases:  []string{"i"},
						Usage:    "The `NAME` of the image you want to create.",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "type",
						Aliases:  []string{"t"},
						Usage:    "The unikernel `TYPE`.",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "architecture",
						Aliases:  []string{"a", "arch"},
						Usage:    "The image's `ARCH`.",
						Required: false,
						Value:    "",
					},
					&cli.StringSliceFlag{
						Name:      "extra",
						Aliases:   []string{"e"},
						Usage:     "Path of the `EXTRA` file or directory you want to include in the container image.",
						Required:  false,
						TakesFile: true,
					},
					&cli.StringFlag{
						Name:     "cmdline",
						Aliases:  []string{"c"},
						Usage:    "The `CMDLINE` you want to include in the container image annotations.",
						Required: false,
						Value:    "",
					},
					&cli.BoolFlag{
						Name:     "import",
						Usage:    "[Optional] Use to import the produced image to containerd",
						Required: false,
					},
					&cli.StringFlag{
						Name:     "remote",
						Usage:    "[Optional] The remote registry details to push the img (eg `USERNAME:PASSWORD@REGISTRY`)",
						Required: false,
					},
				},
			}},
	}
}
