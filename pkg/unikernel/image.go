package unikernel

import (
	"encoding/json"
	"io/ioutil"
	"strings"

	"github.com/nubificus/bima/pkg/utils"

	"fmt"
	"os"
	"path/filepath"

	"github.com/google/go-containerregistry/pkg/crane"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/partial"
)

type ImageType uint8

const (
	Unikernel ImageType = iota
	IOT
	// we can add more supported formats here
)

func (s ImageType) String() string {
	switch s {
	case Unikernel:
		return "unikernel"
	case IOT:
		return "iot"
	}
	return "unknown"
}

// The configuration used by bima to build the image
type UnikernelImageConfig struct {
	Name      string   // the container image name
	Type      string   // the type of the unikernel (eg hvt, qemu)
	Unikernel string   // the unikernel binary
	Extra     []string // any extra file or directory required by the unikernel
	Arch      string   // the CPU architecture of the unikernel binary
	CmdLine   string   // the cmdline for the unikernel
}

// The unikernel configuration required by urunc
type UnikernelConfig struct {
	VmmType         string `json:"type"`
	UnikernelCmd    string `json:"cmdline,omitempty"`
	UnikernelBinary string `json:"binary"`
}

func (c *UnikernelConfig) encode() {
	c.VmmType = utils.Base64Encode(c.VmmType)
	c.UnikernelCmd = utils.Base64Encode(c.UnikernelCmd)
	c.UnikernelBinary = utils.Base64Encode(c.UnikernelBinary)
}

type UnikernelImage struct {
	Image v1.Image
}

func CreateImage(config UnikernelImageConfig) (v1.Image, error) {
	// create an empty base image
	newImage := empty.Image

	// add arch/os
	ociConfigFile, err := partial.ConfigFile(newImage)
	if err != nil {
		return nil, err
	}
	ociConfigFile.Architecture = config.Arch
	ociConfigFile.OS = "linux"
	newImage, err = mutate.ConfigFile(newImage, ociConfigFile)
	if err != nil {
		return nil, err
	}

	// create a temporary dir to store urunc.json
	tempDir, err := utils.CreateRandomDirectory()
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tempDir)

	// create the urunc.json config file inside the temporary directory
	unikernelConfig := &UnikernelConfig{
		VmmType:         config.Type,
		UnikernelCmd:    config.CmdLine,
		UnikernelBinary: filepath.Base(config.Unikernel),
	}
	unikernelConfig.encode()
	file, err := json.MarshalIndent(unikernelConfig, " ", "")
	if err != nil {
		return nil, err
	}
	err = ioutil.WriteFile(tempDir+"/urunc.json", file, 0644)
	if err != nil {
		return nil, err
	}

	// populate filesystem layer-map for 1st layer
	baseLayerMap := make(map[string][]byte)

	// add urunc.json to layer map
	configData, err := os.ReadFile(tempDir + "/urunc.json")
	if err != nil {
		return nil, err
	}
	baseLayerMap["/urunc.json"] = configData

	// add unikernel binary to layer map
	unikernelData, err := os.ReadFile(config.Unikernel)
	if err != nil {
		return nil, err
	}
	unikernelName := "/unikernel/" + filepath.Base(config.Unikernel)
	baseLayerMap[unikernelName] = unikernelData

	baseLayer, err := crane.Layer(baseLayerMap)
	if err != nil {
		return nil, err
	}
	newImage, err = mutate.AppendLayers(newImage, baseLayer)
	if err != nil {
		return nil, err
	}

	// create layers for extra file(s)
	var extraLayers []v1.Layer
	for _, val := range config.Extra {
		temp, err := layerFromPath(val, "/extra/")
		if err != nil {
			return nil, err
		}
		extraLayers = append(extraLayers, temp)
	}
	for _, layer := range extraLayers {
		newImage, err = mutate.AppendLayers(newImage, layer)
		if err != nil {
			return nil, err
		}
	}

	// save urunc specific annotations
	encodedAnnotations := make(map[string]string)
	encodedAnnotations["com.urunc.unikernel.cmdline"] = unikernelConfig.UnikernelCmd
	encodedAnnotations["com.urunc.unikernel.type"] = unikernelConfig.VmmType
	encodedAnnotations["com.urunc.unikernel.binary"] = unikernelConfig.UnikernelBinary
	newImage = mutate.Annotations(newImage, encodedAnnotations).(v1.Image)

	return newImage, nil
}

func (u *UnikernelImage) AddAnnotations(annotations map[string]string) {
	encodedAnnotations := make(map[string]string)
	for key, value := range annotations {
		encodedAnnotations[key] = utils.Base64Encode(value)
	}
	u.Image = mutate.Annotations(u.Image, encodedAnnotations).(v1.Image)
}

func layerFromFile(source string, target string) (v1.Layer, error) {
	fileData, err := os.ReadFile(source)
	if err != nil {
		return nil, err
	}
	fileName := filepath.Base(source)
	fileName = "/" + target + "/" + fileName
	fileName = strings.ReplaceAll(fileName, "//", "/")
	layerMap := make(map[string][]byte)
	layerMap[fileName] = fileData
	layer, err := crane.Layer(layerMap)
	if err != nil {
		return nil, err
	}
	return layer, nil
}
func layerFromDir(source string, target string) (v1.Layer, error) {
	var files []string
	err := filepath.Walk(source, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			files = append(files, p)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	layerMap := make(map[string][]byte)
	for _, file := range files {
		bytes, err := ioutil.ReadFile(file)
		if err != nil {
			return nil, err
		}
		fileName, err := filepath.Rel(source, file)
		if err != nil {
			return nil, err
		}
		fileName = "/" + target + "/" + fileName
		fileName = strings.ReplaceAll(fileName, "//", "/")
		layerMap[fileName] = bytes
	}
	layer, err := crane.Layer(layerMap)
	if err != nil {
		return nil, err
	}
	return layer, nil
}
func layerFromPath(source string, target string) (v1.Layer, error) {
	absPath, err := filepath.Abs(source)
	if err != nil {
		return nil, err
	}
	if err := utils.FileExists(absPath); err == nil {
		return layerFromFile(absPath, target)
	}
	if err := utils.DirExists(absPath); err == nil {
		return layerFromDir(absPath, target)
	}
	return nil, fmt.Errorf("unable to read %s", source)
}
