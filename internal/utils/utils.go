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

package utils

import (
	"bufio"
	"encoding/binary"
	"encoding/base64"
	"os"
	"fmt"
	"debug/elf"
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

// UKLibraryInfoRecord holds the information parsed from the library info record
type UKLibraryInfoRecord struct {
	Type    uint16
	Data    []byte
}

// UKLibraryInfoHeader holds the information parsed from the library info header
type UKLibraryInfoHeader struct {
	Version uint16
	Records []UKLibraryInfoRecord
}

// parseUKLibInfo parses the data from the uklibinfo section
func parseUKLibInfo(data []byte) ([]UKLibraryInfoHeader, error) {
	var headers []UKLibraryInfoHeader
	offset := 0

	for offset < len(data) {
		if len(data[offset:]) < 6 {
			return nil, fmt.Errorf("incomplete header")
		}
		hdrLen := binary.LittleEndian.Uint32(data[offset:])
		hdrVersion := binary.LittleEndian.Uint16(data[offset+4:])
		offset += 6

		header := UKLibraryInfoHeader{Version: hdrVersion}

		recEnd := offset + int(hdrLen)-6
		for offset < recEnd {
			if len(data[offset:]) < 6 {
				return nil, fmt.Errorf("incomplete record")
			}
			recType := binary.LittleEndian.Uint16(data[offset:])
			recLen := binary.LittleEndian.Uint32(data[offset+2:])
			if recType != 0x0007 {
				offset += int(recLen)
				continue
			}
			offset += 6

			record := UKLibraryInfoRecord{
				Type: recType,
				Data: data[offset : offset+int(recLen)-6],
			}
			header.Records = append(header.Records, record)
			offset += int(recLen) - 6
		}

		headers = append(headers, header)
	}
	return headers, nil
}

// uklibinfoELFLoad loads the uklibinfo section from the ELF binary
func uklibinfoELFLoad(filePath string) ([]byte, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	e, err := elf.NewFile(f)
	if err != nil {
		return nil, err
	}

	section := e.Section(".uk_libinfo")
	if section == nil {
		return nil, fmt.Errorf("section .uk_libinfo not found")
	}
	return section.Data()
}

func GetUnikraftVersion(filePath string) ([]byte, error) {

	uklibinfo, err := uklibinfoELFLoad(filePath)
	if err != nil {
		fmt.Errorf("Failed to load uklibinfo data: %v", err)
		return nil, err
	}

	headers, err := parseUKLibInfo(uklibinfo)
	if err != nil {
		fmt.Errorf("Failed to parse uklibinfo data: %v", err)
		return nil, err
	}

	for _, header := range headers {
		for _, record := range header.Records {
			if record.Type == 0x0007 { // VERSION
	//			fmt.Printf("Type: %v\n", record.Type)
	//			fmt.Printf("Version: %s\n", string(record.Data))
				return record.Data, nil
			}
		}
	}
	return nil, err
}
