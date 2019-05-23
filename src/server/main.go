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
	infogen "server_answer"

	"github.com/pkg/errors"
)

var (
	// flagPort          = flag.String("port", port, "Port to listen on")
	imageFormat       = ".jpg"
	descriptionFormat = ".txt"
	imagesStorePath   = filepath.Join(os.Getenv("GOPATH"), "data_base", "queries")
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

// 	mux.HandleFunc("/", GetHandler)
func GetHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("in get function")
	jsonBody, err := json.Marshal(results)
	if err != nil {
		http.Error(w, "Error converting results to json",
			http.StatusInternalServerError)
	}
	w.Write(jsonBody)
}

type jsonToken struct {
	Token string `json:"token"`
}

// mux.HandleFunc("/get_history", createHistoryEndpoint)
func createHistoryEndpoint(w http.ResponseWriter, r *http.Request) {
	bodyP, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body",
			http.StatusInternalServerError)
	}

	var bodyObj jsonToken
	if err := json.Unmarshal(bodyP, &bodyObj); err != nil {
		http.Error(w, "Error while apply unmarshal on the received body to get the JSON",
			http.StatusInternalServerError)
	}

	user, err := getUserFromToken(bodyObj.Token)
	if err != nil {
		http.Error(w, "failed to find user from token",
			http.StatusInternalServerError)
	}

	entireHistory, err := GenerateHistoryInfo(user)
	if err != nil {
		http.Error(w, "Internal error while creating the GenerateHistoryInfo",
			http.StatusInternalServerError)
	}

	ret, err := json.Marshal(&entireHistory)
	if err != nil {
		http.Error(w, "Internal error: failed to Marshal the json",
			http.StatusInternalServerError)
	}

	w.Write(ret)
}

// mux.HandleFunc("/all_classes", createAllClassesEndpoint)
func createAllClassesEndpoint(w http.ResponseWriter, r *http.Request) {
	allClasses, err := infogen.GenerateAllClassesInfo()
	if err != nil {
		http.Error(w, "Internal error while creating the AllClassesInfo",
			http.StatusInternalServerError)
	}

	ret, err := json.Marshal(&allClasses)
	if err != nil {
		http.Error(w, "Internal error: failed to Marshal the json",
			http.StatusInternalServerError)
	}

	w.Write(ret)
}

type jsonClass struct {
	Token string `json:"token"`
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

func evaluateReceivedBody(bodyP []byte) ([]byte, error) {
	var bodyObj jsonClass

	if err := json.Unmarshal(bodyP, &bodyObj); err != nil {
		return nil, errors.Wrapf(err, "Error while apply unmarshal on the received body to get the JSON")
	}

	imgBytes, err := base64.StdEncoding.DecodeString(bodyObj.Photo)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed the Base64 decoding of the JSON object")
	}

	user, err := getUserFromToken(bodyObj.Token)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to find user from token")
	}

	path := filepath.Join(os.Getenv("GOPATH"), "data_base", "tmp", "last_post.jpg")
	if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
		return nil, errors.Wrapf(err, "Can not create directory %v", path)
	}

	err = ioutil.WriteFile(path, imgBytes, 0644)
	if err != nil {
		return nil, errors.Wrapf(err, "Internal error while saving the image on local disk")
	}

	randID := getRandomID(randStringSize)
	recImgPath := filepath.Join(imagesStorePath, randID+imageFormat)

	err = ioutil.WriteFile(recImgPath, imgBytes, 0644)
	if err != nil {
		return nil, errors.Wrapf(err, "Internal error while saving the image on local disk")
	}

	classID, err := executeNNQuery(randID)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get classID after NN execution")
	}

	jsonObj, err := infogen.GetInfoForClass(classID)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get YAML bytes for classID %v", classID)
	}
	ret, err := json.Marshal(&jsonObj)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to Marshal the json")
	}

	// Do not forget the history.
	if err := userSaveQuery(user, randID, imgBytes, ret); err != nil {
		return nil, errors.Wrapf(err, "error while saving query in user directory")
	}

	return ret, nil
}

// mux.HandleFunc("/post", PostHandler)
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

	message, err := evaluateReceivedBody(bodyP)
	if err != nil {
		http.Error(w, "Error evaluating the received body",
			http.StatusInternalServerError)
		fmt.Printf("Error while evaluating body: %v", err)
	}
	fmt.Fprintf(w, "%s\n", string(message))
}

func init() {
	log.SetFlags(log.Lmicroseconds | log.Lshortfile)
	flag.Parse()
}

func startServer(port string, needTrain bool) {
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

	mux.HandleFunc("/authenticate", createTokenEndpoint)
	mux.HandleFunc("/protected", protectedEndpoint)
	mux.HandleFunc("/test", validateMiddleware(testEndpoint))

	mux.HandleFunc("/all_classes", createAllClassesEndpoint)
	mux.HandleFunc("/get_history", createHistoryEndpoint)

	if needTrain {
		// Train the neural network using python script.
		fmt.Printf("Starting to train the Neural Network...\n")
		cmd := exec.Command("python3", pythonTrainPath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			log.Fatalf("Failed to execute NN training script %v.", pythonTrainPath)
		}
		fmt.Printf("Training the Neural Network finished!\n")
	}

	if port != "" {
		log.Printf("listening on port %s", port)
		log.Fatal(http.ListenAndServe(":"+port, mux))
	} else {
		fmt.Printf("Please set a PORT and rerun the cmd. ex: $PORT=2020 bin/server\n")
	}
}

func main() {
	port := ""
	if os.Getenv("PORT") != "" {
		port = os.Getenv("PORT")
	}
	needTrain := false
	if os.Getenv("TRAIN") == "true" {
		needTrain = true
	}

	startServer(port, needTrain)

}
