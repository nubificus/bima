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
	"strings"

	v1 "github.com/google/go-containerregistry/pkg/v1"
)

// supportedOperations returns a list of all supported operations
func supportedOperations() [5]string {
	return [5]string{"FROM", "COPY", "LABEL", "NOOP"}
}

// InstructionLine represents a single line from the Containerfile
type InstructionLine string

func NewInstructionLine(line string) InstructionLine {
	line = strings.TrimSpace(line)
	first := line[0]
	if first == '#' {
		line = "NOOP " + line
	}
	return InstructionLine(line)
}

// Operation returns the operation defined on the instructio line (1st word of the line)
func (i InstructionLine) operation() string {
	return strings.Split(string(i), " ")[0]
}

// isSupported checks if the operation defined in the instruction line is supported by bima.
func (i InstructionLine) isSupported() bool {
	op := i.operation()
	supportedOps := supportedOperations()
	for _, thisOp := range supportedOps {
		if thisOp == op {
			return true
		}
	}
	return false
}

// ToBimaOperation creates a new BimaOperation based on the content of a single instruction line.
func (i InstructionLine) ToBimaOperation() (BimaOperation, error) {
	if !i.isSupported() {
		return nil, fmt.Errorf("operation %q is not supported", i.operation())
	}
	op := i.operation()
	switch op {
	case "FROM", "NOOP":
		return nil, nil
	case "COPY":
		return newCopyOperation(i)
	case "LABEL":
		return newLabelOperation(i)
	default:
		return nil, fmt.Errorf("ERR: Unsupported operation %q", op)
	}
}

type BimaOperation interface {
	Line() string
	Info() string
	Type() string
	UpdateImage(v1.Image) (v1.Image, error)
}
