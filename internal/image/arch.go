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
	"fmt"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/partial"
)

// ArchOperation holds the information of the target architecture
type ArchOperation struct {
	Arch string
}

// newCopyOperation creates a new label operation
// based on the provided architecture
func newArchOperation(architecture string) (ArchOperation, error) {
	// TODO: Add verification based on valid architectures
	return ArchOperation{
		Arch: architecture,
	}, nil

}

func (o ArchOperation) Line() string {
	return o.Arch
}

func (o ArchOperation) Info() string {
	return fmt.Sprintf("Setting image architecture to: %q", o.Arch)
}

func (o ArchOperation) Type() string {
	return "ARCH"
}

func (o ArchOperation) UpdateImage(image v1.Image) (v1.Image, error) {
	ociConfigFile, err := partial.ConfigFile(image)
	if err != nil {
		return image, err
	}
	ociConfigFile.Architecture = o.Arch
	newImage, err := mutate.ConfigFile(image, ociConfigFile)
	if err != nil {
		return image, err
	}
	return newImage, nil
}
