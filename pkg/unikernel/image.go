package unikernel

import (
	"io/ioutil"
	"strings"

	"github.com/nubificus/bima/pkg/utils"

	"fmt"
	"os"
	"path/filepath"

	"github.com/google/go-containerregistry/pkg/crane"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
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

type UnikernelImage struct {
	Image v1.Image
}

func NewUnikernelImage() *UnikernelImage {
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
