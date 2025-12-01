package parser

import (
	"testing"
	"os"
	"strings"
	"bufio"
	"io"
)

func TestParseJSONValid(t *testing.T) {
	dirPath := "./test"

	files, err := os.ReadDir(dirPath)
	if err != nil {
		t.Errorf("Failed to read directory: %v", dirPath)
	}

	for _, file := range files {
		if !file.IsDir() && strings.Contains(file.Name(), "pass") {
			json, err := readFile(dirPath + "/" + file.Name())
			if err != nil {
				t.Errorf("Failed to read file: %v", file.Name())
			}
			
			isValidJSON, err := IsValidJSON(json)
			if !isValidJSON || err != nil {
				t.Errorf(`%s -> ParseJSON() = %t, %v, want match for %t, nil`, file.Name(), isValidJSON, err, true)
			}
		}
	}
}

func TestParseJSONInvalid(t *testing.T) {
	dirPath := "./test"

	files, err := os.ReadDir(dirPath)
	if err != nil {
		t.Errorf("Failed to read directory: %v", dirPath)
	}

	for _, file := range files {
		if !file.IsDir() && strings.Contains(file.Name(), "fail") {
			json, err := readFile(dirPath + "/" + file.Name())
			if err != nil {
				t.Errorf("Failed to read file: %v", file.Name())
			}
			
			isValidJSON, err := IsValidJSON(json)
			if isValidJSON || err == nil {
				t.Errorf(`%s -> ParseJSON() = %t, %v, want match for %t, some error`, file.Name(), isValidJSON, err, false)
			}
		}
	}
}

func readFile(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return "", err
	}

	data := make([]byte, stat.Size())
	_, err = bufio.NewReader(file).Read(data)
	if err != nil && err != io.EOF {
		return "", err
	}

	return string(data), nil
}
