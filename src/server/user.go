package main

import (
	"encoding/base64"
	"encoding/json"

	"github.com/pkg/errors"

	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/mitchellh/mapstructure"
	infogen "server_answer"
)

type QueryHistory struct {
	Photo  string
	Answer infogen.QueryInfo
}

type HistoryData struct {
	Data []QueryHistory
}

// GenerateHistoryInfo returns all the info about the user's history.
func GenerateHistoryInfo(user string) (*HistoryData, error) {
	// var ret HistoryData
	userDir := getUserDir(user)
	sizeHistory := getNoQueries(user, false)

	history := make([]QueryHistory, sizeHistory)

	for index := 1; index <= sizeHistory; index++ {
		fdImg, err := os.Open(filepath.Join(userDir, strconv.Itoa(index)+".jpg"))
		if err != nil {
			return nil, errors.Wrapf(err, "failed to open img file")
		}
		defer fdImg.Close()

		// Make base64 from []byte image.
		imgBytes := make([]byte, 200000)
		sizeOutput, err := fdImg.Read(imgBytes)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to read image bytes")
		}
		imgBytes = imgBytes[:sizeOutput]

		var act QueryHistory

		act.Photo = base64.StdEncoding.EncodeToString(imgBytes)

		fdJSON, err := os.Open(filepath.Join(userDir, strconv.Itoa(index)+".json"))
		if err != nil {
			return nil, errors.Wrapf(err, "failed to open img file")
		}
		defer fdJSON.Close()

		jsonBytes := make([]byte, 200000)
		sizeOutput, err = fdJSON.Read(jsonBytes)
		jsonBytes = jsonBytes[:sizeOutput]

		err = json.Unmarshal(jsonBytes, &act.Answer)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to unmarshal the internal history json")
		}

		history[index - 1] = act
	}
	var ret HistoryData
	ret.Data = history
	return &ret, nil
}

func getUserDir(user string) string {
	path := filepath.Join(os.Getenv("GOPATH"), "data_base", "users_data", user)
	os.MkdirAll(path, os.ModePerm)
	return path
}

func userServerInteraction(user string, message string) error {
	userDir := getUserDir(user)

	// Add an login note into a log file.
	f, err := os.OpenFile(filepath.Join(userDir, "user.log"), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return errors.Wrapf(err, "error while opening 'user.log'")
	}
	defer f.Close()
	if _, err = f.WriteString(
		fmt.Sprintf("%s: %v\n", message, time.Now().Format("2006-01-02 15:04:05.000000000")),
	); err != nil {
		return errors.Wrapf(err, "error while updating 'user.log'")
	}
	return nil
}

func getUserFromToken(tokenStr string) (string, error) {
	if tokenStr == "" {
		return "none", nil
	}

	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("There was an error")
		}
		return []byte("secret"), nil
	})
	if err != nil {
		return "", errors.Wrapf(err, "could not parse received token '%s'", tokenStr)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		var user User
		mapstructure.Decode(claims, &user)

		return user.Username, nil
	}
	// If token not valid, return empty-name user.
	return "none", nil
}

func getNoQueries(user string, increment bool) int {
	countFilePath := filepath.Join(getUserDir(user), "counter.txt")

	if _, err := os.Stat(countFilePath); os.IsNotExist(err) {
		f, err := os.Create(countFilePath)
		if err != nil {
			panic(fmt.Sprintf("error while creating %s: %v", countFilePath, err))
		}
		f.WriteString("0")
		f.Close()
	}

	fd, err := os.Open(countFilePath)
	if err != nil {
		panic(fmt.Sprintf("open %s: %v", countFilePath, err))
	}
	var line int
	_, err = fmt.Fscanf(fd, "%d", &line)
	if err != nil {
		panic(fmt.Sprintf("error while reading int from %s: %v", countFilePath, err))
	}
	if err := fd.Close(); err != nil {
		panic(fmt.Sprintf("error closing file descriptor: %v", err))
	}
	if increment == false {
		return line
	}

	f, err := os.Create(countFilePath)
	if err != nil {
		fmt.Printf("error creating file: %v", err)
		return 0
	}
	defer f.Close()
	_, err = f.WriteString(fmt.Sprintf("%d\n", line+1))
	if err != nil {
		fmt.Printf("error writing string: %v", err)
	}

	return line + 1
}

func userSaveQuery(user string, randID string, imgBytes []byte, answerContent []byte) error {
	if err := userServerInteraction(
		user,
		"ask query "+randID,
	); err != nil {
		return errors.Wrapf(err, "failed to save in user log file")
	}

	userDir := getUserDir(user)

	indexActual := strconv.Itoa(getNoQueries(user, true))

	err := ioutil.WriteFile(filepath.Join(userDir, indexActual+".jpg"), imgBytes, 0644)
	if err != nil {
		return errors.Wrapf(err, "Internal error while saving the image on local disk")
	}

	err = ioutil.WriteFile(filepath.Join(userDir, indexActual+".json"), answerContent, 0644)
	if err != nil {
		return errors.Wrapf(err, "Internal error while saving the answer on local disk")
	}

	return nil
}
