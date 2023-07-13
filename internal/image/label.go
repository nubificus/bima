package image

import (
	"fmt"
	"strings"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/nubificus/bima/internal/utils"
)

// LabelOperation hols the information needed
// to create a new layer with an annotation.
type LabelOperation struct {
	Key   string
	Value string
	line  string
}

// newLabelOperation creates a new label operation
// based on the provided instruction line.
func newLabelOperation(instructionLine InstructionLine) (LabelOperation, error) {
	arg := strings.ReplaceAll(string(instructionLine), "LABEL", "")
	arg = strings.TrimSpace(arg)
	parts := strings.Split(string(arg), "=")
	if len(parts) != 2 {
		return LabelOperation{}, fmt.Errorf("invalid LABEL format: %q", instructionLine)
	}
	key := strings.Trim(parts[0], "\"")
	val := strings.Trim(parts[1], "\"")
	return LabelOperation{
		Key:   key,
		Value: utils.Base64Encode(val),
		line:  string(instructionLine),
	}, nil
}

func (o LabelOperation) Line() string {
	return o.line
}

func (o LabelOperation) Info() string {
	return fmt.Sprintf("Performing instruction: %q\nSetting label %q to %q", o.line, o.Key, o.Value)
}

func (o LabelOperation) Type() string {
	return "LABEL"
}

func (o LabelOperation) UpdateImage(image v1.Image) (v1.Image, error) {
	annotations := make(map[string]string)
	annotations[o.Key] = o.Value
	newImage := mutate.Annotations(image, annotations).(v1.Image)
	return newImage, nil
}
