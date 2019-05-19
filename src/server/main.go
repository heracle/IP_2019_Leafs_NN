/*
* Sample API with GET and POST endpoint.
* POST data is converted to string and saved in internal memory.
* GET endpoint returns all strings in an array.
 */
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"encoding/base64"
	"lib"

	"github.com/pkg/errors"
)

var (
	// port                = "2020"
	// flagPort          = flag.String("port", port, "Port to listen on")
	imageFormat       = ".jpg"
	descriptionFormat = ".txt"
	imagesStorePath   = filepath.Join(os.Getenv("GOPATH"), "data_store")
	randStringSize    = 5
	receiveURL        = "/receive/"
	// receiveCompleteURL  = port + receiveURL
	pythonAskForJobPath = filepath.Join(os.Getenv("GOPATH"), "src", "nn", "ask_for_job.py")
	pythonTrainPath     = filepath.Join(os.Getenv("GOPATH"), "src", "nn", "train.py")
	maxDescriptionSize  = 10000
)

var jobQueueMutex = &sync.Mutex{}
var jobQueue []string
var results []string

func getRandomID(n int) string {
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	result := make([]rune, n)
	for i := range result {
		result[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(result)
}

// GetHandler handles the index route
func GetHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("in get function")
	jsonBody, err := json.Marshal(results)
	if err != nil {
		http.Error(w, "Error converting results to json",
			http.StatusInternalServerError)
	}
	w.Write(jsonBody)
}

type jsonClass struct {
	Photo string `json:"photo"`
}

func executeNNQuery(randID string) (int, error) {
	cmd := exec.Command("python3", pythonAskForJobPath, randID)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return 0, errors.Wrapf(err, "Failed to execute %v for job %v.", pythonAskForJobPath, randID)
	}

	f, err := os.Open(filepath.Join(imagesStorePath, randID+descriptionFormat))
	if err != nil {
		return 0, errors.Wrapf(err, "Failed to open the file created after execution of %v for job %v.", pythonAskForJobPath, randID)
	}
	output := make([]byte, maxDescriptionSize)
	sizeOutput, err := f.Read(output)
	if err != nil {
		return 0, errors.Wrapf(err, "Failed to read the description file created after execution of %v for job %v.", pythonAskForJobPath, randID)
	}
	// We need just the first sizeOutput bytes.
	output = output[:sizeOutput]

	classID, err := strconv.Atoi(string(output))
	if err != nil {
		return 0, errors.Wrapf(err, "failed to get int from byte-string '%s'", output)
	}

	return classID, nil
}

// PostHandler converts post request body to string
func PostHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Received a request from %v\n", r.RemoteAddr)

	if r.Method != "POST" {
		fmt.Println("\t Error: request is not of type POST")
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}

	bodyP, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body",
			http.StatusInternalServerError)
	}

	var bodyObj jsonClass
	err = json.Unmarshal(bodyP, &bodyObj)
	if err != nil {
		http.Error(w, "Error while apply unmarshal on the received body to get the JSON",
			http.StatusInternalServerError)
	}

	body, err := base64.StdEncoding.DecodeString(bodyObj.Photo)
	if err != nil {
		http.Error(w, "The JSON object does not contain only the binary photo field",
			http.StatusInternalServerError)
	}

	err = ioutil.WriteFile("last_post.jpg", body, 0644)
	if err != nil {
		http.Error(w, "Internal error while saving the image on local disk",
			http.StatusInternalServerError)
	}

	randID := getRandomID(randStringSize)
	recImgPath := filepath.Join(imagesStorePath, randID+imageFormat)

	err = ioutil.WriteFile(recImgPath, body, 0644)
	if err != nil {
		http.Error(w, "Internal error while saving the image on local disk",
			http.StatusInternalServerError)
	}

	classID, err := executeNNQuery(randID)
	if err != nil {
		log.Panicf("failed to get classID after NN execution")
	}
	fmt.Fprintf(w, "POST done. The result is:\n%v\n", classID)

	ret, err := infogen.GetInfoForClass(classID)
	if err != nil {
		log.Panicf("failed to get YAML bytes for classID %v", classID)
	}

	fmt.Fprintf(w, "YAML %s\n", string(ret))
}

func init() {
	log.SetFlags(log.Lmicroseconds | log.Lshortfile)
	flag.Parse()
}

func main() {
	port := os.Getenv("PORT")

	os.RemoveAll(imagesStorePath)
	absPath, err := filepath.Abs(imagesStorePath)
	if err != nil {
		log.Fatalf("Failed to get absolute path for image store path %v", imagesStorePath)
	}
	os.MkdirAll(absPath, os.ModePerm)

	results = append(results, time.Now().Format(time.RFC3339))

	mux := http.NewServeMux()
	mux.HandleFunc("/", GetHandler)
	mux.HandleFunc("/post", PostHandler)

	// go listenToJobs(mux)

	// Train the neural network using python script.
	// cmd := exec.Command("python3", pythonTrainPath)
	// cmd.Stdout = os.Stdout
	// cmd.Stderr = os.Stderr
	// if err := cmd.Run(); err != nil {
	// 	log.Fatalf("Failed to execute NN training script %v.", pythonTrainPath)
	// }

	log.Printf("listening on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
