package image

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/go-containerregistry/pkg/crane"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/partial"
	"github.com/nubificus/bima/internal/utils"
)

func baseImage() (*v1.Image, error) {
	newImage := empty.Image
	ociConfigFile, err := partial.ConfigFile(newImage)
	if err != nil {
		return nil, err
	}
	ociConfigFile.OS = "linux"
	newImage, err = mutate.ConfigFile(newImage, ociConfigFile)
	if err != nil {
		return nil, err
	}
	return &newImage, nil
}

type BimaImage struct {
	Image   *v1.Image
	labels  []LabelOperation
	archSet bool
}

func NewBimaImage() (*BimaImage, error) {
	img, err := baseImage()
	if err != nil {
		return nil, err
	}
	return &BimaImage{
		Image: img,
	}, nil
}

func (i *BimaImage) ApplyOperation(operation BimaOperation) error {
	img := *i.Image
	newImg, err := operation.UpdateImage(img)
	if err != nil {
		return err
	}
	i.Image = &newImg
	// persist all labels
	if operation.Type() == "LABEL" {
		i.labels = append(i.labels, operation.(LabelOperation))
	}
	// set architecture flag
	if operation.Type() == "ARCH" {
		i.archSet = true
	}
	return nil
}

func (i *BimaImage) getLabelKeys() []string {
	labels := []string{}
	for _, label := range i.labels {
		labels = append(labels, label.Key)
	}
	return labels
}

func (i *BimaImage) getLabelMap() map[string]string {
	labels := make(map[string]string)
	for _, label := range i.labels {
		labels[label.Key] = label.Value
	}
	return labels
}

func (i *BimaImage) imageType() string {
	labels := i.getLabelKeys()
	hasUnikernel := false
	hasIoT := false
	for _, label := range labels {
		if strings.Contains(label, "unikernel") {
			hasUnikernel = true
		}
		if strings.Contains(label, "iot") {
			hasIoT = true
		}
	}
	if hasUnikernel && !hasIoT {
		return "unikernel"
	}
	if hasIoT && !hasUnikernel {
		return "iot"
	}
	return "unknown"
}

func (i *BimaImage) Validate() error {
	imgType := i.imageType()
	if imgType == "unknown" {
		return fmt.Errorf("not a valid bima image")
	}
	if imgType == "unikernel" {
		return i.validateUnikernel()
	}
	if imgType == "iot" {
		return i.validateIoT()
	}
	return nil
}

func (i *BimaImage) validateUnikernel() error {
	missingLabels := []string{}
	requiredLabels := RequiredUnikernelAnnotations()
	currentLabels := i.getLabelMap()
	for _, label := range requiredLabels {
		_, ok := currentLabels[label]
		if !ok {
			missingLabels = append(missingLabels, label)
		}
	}
	if len(missingLabels) == 0 {
		return nil
	}
	return fmt.Errorf("ERR: invalid bima unikernel image - missing labels %v", missingLabels)
}
func (i *BimaImage) validateIoT() error {
	// todo
	return nil
}

func (i *BimaImage) AddCmd() error {
	imgType := i.imageType()
	if imgType == "unknown" {
		return fmt.Errorf("not a valid bima image")
	}
	if imgType == "unikernel" {
		return i.addUnikernelCmd()
	}
	if imgType == "iot" {
		return i.addIoTCmd()
	}
	return nil
}

func (i *BimaImage) addUnikernelCmd() error {
	cmd, ok := i.getLabelMap()[cmdAnnotation()]
	if !ok {
		return fmt.Errorf("invalid bima unikernel image - %q missing", cmdAnnotation())
	}
	img := *i.Image

	cfg, err := img.ConfigFile()
	if err != nil {
		return err
	}
	cfg = cfg.DeepCopy()
	decodedCmd, err := utils.Base64Decode(cmd)
	if err != nil {
		return err
	}
	cfg.Config.Cmd = strings.Split(decodedCmd, " ")
	img, err = mutate.Config(img, cfg.Config)
	if err != nil {
		return err
	}
	i.Image = &img
	return nil
}

func (i *BimaImage) addIoTCmd() error {
	// todo
	return nil
}

func (i *BimaImage) AddUruncJSON() error {
	imgType := i.imageType()
	if imgType == "unknown" {
		return fmt.Errorf("not a valid bima image")
	}
	if imgType == "unikernel" {
		return i.addUnikernelJSON()
	}
	if imgType == "iot" {
		return i.addIoTJSON()
	}
	return nil
}

func (i *BimaImage) addUnikernelJSON() error {
	img := *i.Image
	annotations := i.getLabelKeys()
	uruncMap := make(map[string]string)
	currentAnnotationMap := i.getLabelMap()
	for _, key := range annotations {
		uruncMap[key] = currentAnnotationMap[key]
	}
	byteObj, err := json.Marshal(uruncMap)
	if err != nil {
		return err
	}
	layerMap := map[string][]byte{"/urunc.json": byteObj}
	layer, err := crane.Layer(layerMap)
	if err != nil {
		return err
	}
	newImg, err := mutate.AppendLayers(img, layer)
	if err != nil {
		return err
	}
	i.Image = &newImg
	return nil
}

func (i *BimaImage) addIoTJSON() error {
	return nil
}

func (i *BimaImage) EnsureArchSet() error {
	if i.archSet {
		return nil
	}
	return fmt.Errorf("invalid Containerfile - target ARCH not set")
}
