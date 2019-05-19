package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"strconv"
)

func copy(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	if err := os.MkdirAll(filepath.Dir(dst), os.ModePerm); err != nil {
		return 0, err
	}

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}

func createFakeImage(fakeName string, classID int) error {
	source := filepath.Join(os.Getenv("GOPATH"), "src", "server", "testAux", "standardleaves", strconv.Itoa(classID+1)+".jpg")
	destination := filepath.Join(os.Getenv("GOPATH"), "data_store", fakeName+".jpg")

	_, err := copy(source, destination)
	return err
}

func TestGenerateOutputFile(t *testing.T) {
	fakeName := "test"

	if err := createFakeImage(fakeName, 0); err != nil {
		t.Errorf("failed to create fake image: %v", err)
	}

	if _, err := executeNNQuery(fakeName); err != nil {
		t.Errorf("failed to execute NN Query for '%v' err: %v", fakeName, err)
	}
}
