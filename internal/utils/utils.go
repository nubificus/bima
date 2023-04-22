package utils

import (
	"debug/elf"
	"encoding/base64"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/containerd/containerd/reference"
	v1 "github.com/google/go-containerregistry/pkg/v1"
)

// Returns the base64 encoding of data as string.
func Base64Encode(data string) string {
	encoded := base64.StdEncoding.EncodeToString([]byte(data))
	return encoded
}

// Checks if the file in the given path exists. If file exists and is not directory returns nil, else returns err.
func FileExists(path string) error {
	filePath, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}
	if fileInfo.IsDir() {
		return fmt.Errorf("file %s is a directory", filePath)
	}
	return nil
}

// Returns the absolute filepath and ensures it points to an existing file
func ValidAbsoluteFilePath(path string) (string, error) {
	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(absolutePath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("%s does not exist", absolutePath)
		}
		return "", err
	}
	if info.IsDir() {
		return "", fmt.Errorf("%s is a directory, not a file", absolutePath)
	}
	return absolutePath, nil
}

// Returns the absolute filepath and ensures it points to an existing file or directory
func ValidAbsolutePath(path string) (string, error) {
	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(absolutePath); os.IsNotExist(err) {
		return "", fmt.Errorf("%s does not exist", absolutePath)
	}
	return absolutePath, nil
}

// Checks if the directory in the given path exists. If directory exists and is not file returns nil, else returns err.
func DirExists(path string) error {
	filePath, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}
	if !fileInfo.IsDir() {
		return fmt.Errorf("file %s is a file", filePath)
	}
	return nil
}

func ValidImageName(name string) bool {
	_, err := reference.Parse(name)
	return err == nil
}

func GetBinaryArchitecture(name string) (string, error) {
	file, err := elf.Open(name)
	if err != nil {
		return "", err
	}
	defer file.Close()
	switch file.Machine {
	case elf.EM_AARCH64:
		return "arm64", nil
	case elf.EM_X86_64:
		return "amd64", nil
	default:
		return "", fmt.Errorf("unknown architecture")
	}
}

func CreateRandomDirectory() (string, error) {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	rand.Seed(time.Now().UnixNano())
	b := make([]rune, 10)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	directory := "/tmp/bima-" + string(b)
	err := os.Mkdir(directory, os.ModePerm)
	if err != nil {
		return "", nil
	}
	return strings.TrimSuffix(directory, "/"), nil
}

func SplitImageName(imageName string) (string, string) {
	parts := strings.Split(imageName, "/")
	nameWithTag := parts[len(parts)-1]
	nameParts := strings.Split(nameWithTag, ":")
	tag := ""
	if len(nameParts) > 1 {
		tag = nameParts[len(nameParts)-1]
	} else {
		tag = "latest"
	}

	name := nameParts[0]
	return name, tag
}

func ImportImage(image v1.Image, name string) error {
	return nil
}
