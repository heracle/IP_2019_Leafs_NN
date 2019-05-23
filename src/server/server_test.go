package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/pkg/errors"

	"io/ioutil"
	"os/exec"

	infogen "lib"
)

const (
	maxImageSize = 100000000
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

func getTestImage(classID int) string {
	return filepath.Join(os.Getenv("GOPATH"), "data_base", "testAux", "standardleaves", strconv.Itoa(classID+1)+".jpg")
}

func createFakeImage(fakeName string, classID int) error {
	source := getTestImage(classID)
	destination := filepath.Join(os.Getenv("GOPATH"), "data_base", "queries", fakeName+".jpg")
	_, err := copy(source, destination)
	return err
}

func encodeInJSON(classID int) ([]byte, error) {
	f, err := os.Open(getTestImage(classID))
	if err != nil {
		return nil, errors.Wrapf(err, "could not open the image file")
	}
	imgBytes := make([]byte, maxImageSize)
	sizeOutput, err := f.Read(imgBytes)
	if err != nil {
		return nil, errors.Wrapf(err, "could not read the image file")
	}
	// We need just the first sizeOutput bytes.
	imgBytes = imgBytes[:sizeOutput]

	bodyObj := new(jsonClass)

	bodyObj.Photo = base64.StdEncoding.EncodeToString(imgBytes)

	jsonBytes, err := json.Marshal(bodyObj)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to Marshal the json bytes")
	}

	return jsonBytes, nil
}

func TestGenerateOutputFile(t *testing.T) {
	fakeName := "test"

	if err := createFakeImage(fakeName, 12); err != nil {
		t.Errorf("failed to create fake image: %v", err)
	}

	if _, err := executeNNQuery(fakeName); err != nil {
		t.Errorf("failed to execute NN Query for '%v' err: %v", fakeName, err)
	}
}

// func TestAllClasses(t *testing.T) {
// 	totalImgs := 32
// 	correctAnswer := 0
// 	fakeName := "test"

// 	for picIdx := 0; picIdx < totalImgs; picIdx ++ {
// 		if err := createFakeImage(fakeName, picIdx); err != nil {
// 			t.Errorf("failed to create fake image: %v", err)
// 		}

// 		prediction, err := executeNNQuery(fakeName)
// 		if err != nil {
// 			t.Errorf("failed to execute NN Query for '%v' err: %v", fakeName, err)
// 		}

// 		if prediction == picIdx {
// 			correctAnswer ++
// 			fmt.Printf("OK %d\n", picIdx)
// 		} else {
// 			fmt.Printf("NO %d\n", picIdx)
// 		}
// 	}
// 	fmt.Printf("correct answer at %d of %d", correctAnswer, totalImgs)

// 	if correctAnswer  < totalImgs * 3 / 4 {
// 		t.Error("prediction under 75%")
// 	}
// }

func TestBodyFormat(t *testing.T) {
	imgEncoded, err := encodeInJSON(0)
	if err != nil {
		t.Errorf("could not encode in JSON: %v", err)
	}

	if _, err := evaluateReceivedBody(imgEncoded); err != nil {
		t.Errorf("failed to execute evaluateReceivedBody; err: %v", err)
	}
}

func TestStartStopServer(t *testing.T) {
	testPort := "2029"

	cmdServer := exec.Command(filepath.Join(os.Getenv("GOPATH"), "bin", "server"))
	cmdServer.Env = append(os.Environ(),
		"PORT="+testPort,
	)
	cmdServer.Start()

	if err := cmdServer.Process.Kill(); err != nil {
		t.Errorf("failed to kill process: %v", err)
	}
}

func TestEntireServer(t *testing.T) {
	jsonPath := filepath.Join(os.Getenv("GOPATH"), "data_base", "jsons", "temp_json.json")
	testPort := "2029"

	cmdServer := exec.Command(filepath.Join(os.Getenv("GOPATH"), "bin", "server"))
	cmdServer.Env = append(os.Environ(),
		"PORT="+testPort,
	)
	cmdServer.Start()

	imgEncoded, err := encodeInJSON(21)
	if err != nil {
		t.Errorf("could not encode in JSON: %v", err)
	}

	if err := ioutil.WriteFile(jsonPath, imgEncoded, 0644); err != nil {
		t.Errorf("could not write a json file: %v", err)
	}

	cmdCurl := exec.Command("curl", "--request", "POST", "--data-binary", "@"+jsonPath, "localhost:"+testPort+"/post")
	// cmdCurl.Stdout = os.Stdout
	// cmdCurl.Stderr = os.Stderr

	output, err := cmdCurl.Output()
	if err != nil {
		t.Errorf("query to server failed, err: %v", err)
	}

	fmt.Printf("output = \n%v", string(output))

	// We need that the result to be unmarshable.
	var retJSON infogen.QueryInfo
	if err := json.Unmarshal(output, &retJSON); err != nil {
		t.Errorf("the output is not a legit JSON file")
	}

	if err := cmdServer.Process.Kill(); err != nil {
		t.Errorf("failed to kill process: %v", err)
	}
}
