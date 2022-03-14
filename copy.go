package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"sync"
	"time"
)

var config = Configuration{}

type Configuration struct {
	MaxThreads           int      `json:"maxThreads"`
	Computers            []string `json:"computers"`
	Users                []string `json:"users"`
	BackupFolder         string   `json:"backupFolder"`
	SourceFolderTemplate string   `json:"sourceFolderTemplate"`
	NameTemplate         string   `json:"nameTemplate"`
	UserTemplate         string   `json:"userTemplate"`
	DatetimeFormat       string   `json:"datetimeFormat"`
}

func main() {

	initConfig()
	// prepare
	t := time.Now()
	bFolder := config.BackupFolder + "\\" + t.Format(config.DatetimeFormat)
	os.Mkdir(bFolder, 0777)

	name := [100]string{}
	user := [100]string{}
	folder := [100]string{}
	for i, s := range config.Computers {
		name[i] = s
		user[i] = config.Users[i]
		folder[i] = fmt.Sprintf(config.SourceFolderTemplate, name[i], user[i])
	}

	// start threads

	var ch = make(chan int, len(config.Users)) // This number 50 can be anything as long as it's larger than xthreads
	var wg sync.WaitGroup

	// This starts xthreads number of goroutines that wait for something to do
	wg.Add(config.MaxThreads)
	for i := 0; i < config.MaxThreads; i++ {
		go func() {
			for {
				idx, ok := <-ch
				if !ok { // if there is nothing to do and the channel has been closed then end the goroutine
					wg.Done()
					return
				}
				arch(user[idx], folder[idx], bFolder)
			}
		}()
	}

	// Now the jobs can be added to the channel, which is used as a queue
	for i := 0; i < len(config.Users); i++ {
		ch <- i // add i to the queue
	}

	close(ch) // This tells the goroutines there's nothing else to do
	wg.Wait() // Wait for the threads to finish
}

func initConfig() {

	filename := "./config.json"
	jsonFile, err := os.Open(filename)
	defer func(jsonFile *os.File) {
		err := jsonFile.Close()
		if err != nil {

		}
	}(jsonFile)
	if err != nil {
		fmt.Printf("failed to open json file: %s, error: %v", filename, err)
		return
	}

	jsonData, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		fmt.Printf("failed to read json file, error: %v", err)
		return
	}
	err = json.Unmarshal(jsonData, &config)
	if err != nil {
		fmt.Println("Error decode config")
		panic(err)
	}
}

func arch(name string, path string, backupFolder string) {

	out, err := exec.Command("7z", "a", "-mx0", backupFolder+"\\"+name+".7z", path).Output()

	if err != nil {
		fmt.Println("---ERROR work with " + name + " Folder " + path)
		log.Fatal(err)
	} else {
		fmt.Printf("%s", out)
		fmt.Println("---Done work with " + name + " Folder " + path)
	}
}
