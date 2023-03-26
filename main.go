package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/google/go-containerregistry/pkg/v1/partial"

	cp "github.com/otiai10/copy"

	unik "github.com/nubificus/bima/pkg/unikernel"
	"github.com/urfave/cli/v2"
)

func main() {

	imageTag := "docker.io/library/alpine:3.17.3"
	manifest, err := crane.Manifest(imageTag)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(manifest))

	imageConfig, err := crane.Config(imageTag)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(imageConfig))
	// partial.ConfigLayer(imageConfig)

	testHello, err := crane.Pull(imageTag)
	if err != nil {
		panic(err)
	}
	configFile, err := partial.ConfigFile(testHello)
	// fmt.Println(&configFile)
	// fmt.Println(configFile)
	fmt.Println("Architecture: ", configFile.Architecture)
	fmt.Println("OS: ", configFile.OS)
	// fmt.Println("RootFS: ", configFile.RootFS)
	fmt.Println("RootFS type: ", configFile.RootFS.Type)
	fmt.Println("RootFS diff ids: ", configFile.RootFS.DiffIDs)
	fmt.Println("Config: ", configFile.Config)

	os.Exit(0)
	m := "/home/gntouts/develop/bima/Makefile"
	n := "/home/gntouts/develop/bima/"
	// unik.NewUnikernelImage(m, m, m, m, m)
	image := unik.NewUnikernelImage()
	err = image.AddUnikernelFile(m)
	if err != nil {
		panic(err)
	}
	err = image.AddExtraFiles(n)
	if err != nil {
		panic(err)
	}
	annot := map[string]string{
		"test":  "ok",
		"hello": "world",
	}
	image.AddAnnotations(annot)
	// err = image.SaveOCI("/home/gntouts/develop/bima/image")
	// if err != nil {
	// 	panic(err)
	// }
	image.Trouble()

	err = image.Save("/home/gntouts/develop/bima/image.tar", "gntouts/bima:0.1")
	if err != nil {
		panic(err)
	}
	os.Exit(0)
	app := cli.NewApp()
	app.Name = "bima"
	app.Version = "0.0.1"
	app.Description = "bima is a simple CLI tool to build OCI images for urunc runtime."
	app.Flags = []cli.Flag{
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
			Aliases:  []string{"a"},
			Usage:    "The image's architecture.",
			Required: true,
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
		},
	}
	app.Action = func(ctx *cli.Context) (err error) {
		defer func() {
			if err != nil {
				handle_err(err)
			}
		}()

		// We need to make sure ctr and docker are available
		ensureDependencies()

		unikernel := ctx.String("unikernel")
		image := ctx.String("image")
		if !strings.Contains(image, ":") {
			image = image + ":latest"
		}

		uType := ctx.String("type")

		extra := ctx.String("extra")

		cmdline := ctx.String("cmdline")

		// We create a temporary dir to store artifacts
		tempPath := "/tmp/bima" + randomTemp()
		err = os.Mkdir(tempPath, os.ModePerm)
		if err != nil {
			return err
		}
		// make sure we only have one trailing /
		tempPath = tempPath + "/"
		tempPath = strings.ReplaceAll(tempPath, "//", "/")

		// Create urunc.json config file with encoded values
		config := &config{uType, cmdline, "/unikernel/" + unikernel}
		config.encode()
		file, _ := json.MarshalIndent(config, " ", "")

		// Save it to our temporary dir
		err = ioutil.WriteFile(tempPath+"urunc.json", file, 0644)
		if err != nil {
			return err
		}

		// Check that the unikernel file exists and is not a directroy
		isFile, err := fileExists(unikernel)
		if err != nil {
			return err
		}
		if !isFile {
			return errors.New(unikernel + " is a directory")
		}

		// find the name of the unikernel file
		unikernelParts := strings.Split(unikernel, "/")
		unikernelName := unikernelParts[len(unikernelParts)-1]

		// copy the unikernel file to temporary dir
		err = copyFile(unikernel, tempPath+unikernelName)
		if err != nil {
			return err
		}

		// check if extrafile exists
		if extra != "" {
			isDir, err := dirExists(extra)
			if err != nil {
				return err
			}
			if isDir {
				tempExtraDir := tempPath + extra
				tempExtraDir = strings.ReplaceAll(tempExtraDir, "//", "/")
				err = cp.Copy(extra, tempExtraDir)
				if err != nil {
					return err
				}
				// extra = tempExtraDir
			}

			isFile, err := fileExists(extra)
			if err != nil {
				return err
			}
			if isFile {
				tempExtraDir := tempPath + extra
				tempExtraDir = strings.ReplaceAll(tempExtraDir, "//", "/")

				err = copyFile(extra, tempExtraDir)
				if err != nil {
					return err
				}
				// extra = tempExtraDir
			}
		}

		// create dockerfile
		docker := dockerdata{
			unikernel: unikernel,
			extrafile: extra,
			utype:     uType,
			cmdline:   cmdline,
		}
		err = docker.createFile(tempPath)
		if err != nil {
			return err
		}

		// run docker build
		execCmd := "sudo docker build -t " + image + " -f " + tempPath + "Dockerfile " + strings.TrimSuffix(tempPath, "/")
		err = executeFromString(execCmd)
		if err != nil {
			return err
		}

		// docker export
		execCmd = "sudo docker save -o" + tempPath + "img.tar " + image
		err = executeFromString(execCmd)
		if err != nil {
			return err
		}

		// importToCtr(tarPath)
		execCmd = "sudo ctr images import " + tempPath + "img.tar"
		err = executeFromString(execCmd)
		if err != nil {
			return err
		}

		// clean up
		err = os.RemoveAll(tempPath)
		if err != nil {
			return (err)
		}
		return nil
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
