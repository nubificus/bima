package utils

import (
	"bufio"
	"encoding/base64"
	"os"
)

// FileExists checks if a file exists and is indeed a file.
// Returns true if the file exists and is a file,
// false if the file does not exist or is a directory,
// and an error if an error occurs while checking.
func FileExists(filename string) (bool, error) {
	info, err := os.Stat(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil // File does not exist
		}
		return false, err // Error occurred while checking
	}

	if info != nil && info.IsDir() {
		return false, nil // Path is a directory
	}

	return true, nil // File exists and is a file
}

// DirExists checks if a directory exists and is indeed a directory.
// Returns true if the directory exists and is a directory,
// false if the directory does not exist or is a file,
// and an error if an error occurs while checking.
func DirExists(dirname string) (bool, error) {
	info, err := os.Stat(dirname)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil // Directory does not exist
		}
		return false, err // Error occurred while checking
	}

	if info != nil && !info.IsDir() {
		return false, nil // Path is a file
	}

	return true, nil // Directory exists and is a directory
}

// SplitFileToLines reads a file and splits its contents into individual lines, excluding empty lines.
// It takes the file path as input and returns a slice of strings representing each non-empty line,
// along with an error if an error occurs while reading the file.
func SplitFileToLines(file string) ([]string, error) {
	readFile, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer readFile.Close()
	var lines []string
	fileScanner := bufio.NewScanner(readFile)

	for fileScanner.Scan() {
		if fileScanner.Text() != "" {
			lines = append(lines, fileScanner.Text())
		}
	}
	return lines, nil
}

// Base64Encode encodes the given string data to Base64 format.
// It takes a string as input and returns the Base64-encoded representation of the input data.
func Base64Encode(data string) string {
	encoded := base64.StdEncoding.EncodeToString([]byte(data))
	return encoded
}

// Base64Decode decodes the given Base64-encoded string to the original data.
// It takes a Base64-encoded string as input and returns the decoded data as a string.
func Base64Decode(encodedData string) (string, error) {
	decodedBytes, err := base64.StdEncoding.DecodeString(encodedData)
	if err != nil {
		return "", err
	}

	return string(decodedBytes), nil
}
