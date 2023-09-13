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
