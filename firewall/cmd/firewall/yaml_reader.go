// File to implement YAMLReader interface defined in RulesExecutorYaml component
package main

import (
	"fmt"
	"io"
	"os"
)

type YAMLFileReader struct {
	src string
}

func NewYAMLFileReader(src string) *YAMLFileReader {
	return &YAMLFileReader{src}
}

func (r *YAMLFileReader) Readall() ([]byte, error) {
	confPath := r.src
	f, err := os.Open(confPath)
	if err != nil {
		return []byte{}, fmt.Errorf("cannot open file: %w", err)
	}

	content, err := io.ReadAll(f)

	if err != nil {
		return []byte{}, fmt.Errorf("file reading failed: %w", err)
	}

	return content, nil
}
