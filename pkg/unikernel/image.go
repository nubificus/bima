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
	Name      string // the container image name
	Type      string // the type of the unikernel (eg hvt, qemu)
	Unikernel string // the unikernel binary
	Extra     string // any extra file or directory required by the unikernel
	Arch      string // the CPU architecture of the unikernel binary
	CmdLine   string // the cmdline for the unikernel
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
	// type ConfigFile struct {
	// 	Architecture  string    `json:"architecture"`
	// 	Author        string    `json:"author,omitempty"`
	// 	Container     string    `json:"container,omitempty"`
	// 	Created       Time      `json:"created,omitempty"`
	// 	DockerVersion string    `json:"docker_version,omitempty"`
	// 	History       []History `json:"history,omitempty"`
	// 	OS            string    `json:"os"`
	// 	RootFS        RootFS    `json:"rootfs"`
	// 	Config        Config    `json:"config"`
	// 	OSVersion     string    `json:"os.version,omitempty"`
	// 	Variant       string    `json:"variant,omitempty"`
	// 	OSFeatures    []string  `json:"os.features,omitempty"`
	// }

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
	// create layer for extra file(s)
	extraLayer, err := layerFromPath(config.Extra, "/extra/")
	if err != nil {
		return nil, err
	}
	newImage, err = mutate.AppendLayers(newImage, extraLayer)
	if err != nil {
		return nil, err
	}

	// save urunc specific annotations
	encodedAnnotations := make(map[string]string)
	encodedAnnotations["com.urunc.unikernel.cmdline"] = unikernelConfig.UnikernelCmd
	encodedAnnotations["com.urunc.unikernel.type"] = unikernelConfig.VmmType
	encodedAnnotations["com.urunc.unikernel.binary"] = unikernelConfig.UnikernelBinary
	newImage = mutate.Annotations(newImage, encodedAnnotations).(v1.Image)

	return newImage, nil
}

// save the image as tarbal
func Save(image v1.Image, path string, tag string) error {
	imageManifest, err := image.Manifest()
	if err != nil {
		return err
	}
	fmt.Println("imageManifest: ", imageManifest)
	fmt.Println("imageManifestSubject: ", imageManifest.Subject)
	return crane.Save(image, tag, path)
}

func NewUnikernelImage() *UnikernelImage {

	// "platform": {
	//     "architecture": "arm64",
	//     "os": "linux",
	//     "variant": "v8
	// }
	// newImage := empty.Image

	imageMap := make(map[string][]byte)
	image, err := crane.Image(imageMap)
	if err != nil {
		panic(err)
	}
	return &UnikernelImage{
		Image: image,
	}
}

// Adds the given file to the image's rootfs at "/unikernel/$unikernel"
func (u *UnikernelImage) AddUnikernelFile(unikernel string) error {
	unikernelLayer, err := layerFromFile(unikernel, "/unikernel/")
	if err != nil {
		return err
	}
	u.Image, err = mutate.AppendLayers(u.Image, unikernelLayer)
	if err != nil {
		return err
	}
	return nil
}

func (u *UnikernelImage) AddExtraFiles(extrafile string) error {
	extraLayer, err := layerFromPath(extrafile, "/extra/")
	if err != nil {
		return err
	}
	u.Image, err = mutate.AppendLayers(u.Image, extraLayer)
	if err != nil {
		return err
	}
	return nil
}

func (u *UnikernelImage) AddAnnotations(annotations map[string]string) {
	encodedAnnotations := make(map[string]string)
	for key, value := range annotations {
		encodedAnnotations[key] = utils.Base64Encode(value)
	}
	u.Image = mutate.Annotations(u.Image, encodedAnnotations).(v1.Image)
}

// This functions saves the docker image as an OCI combatible bundle
func (u *UnikernelImage) SaveOCI(path string) error {
	return crane.SaveOCI(u.Image, path)
}

// This functions saves the docker image as a tarbal
func (u *UnikernelImage) Save(path string, tag string) error {
	imageManifest, err := u.Image.Manifest()
	if err != nil {
		panic(err)
	}
	fmt.Println("imageManifest: ", imageManifest)
	fmt.Println("imageManifestSubject: ", imageManifest.Subject)
	return crane.Save(u.Image, tag, path)
}

func (u *UnikernelImage) Trouble() {
	imageManifest, err := u.Image.Manifest()
	if err != nil {
		panic(err)
	}

	platform := v1.Platform{
		Architecture: "amd64",
		OS:           "linux",
	}
	subject := v1.Descriptor{
		MediaType:   imageManifest.MediaType,
		Size:        imageManifest.Config.Size,
		Digest:      imageManifest.Config.Digest,
		Annotations: imageManifest.Config.Annotations,
		Platform:    &platform,
	}
	imageManifest.Subject = &subject
	// mutate.AppendManifests(u.Image, mutate.IndexAddendum{Add: imageManifest})
	// fmt.Println("Platform: ", imageManifest.Subject.Platform.Architecture)
	// fmt.Println("Platform: ", imageManifest.Subject.Platform.OS)
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
