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
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var (
	port                = "23"
	flagPort            = flag.String("port", port, "Port to listen on")
	imageFormat         = ".jpeg"
	descriptionFormat   = ".txt"
	imagesStorePath     = "data_store/"
	randStringSize      = 20
	receiveURL          = "/receive/"
	receiveCompleteURL  = port + receiveURL
	pythonAskForJobPath = "server/ask_for_job.py"
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

// PostHandler converts post request body to string
func PostHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("here in POST function")

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		http.Error(w, "Error reading actual Ip",
				http.StatusInternalServerError)
	}
	userIP := net.ParseIP(ip)
	fmt.Println("my ip is " + userIP.String() )

	if r.Method == "POST" {
		fmt.Println("also a POST method")
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Error reading request body",
				http.StatusInternalServerError)
		}
		results = append(results, string(body))

		randID := getRandomID(randStringSize)
		recImgPath := filepath.Join(imagesStorePath, randID+imageFormat)

		fmt.Println(recImgPath)

		out, err := os.Create(recImgPath)
		if err != nil {
			http.Error(w, "Error creating the empty image file",
				http.StatusInternalServerError)
		}

		if _, err = out.Write(body); err != nil {
			http.Error(w, "Error adding the jpeg bytes to the created file",
				http.StatusInternalServerError)
		}

		jobQueueMutex.Lock()
		jobQueue = append(jobQueue, randID)
		jobQueueMutex.Unlock()

		fmt.Fprintf(w, "POST done. Please wait for the result at: %s", receiveCompleteURL+randID)
	} else {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
}

func init() {
	log.SetFlags(log.Lmicroseconds | log.Lshortfile)
	flag.Parse()
}

func main() {
	results = append(results, time.Now().Format(time.RFC3339))

	mux := http.NewServeMux()
	mux.HandleFunc("/", GetHandler)
	mux.HandleFunc("/post", PostHandler)

	go listenToJobs(mux)

	log.Printf("listening on port %s", *flagPort)
	log.Fatal(http.ListenAndServe(":"+*flagPort, mux))
}
