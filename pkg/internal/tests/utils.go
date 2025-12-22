package tests

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

func getDirectory() (string, error) {
	_, filePath, _, _ := runtime.Caller(0) //nolint:dogsled
	testDir := filepath.Dir(filePath)
	_, err := os.Stat(testDir)
	if err != nil {
		return "", err
	}
	return testDir, err
}

func getGrantScript() (string, error) {
	dir, err := getDirectory()
	if err != nil {
		return "", fmt.Errorf("failed current directory find: %w", err)
	}
	return filepath.Join(dir, "fixtures", "0-grant.sql"), nil
}

func GetFilePathWithError(name string) (string, error) {
	dir, err := getDirectory()
	if err != nil {
		return "", fmt.Errorf("failed current directory find: %w", err)
	}
	return path.Join(dir, "fixtures", name), nil
}

func GetFilePath(t *testing.T, name string) string {
	dir, err := GetFilePathWithError(name)
	require.NoError(t, err)
	return dir
}

func GetFileWithError(name string) ([]byte, error) {
	dir, err := getDirectory()
	if err != nil {
		return nil, fmt.Errorf("failed current directory find: %w", err)
	}
	filename := path.Join(dir, "fixtures", name)
	f, err := os.Open(filename) // nolint
	if err != nil {
		return nil, err
	}
	defer f.Close() // nolint
	return io.ReadAll(f)
}

func GetFile(t *testing.T, name string) []byte {
	buf, err := GetFileWithError(name)
	require.NoError(t, err)
	return buf
}
