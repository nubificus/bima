package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/containerd/containerd/reference"
	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/nubificus/bima/internal/ctr"
	"github.com/nubificus/bima/internal/image"
	"github.com/nubificus/bima/internal/utils"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func bimaBuild(ctx *cli.Context) error {
	// Parse CLI flags
	buildContext := ctx.Args().First()
	namespace := ctx.String("namespace")
	address := ctx.String("address")
	snapshotter := ctx.String("snapshotter")
	output := ctx.String("output")
	tag := ctx.String("tag")
	tarOutput := ctx.Bool("tar")
	file := ctx.String("file")
	if tarOutput {
		output = "tar"
	}
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	log.Tracef("Started in %q", wd)
	log.Tracef("Got buildContext %q", buildContext)
	log.Tracef("Got namespace %q", namespace)
	log.Tracef("Got address %q", address)
	log.Tracef("Got snapshotter %q", snapshotter)
	log.Tracef("Got output %q", output)
	log.Tracef("Got tarOutput %v", tarOutput)
	log.Tracef("Got tag %q", tag)
	log.Tracef("Got file %q", file)

	// Verify tag
	spec, err := reference.Parse(tag)
	if err != nil {
		log.Fatalf("ERROR: invalid image tag - %q", err.Error())
	}
	log.Debugf("Got spec %q with locator %q and object %q", spec, spec.Locator, spec.Object)
	log.Infof("Creating image %q", spec)

	// Verify context directory exists
	buildContext, err = filepath.Abs(buildContext)
	if err != nil {
		log.Fatalf("ERROR: invalid context directory path - %q", err.Error())
	}
	exists, err := utils.DirExists(buildContext)
	if err != nil {
		log.Fatalf("ERROR: error checking directory %q - %v", buildContext, err.Error())
	}
	if !exists {
		log.Fatalf("ERROR: given context directory %q does not exist or is a file", buildContext)
	}
	log.Debugf("Got absolute path for context directory: %q", buildContext)

	// Verify given file path exists. If not, check for ./Dockerfile and ./Containerfile
	file, err = filepath.Abs(file)
	if err != nil {
		log.Fatalf("ERROR: invalid Containerfile path - %q", err.Error())
	}
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("ERROR: error getting current working directory - %q", err.Error())
	}
	containerfilePath := filepath.Join(cwd, "Containerfile")
	dockerfilePath := filepath.Join(cwd, "Dockerfile")

	possibleFiles := [3]string{file, containerfilePath, dockerfilePath}

	// Check all possible files to find the correct one
	found := false
	for _, f := range possibleFiles {
		exists, _ := utils.FileExists(f)
		if exists {
			file = f
			found = true
			break
		}
	}
	if !found {
		log.Fatal("ERROR: could not find given Containerfile")
	}
	log.Tracef("Got absolute path for Containerfile: %q", file)
	if file != ctx.String("file") {
		log.Infof("Could not find %q, will use %q", ctx.String("file"), file)
	}

	// verify given output is supported
	if output != "tar" && output != "ctr" {
		log.Fatal("ERROR: invalid output type")
	}

	// create image based on context and containerfile
	img, err := buildImage(buildContext, file)
	if err != nil {
		log.Fatal(err.Error())
	}
	log.Debugf("Built image %v", img)

	// chdir to previous workdir to properly save image
	err = os.Chdir(wd)
	if err != nil {
		return err
	}
	targetPathParts := strings.Split(tag, "/")
	targetPath := targetPathParts[len(targetPathParts)-1]
	targetPath, err = filepath.Abs(targetPath)
	if err != nil {
		return err
	}
	// save image to tarball
	err = crane.Save(*img.Image, tag, targetPath)
	if err != nil {
		return err
	}
	log.Debugf("Saved image at %v", targetPath)

	// Import to ctr if set
	if output == "ctr" {
		log.Debug("Importing to ctr")
		msg, err := ctr.ImportImage(targetPath, address, namespace, snapshotter)
		if err != nil {
			return err
		}
		log.Infof(msg)
		err = os.RemoveAll(targetPath)
		if err != nil {
			return err
		}
	}

	return nil
}

func getOperations(contextDir string, containerFile string) ([]image.BimaOperation, error) {
	// chdir to context directory
	err := os.Chdir(contextDir)
	if err != nil {
		log.Fatalf("ERROR: error changing directory - %q", err.Error())
	}
	log.Debugf("Changed directory to %q", contextDir)
	lines, err := utils.SplitFileToLines(containerFile)
	if err != nil {
		return nil, err
	}
	log.Tracef("Read following lines from Containerfile: %v", lines)
	operations := []image.BimaOperation{}
	for _, line := range lines {
		log.Tracef("Creating bima operation from line %q", line)
		instruction := image.NewInstructionLine(line)
		operation, err := instruction.ToBimaOperation()
		if err != nil {
			return nil, err
		}
		operations = append(operations, operation)
	}
	log.Debugf("Found %v operations in file %q", len(operations), containerFile)
	validOps := []image.BimaOperation{}
	for _, op := range operations {
		if op != nil {
			validOps = append(validOps, op)
		}
	}
	log.Infof("Found %v operations in file %q", len(validOps), containerFile)
	return validOps, nil
}

func buildImage(buildContext string, file string) (*image.BimaImage, error) {
	// Parse containerfile to find all operations
	operations, err := getOperations(buildContext, file)
	if err != nil {
		return nil, fmt.Errorf("ERROR: failed to convert Containerfile to bima operations - %q", err.Error())
	}
	if logrus.DebugLevel == log.GetLevel() {
		for _, op := range operations {
			log.Debugf("Got %q operation: %v", op.Type(), op)
		}
	}

	// create new bima image
	img, err := image.NewBimaImage()
	if err != nil {
		return nil, err
	}
	log.Debug("Created new empty image")
	for _, op := range operations {
		err := img.ApplyOperation(op)
		if err != nil {
			return nil, fmt.Errorf("ERROR: failed to add operation %v to image - %q", op, err.Error())
		}
		log.Debug(op.Info())
		log.Infof("Appending layer %q", op.Line())
	}

	// verify all mandatory labels are set
	err = img.Validate()
	if err != nil {
		return nil, err
	}

	// check if arch was defined, if not use the host's arch
	err = img.EnsureArchSet()
	if err != nil {
		return nil, err
	}

	// add cmd
	err = img.AddCmd()
	if err != nil {
		return nil, err
	}

	// add urunc.json file from labels
	err = img.AddUruncJSON()
	if err != nil {
		return nil, err
	}

	return img, nil
}
