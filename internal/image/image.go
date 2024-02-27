// Copyright 2023 Nubificus LTD.

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package image

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"debug/elf"
	"debug/pe"

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
	Image  *v1.Image
	labels []LabelOperation
	copies []CopyOperation
	arch   string
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
	} else if operation.Type() == "COPY" {
		i.copies = append(i.copies, operation.(CopyOperation))
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

func (i *BimaImage) extractIUnikernelArch() error {
	// first we need to find the value of annotation "com.urunc.unikernel.binary"
	targetKey := cmdAnnotation()
	targetVal := ""
	for _, val := range i.labels {
		if val.Key == targetKey {
			targetVal = val.Value
			break
		}
	}
	if targetVal == "" {
		return fmt.Errorf("unikernel annotation was not set")
	}
	targetVal, err := utils.Base64Decode(targetVal)
	if err != nil {
		return fmt.Errorf("failed to decode unikernel annotation value")

	}
	// next, we need to find the file name of the unikernel
	unikernelName := filepath.Base(targetVal)
	unikernelPath := ""

	// search COPY operations to find the local unikernel file
	for _, val := range i.copies {
		tmp := filepath.Base(val.Destination)
		if tmp == unikernelName {
			unikernelPath = val.Source
			break
		}
	}
	if unikernelPath == "" {
		return fmt.Errorf("unikernel defined by annotation was not copied in image rootfs")
	}

	elfFile, err := elf.Open(unikernelPath)
	if err == nil {
		defer elfFile.Close()
		switch elfFile.Machine {
		case elf.EM_ARM:
			i.arch = "arm64"
			return nil
		case elf.EM_AARCH64:
			i.arch = "arm64"
			return nil
		case elf.EM_X86_64:
			i.arch = "amd64"
			return nil
		case elf.EM_386:
			i.arch = "amd64"
			return nil
		default:
			return fmt.Errorf("unknown architecture")
		}
	}

	// We are not dealing with an elf binary. Maybe we can try PE
	peFile, err := pe.Open(unikernelPath)
	if err == nil {
		defer peFile.Close()
		switch peFile.FileHeader.Machine {
		case pe.IMAGE_FILE_MACHINE_AMD64:
			i.arch = "arm64"
			return nil
		case pe.IMAGE_FILE_MACHINE_ARM64:
			i.arch = "arm64"
			return nil
		default:
			return fmt.Errorf("unknown architecture")
		}
	}

	// We are not dealing with an elf or PE binary.
	// Let's check if it is Linux kernel ARM64 boot executable Image
	file, err := os.Open(unikernelPath)
	if err != nil {
		return err
	}
	defer file.Close()

	var dosheader [64]byte

	if _, err = file.ReadAt(dosheader[0:], 0); err != nil {
		return err
	}
	if dosheader[0] == 'M' && dosheader[1] == 'Z' {
		if dosheader[56] == 'A' && dosheader[57] == 'R' {
			if dosheader[58] == 'M' && dosheader[59] == 'd' {
				i.arch = "arm64"
				return nil
			}
		}
	}

	return err
}

func (i *BimaImage) SetArchitecture() error {
	err := i.extractIUnikernelArch()
	if err != nil {
		return err
	}
	newOp, err := newArchOperation(i.arch)
	if err != nil {
		return err
	}
	img := *i.Image
	newImg, err := newOp.UpdateImage(img)
	if err != nil {
		return err
	}
	i.Image = &newImg
	return nil
}
