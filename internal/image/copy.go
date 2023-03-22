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
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-containerregistry/pkg/crane"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	l "github.com/nubificus/bima/internal/log"
	"github.com/nubificus/bima/internal/utils"
)

var log = l.Logger()

// CopyOperation hols the information needed
// to create a new layer with a copied file or directory.
type CopyOperation struct {
	Source      string
	Destination string
	line        string
}

// newCopyOperation creates a new copy operation
// based on the provided instruction line.
func newCopyOperation(instructionLine InstructionLine) (CopyOperation, error) {
	parts := strings.Split(string(instructionLine), " ")
	if len(parts) != 3 {
		return CopyOperation{}, fmt.Errorf("invalid COPY format: %q", instructionLine)
	}
	source, err := filepath.Abs(parts[1])
	if err != nil {
		return CopyOperation{}, err
	}
	dest, err := filepath.Abs(parts[2])
	if err != nil {
		return CopyOperation{}, err
	}
	return CopyOperation{
		Source:      source,
		Destination: dest,
		line:        string(instructionLine),
	}, nil
}

func (o CopyOperation) Info() string {
	return fmt.Sprintf("Performing instruction: %q\nCopying %q to %q", o.line, o.Source, o.Destination)
}

func (o CopyOperation) Line() string {
	return o.line
}

func (o CopyOperation) Type() string {
	return "COPY"
}

func (o CopyOperation) UpdateImage(image v1.Image) (v1.Image, error) {
	log.Debugf("Checking path: %q", o.Source)

	fileMap := make(map[string][]byte)
	transformedMap := make(map[string][]byte)
	exists, err := utils.DirExists(o.Source)
	if err != nil {
		return nil, err
	}
	if exists {
		err := scanDirectory(o.Source, fileMap)
		if err != nil {
			return nil, err
		}
	} else {
		fileContent, err := os.ReadFile(o.Source)
		if err != nil {
			return nil, err
		}
		fileMap[o.Source] = fileContent
	}
	log.Debugf("Found %v files in %q", len(fileMap), o.Source)
	if len(fileMap) == 0 {
		return nil, fmt.Errorf("%q does not exist or is empty", o.Source)
	}
	if len(fileMap) == 1 {
		for filePath, fileContent := range fileMap {
			if strings.HasSuffix(o.Destination, "/") {
				fileName := filepath.Base(filePath)
				newPath := filepath.Join(o.Destination, fileName)
				log.Tracef("Transformed %q to %q", filePath, newPath)

				transformedMap[newPath] = fileContent
			} else {
				transformedMap[o.Destination] = fileContent
				log.Tracef("Transformed %q to %q", filePath, o.Destination)

			}
		}
	} else {
		for filePath, fileContent := range fileMap {
			newPath := strings.Replace(filePath, o.Source, o.Destination, 1)
			transformedMap[newPath] = fileContent
			log.Tracef("Transformed %q to %q", filePath, newPath)
		}
	}
	layer, err := crane.Layer(transformedMap)
	if err != nil {
		return nil, err
	}
	newImage, err := mutate.AppendLayers(image, layer)
	if err != nil {
		return nil, err
	}
	return newImage, nil
}

func scanDirectory(dirPath string, fileMap map[string][]byte) error {
	files, err := os.ReadDir(dirPath)
	if err != nil {
		return err
	}
	for _, file := range files {
		filePath := filepath.Join(dirPath, file.Name())

		if file.IsDir() {
			err := scanDirectory(filePath, fileMap)
			if err != nil {
				return err
			}
		} else {
			fileContent, err := os.ReadFile(filePath)
			if err != nil {
				return err
			}
			fileMap[filePath] = fileContent
		}
	}
	return nil
}
