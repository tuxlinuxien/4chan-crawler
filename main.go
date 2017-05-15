package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/tuxlinuxien/4chan-crawler/lib"
)

const (
	pool = 20
)

var (
	dest  = ""
	board = ""
)

func worker(tasks chan lib.Task, wg *sync.WaitGroup) {
	defer wg.Done()
	for task := range tasks {
		outFile := filepath.Join(dest, task.File())
		lib.GetFile(task.URL(), outFile)
	}
}

func createThreadFolder(board string, threadID uint64) {
	outDir := filepath.Join(dest, board, strconv.FormatUint(threadID, 10))
	os.MkdirAll(outDir, 0755)
}

func main() {
	flag.StringVar(&dest, "dest", "./dest/", "destination folder")
	flag.StringVar(&board, "board", "", "board name (ex: hr)")
	flag.Parse()
	if board == "" {
		log.Println("board cannot be empty")
		return
	}

	os.MkdirAll(dest, 0755)
	tasks := make(chan lib.Task, pool)
	wg := &sync.WaitGroup{}

	for i := 0; i < pool; i++ {
		wg.Add(1)
		go worker(tasks, wg)
	}

	for i := 1; i <= 15; i++ {
		threadsID, err := lib.GetThreadPage(board, i)
		if err != nil {
			continue
		}
		for _, tID := range threadsID {
			taskList, err := lib.GetPostPictures(board, tID)
			if err != nil {
				continue
			}
			if len(taskList) == 0 {
				continue
			}
			createThreadFolder(board, tID)
			for _, task := range taskList {
				tasks <- task
			}
		}
	}
	close(tasks)
	wg.Wait()

}
