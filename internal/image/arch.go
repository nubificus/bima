package image

import (
	"fmt"
	"strings"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/partial"
)

// ArchOperation holds the information of the target architecture
type ArchOperation struct {
	Arch string
	line string
}

// newCopyOperation creates a new label operation
// based on the provided instruction line.
func newArchOperation(instructionLine InstructionLine) (ArchOperation, error) {
	// TODO: Add verification based on valid architectures
	arg := strings.ReplaceAll(string(instructionLine), "ARCH", "")
	arg = strings.TrimSpace(arg)
	return ArchOperation{
		Arch: arg,
		line: string(instructionLine),
	}, nil

}

func (o ArchOperation) Line() string {
	return o.line
}

func (o ArchOperation) Info() string {
	return fmt.Sprintf("Performing instruction: %q\nSetting target ARCH to %q", o.line, o.Arch)
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
