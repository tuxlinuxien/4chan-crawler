package lib

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

const (
	baseBoardURL  = "https://a.4cdn.org/%s/%d.json"
	baseThreadURL = "https://a.4cdn.org/%s/thread/%d.json"
	basePicURL    = "https://i.4cdn.org/%s/%d%s"
)

var (
	client = &http.Client{
		Transport: &http.Transport{
			ResponseHeaderTimeout: time.Second * 60,
		},
	}
)

type Task struct {
	Board  string
	Thread uint64
	FileID uint64
	Ext    string
}

func (t *Task) URL() string {
	return fmt.Sprintf(basePicURL, t.Board, t.FileID, t.Ext)
}

func (t *Task) File() string {
	return filepath.Join(
		t.Board,
		strconv.FormatUint(t.Thread, 10),
		strconv.FormatUint(t.FileID, 10)+t.Ext,
	)
}

func GetThreadPage(board string, page int) ([]uint64, error) {
	time.Sleep(time.Second)
	url := fmt.Sprintf(baseBoardURL, board, page)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var out struct {
		Threads []struct {
			Posts []struct {
				No uint64 `json:"no"`
			} `json:"posts"`
		} `json:"threads"`
	}
	json.Unmarshal(content, &out)
	res := make([]uint64, 0)
	for _, thread := range out.Threads {
		for _, post := range thread.Posts {
			res = append(res, post.No)
			break
		}
	}
	return res, nil
}

func GetPostPictures(board string, thread uint64) ([]Task, error) {
	url := fmt.Sprintf(baseThreadURL, board, thread)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	tasks := make([]Task, 0)
	var out struct {
		Posts []struct {
			Ext string `json:"ext"`
			Tim uint64 `json:"tim"`
		} `json:"posts"`
	}
	json.Unmarshal(content, &out)
	for _, post := range out.Posts {
		if post.Ext == "" {
			continue
		}
		t := Task{}
		t.Board = board
		t.Thread = thread
		t.FileID = post.Tim
		t.Ext = post.Ext
		tasks = append(tasks, t)
	}
	return tasks, nil
}

func GetFile(url string, filename string) {
	_, err := os.Stat(filename)
	if !os.IsNotExist(err) {
		log.Println("file", filename, "already exists")
		return
	}
	log.Println("Downloading", url, "to", filename)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println(err)
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return
	}
	if resp.StatusCode != 200 {
		log.Printf("wrong status code: %d\n", resp.StatusCode)
		return
	}
	defer resp.Body.Close()
	f, err := ioutil.TempFile("", "download_")
	if err != nil {
		log.Println(err)
		return
	}
	_, err = io.Copy(f, resp.Body)
	if err != nil {
		log.Println(err)
		f.Close()
		os.Remove(f.Name())
		return
	}
	f.Close()
	os.Rename(f.Name(), filename)
}
