package main

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"mobix-cams-toucher/util"

	"gopkg.in/fsnotify.v1"
)

const (
	URL           = ""
	CONTENT_TYPE  = ""
	GRANT_TYPE    = ""
	CLIENT_ID     = ""
	CLIENT_SECRET = ""

	PENDING_LIST_URL = ""

	FILE_UPLOAD_URL = ""
	MASTER_CATEGORY = "ECIB"
	SUB_CATEGORY    = "ECIB"
	FILE_TYPE       = "pdf"

	DOCUMENT_DIR  = "documents"
	FILE_MOVE_DIR = "uploaded"
)

var watcher *fsnotify.Watcher

func main() {
	// Creates a new file watcher
	watcher, _ = fsnotify.NewWatcher()
	defer watcher.Close()

	// Get the current directory
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))

	// Check for errors
	if err != nil {
		util.Log.Println(err)
	}

	// Starting at the root of the project, walk each file/directory searching for
	// directories
	if err := filepath.Walk(fmt.Sprintf("%s/%s", dir, DOCUMENT_DIR), watchDir); err != nil {
		util.Log.Println("ERROR", err)
	}

	done := make(chan bool)

	// Goroutine to watch for file changes
	go func() {
		for {
			select {
			// watch for events
			case event := <-watcher.Events:
				util.Log.Printf("EVENT! %#v\n", event)

				// Execute the process
				Process()

				// watch for errors
			case err := <-watcher.Errors:
				util.Log.Println("ERROR", err)
			}
		}
	}()

	<-done
}

func watchDir(path string, fi os.FileInfo, err error) error {

	// since fsnotify can watch all the files in a directory, watchers only need
	// to be added to each nested directory
	if fi.Mode().IsDir() {
		return watcher.Add(path)
	}

	return nil
}

func Process() {
	util.Log.Println("Get start process")

	// Get the current directory
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))

	// Check for errors
	if err != nil {
		util.Log.Println(err)
	}

	// Get the files in the current directory
	files, err := ioutil.ReadDir(fmt.Sprintf("%s/%s", dir, DOCUMENT_DIR))
	if err != nil {
		util.Log.Println(err)
	}

	// Get Pending list
	list := GetPendingList()

	// Loop through the pending requests
	for _, p := range list {
		// Loop through the files in the current directory
		for _, f := range files {
			// Check if the file is a pdf
			if Extension(f) {
				// Check if the file name matches the pending request
				if RemoveExtension(f.Name()) == p.Clientele.Identifier {
					// Upload the file
					Upload(f, GetAccessToken(), p.Clientele.IDX, p.Clientele.CreatedBy)
				}
			}
		}
	}
}

func Extension(file fs.FileInfo) bool {

	util.Log.Println(fmt.Sprintf("Get start check extension - %s", file.Name()))

	if file.Name()[len(file.Name())-4:] == ".pdf" {
		return true
	} else {
		return false
	}
}

func RemoveExtension(fileName string) string {
	util.Log.Println(fmt.Sprintf("Remove the file extension - %s", fileName))

	return fileName[:len(fileName)-4]
}

func Upload(f fs.FileInfo, token string, idx string, createdUser string) {

	util.Log.Println(fmt.Sprintf("Get start upload the file - %s", f.Name()))

	// Get the current directory
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))

	// Check for errors
	if err != nil {
		util.Log.Println(err)
	}

	reader, err := os.Open(fmt.Sprintf("%s/%s/%s", dir, DOCUMENT_DIR, f.Name()))

	// Check for errors
	if err == nil {
		// create a new buffer base on file size
		fInfo, _ := reader.Stat()
		var size int64 = fInfo.Size()
		buf := make([]byte, size)

		// read file content into buffer
		fReader := bufio.NewReader(reader)
		fReader.Read(buf)

		// Requst parameters
		data := util.UploadRequest{}
		data.IDX = idx
		data.MasterCategory = MASTER_CATEGORY
		data.SubCategory = SUB_CATEGORY
		data.FileName = f.Name()
		data.ContentType = FILE_TYPE
		data.Longitude = "0"
		data.Latitude = "0"
		data.File = base64.StdEncoding.EncodeToString(buf)

		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}

		client := &http.Client{Transport: tr}

		buffer := new(bytes.Buffer)
		json.NewEncoder(buffer).Encode(data)

		req, err := http.NewRequest("POST", FILE_UPLOAD_URL, buffer)

		if err != nil {
			util.Log.Println(err)
		}

		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Auth-Token", base64.RawStdEncoding.EncodeToString([]byte(createdUser)))

		res, err := client.Do(req)

		if err != nil {
			util.Log.Println(err)
		}

		defer res.Body.Close()

		if res.StatusCode == 200 {
			bodyBytes, err := io.ReadAll(res.Body)

			if err != nil {
				util.Log.Println(err)
			}
			bodyString := string(bodyBytes)
			util.Log.Println(fmt.Sprintf("File uploaded - %s & Response - %s", f.Name(), string(bodyString)))

			// Move the file to the uploaded directory
			MoveUploadedFile(f)
		}
	}

}

func GetAccessToken() string {

	util.Log.Println("Get start get access token")

	params := url.Values{}

	params.Add("grant_type", GRANT_TYPE)
	params.Add("client_id", CLIENT_ID)
	params.Add("client_secret", CLIENT_SECRET)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: tr}

	req, err := http.NewRequest("POST", URL, strings.NewReader(params.Encode()))

	if err != nil {
		util.Log.Println(err)
	}

	req.Header.Set("Content-Type", CONTENT_TYPE)

	res, err := client.Do(req)

	if err != nil {
		util.Log.Println(err)
	}

	defer res.Body.Close()

	util.Log.Println("Get Access Token Response: ", res.Status)

	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		util.Log.Println(err)
	}

	auth := util.AuthResponse{}

	json.NewDecoder(strings.NewReader(string(body))).Decode(&auth)

	return auth.AccessToken
}

func GetPendingList() []util.PendingList {

	util.Log.Println("Get start get pending list")

	// Get access token
	token := GetAccessToken()

	// Get pending list
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: tr}

	req, err := http.NewRequest("GET", PENDING_LIST_URL, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	if err != nil {
		util.Log.Println(err)
	}

	res, err := client.Do(req)

	if err != nil {
		util.Log.Println(err)
	}

	defer res.Body.Close()

	if res.StatusCode == 200 {
		util.Log.Println("Get Pending List Response: ", res.Status)
	} else {
		util.Log.Println("Get Pending List Response: ", res.Status)
	}

	list := []util.PendingList{}

	json.NewDecoder(res.Body).Decode(&list)

	util.Log.Println("Get Pending List: ", list)

	return list
}

func MoveUploadedFile(file fs.FileInfo) {

	util.Log.Println(fmt.Sprintf("Get start move the file - %s", file.Name()))

	// Get the current directory
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))

	// Check for errors
	if err != nil {
		util.Log.Println(err)
	}

	// Move the file
	err = os.Rename(fmt.Sprintf("%s/%s/%s", dir, DOCUMENT_DIR, file.Name()), fmt.Sprintf("%s/%s/%s_%s", dir, FILE_MOVE_DIR, time.Now().Format("20060102150405"), file.Name()))

	if err != nil {
		util.Log.Println(err)
	}

	util.Log.Println(fmt.Sprintf("File moved - %s", file.Name()))
}
