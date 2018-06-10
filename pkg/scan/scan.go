package scan

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/bgo-education/test-grader-client/pkg/option"
	"github.com/bgo-education/test-grader-client/pkg/utils"
)

const (
	POST = "POST"
	GET  = "GET"

	JSON = "application/json"

	OK = "OK"
)

var opt = option.GetInstance()
var lenHeader = len(Header(opt.NumCau))

func CheckFolder(folder string) bool {
	f, err := os.Open(filepath.Join(folder, opt.Dst))
	if err != nil {
		return false
	}
	f.Close()
	return true
}

func ProcessFolder(folder, id string, writeChan chan<- []string) error {
	files := utils.GetFilesByType(folder, opt.FilesExtension)
	fmt.Printf("Found %d files\n", len(files))

	client := &http.Client{}
	url := opt.UploadEndPoint + "/" + id
	parans := map[string]string{
		"folder": folder,
		"name":   "",
		"num":    strconv.Itoa(opt.NumCau),
	}

	for _, file := range files {
		if opt.Verbose {
			fmt.Printf("Read %s\n", file)
		}
		parans["name"] = file

		req, err := UploadFile(file, url, parans)
		if req != nil && err == nil {
			res, err := client.Do(req)
			if err != nil {
				fmt.Println(err)
				continue
			}

			body := &bytes.Buffer{}
			_, err = body.ReadFrom(res.Body)
			if err != nil {
				fmt.Println(err)
				continue
			}
			res.Body.Close()

			var data GraderRes
			err = json.Unmarshal(body.Bytes(), &data)
			if err != nil {
				fmt.Println(err)
				continue
			}

			if data.Msg != OK {
				fmt.Println(data.Msg)
				continue
			}

			writeChan <- data.ToSlice(lenHeader)

			if opt.Verbose {
				fmt.Printf("File %s, status code: %d\n", file, res.StatusCode)
			}
		} else {
			fmt.Println(err)
		}
	}

	return nil
}

func UploadFile(filename, url string, params map[string]string) (*http.Request, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filepath.Base(filename))
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(part, f)

	for key, val := range params {
		_ = writer.WriteField(key, val)
	}

	err = writer.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(POST, url, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req, err
}

func CreateSession() (id string, err error) {
	dir, err := utils.GetCurrentDir()
	if err != nil {
		return
	}

	jsonStr := []byte(`{"name":"` + dir + `","override":"` + strconv.FormatBool(true) + `"}`)
	res, err := http.Post(opt.CreateSessionEndPoint, JSON, bytes.NewBuffer(jsonStr))
	if err != nil {
		return
	}
	defer res.Body.Close()

	response := &SessionRes{}
	err = json.NewDecoder(res.Body).Decode(response)
	if err != nil {
		return
	}

	if response.Msg != OK {
		err = errors.New(response.Msg)
		return
	}
	id = response.Idx

	return
}

func GetImagesResult(path, id string) error {
	zipfile := path + ".zip"

	err := utils.DownloadFile(zipfile, opt.DownloadEndPoint+"/"+id)
	if err != nil {
		return err
	}

	err = utils.Unzip(zipfile, path)
	if err != nil {
		return err
	}

	return nil
}
