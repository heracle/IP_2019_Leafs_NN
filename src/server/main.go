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
	"strings"
	"sync"
	"time"

	"encoding/base64"
)

var (
	// port                = "2020"
	// flagPort          = flag.String("port", port, "Port to listen on")
	imageFormat       = ".jpg"
	descriptionFormat = ".txt"
	imagesStorePath   = "data_store/"
	randStringSize    = 5
	receiveURL        = "/receive/"
	// receiveCompleteURL  = port + receiveURL
	pythonAskForJobPath = "src/nn/ask_for_job.py"
	pythonTrainPath     = "src/nn/train.py"
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

func printPredictionResult(mux *http.ServeMux, jobID string, output []byte) {

	localHandleFunction := func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("in local function")
		jsonBody, err := json.Marshal(strings.TrimSpace(string(output)))
		if err != nil {
			http.Error(w, "Error converting results to json",
				http.StatusInternalServerError)
		}
		w.Write(jsonBody)
	}

	mux.HandleFunc(receiveURL+jobID, localHandleFunction)
}

func listenToJobs(mux *http.ServeMux) {
	for true {
		time.Sleep(2 * time.Second)

		jobQueueMutex.Lock()
		actualQueueSize := len(jobQueue)
		jobQueueMutex.Unlock()

		for i := 0; i < actualQueueSize; i++ {
			jobQueueMutex.Lock()
			actualJobID := jobQueue[0]
			jobQueue = jobQueue[1:]
			jobQueueMutex.Unlock()

			cmd := exec.Command("python3", pythonAskForJobPath, actualJobID)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				log.Panicf("Failed to execute %v for job %v.", pythonAskForJobPath, actualJobID)
			}

			f, err := os.Open(filepath.Join(imagesStorePath, actualJobID+descriptionFormat))
			if err != nil {
				log.Panicf("No file created after execution of %v for job %v.", pythonAskForJobPath, actualJobID)
			}
			output := make([]byte, maxDescriptionSize)
			sizeOutput, err := f.Read(output)
			if err != nil {
				log.Panicf("Failed to read the description file created after execution of %v for job %v.", pythonAskForJobPath, actualJobID)
			}
			// We need just the first sizeOutput bytes.
			output = output[:sizeOutput]

			go printPredictionResult(mux, actualJobID, output)
		}
	}
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

type jsonObj struct {
	Photo string `json:"photo"`
}

// PostHandler converts post request body to string
func PostHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("here in POST function")

	// ip, _, err := net.SplitHostPort(r.RemoteAddr)
	// if err != nil {
	// 	http.Error(w, "Error reading actual Ip",
	// 		http.StatusInternalServerError)
	// }
	// userIP := net.ParseIP(ip)
	// fmt.Println("my ip is " + userIP.String())

	if r.Method == "POST" {
		fmt.Println("also a POST method")
		bodyP, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Error reading request body",
				http.StatusInternalServerError)
		}
		// results = append(results, string(bodyP))

		var bodyObj jsonObj
		err = json.Unmarshal(bodyP, &bodyObj)
		if err != nil {
			http.Error(w, "Error while unmarshal body",
				http.StatusInternalServerError)
		}

		body, err := base64.StdEncoding.DecodeString(bodyObj.Photo)
		if err != nil {
			http.Error(w, "The JSON object does not contain only a photo field",
				http.StatusInternalServerError)
		}

		err = ioutil.WriteFile("last_post.jpg", body, 0644)
		if err != nil {
			http.Error(w, "Internal error",
				http.StatusInternalServerError)
		}

		// fmt.Printf("%s\n", body)

		randID := getRandomID(randStringSize)
		recImgPath := filepath.Join(imagesStorePath, randID+imageFormat)

		fmt.Println(recImgPath)

		out, err := os.Create(recImgPath)
		if err != nil {
			http.Error(w, "Error creating the empty image file",
				http.StatusInternalServerError)
		}

		if _, err = out.Write(body); err != nil {
			http.Error(w, "Error adding the jpg bytes to the created file",
				http.StatusInternalServerError)
		}

		// jobQueueMutex.Lock()
		// jobQueue = append(jobQueue, randID)
		// jobQueueMutex.Unlock()

		// fmt.Fprintf(w, "POST done. Please wait for the result at: %s\n", randID)

		cmd := exec.Command("python3", pythonAskForJobPath, randID)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			log.Panicf("Failed to execute %v for job %v.", pythonAskForJobPath, randID)
		}

		f, err := os.Open(filepath.Join(imagesStorePath, randID+descriptionFormat))
		if err != nil {
			log.Panicf("No file created after execution of %v for job %v.", pythonAskForJobPath, randID)
		}
		output := make([]byte, maxDescriptionSize)
		sizeOutput, err := f.Read(output)
		if err != nil {
			log.Panicf("Failed to read the description file created after execution of %v for job %v.", pythonAskForJobPath, randID)
		}
		// We need just the first sizeOutput bytes.
		output = output[:sizeOutput]
		fmt.Fprintf(w, "POST done. The result is:\n%s\n", string(output))
	} else {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
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
