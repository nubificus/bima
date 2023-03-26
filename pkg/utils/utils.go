package utils

import (
	"debug/elf"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
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

func UnikernelArch(name string) (string, error) {
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
