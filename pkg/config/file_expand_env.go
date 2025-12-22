package config

import (
	"bytes"
	"io"
	"os"
)

func NewExpandEnv(name string) (io.Reader, error) {
	f, err := os.Open(name) // nolint: gosec
	if err != nil {
		return nil, err
	}
	defer f.Close() // nolint: errcheck
	return NewExpandEnvWithReader(f)
}

func NewExpandEnvWithReader(f io.Reader) (io.Reader, error) {
	buf, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	return bytes.NewBufferString(os.ExpandEnv(string(buf))), nil
}
