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
	MaxThreads           int    `json:"maxThreads"`
	MaxComputers         int    `json:"maxComputers"`
	BackupFolder         string `json:"backupFolder"`
	SourceFolderTemplate string `json:"sourceFolderTemplate"`
	NameTemplate         string `json:"nameTemplate"`
	DatetimeFormat       string `json:"datetimeFormat"`
}

func main() {

	initConfig()
	// prepare
	//sourceFolder := "d:\\test\\%[1]v"
	//sourceFolder := "\\\\fs\\students\\events\\WSR2022\\CAD\\%[1]v"
	//sourceFolder := "\\\\%[1]v\\c$\\Users\\%[1]v\\Desktop"

	name := [100]string{}
	folder := [100]string{}
	for i := 1; i <= config.MaxComputers; i++ {
		name[i-1] = fmt.Sprintf(config.NameTemplate, fmt.Sprintf("%02d", i))
		//name[i-1] = "cad-" + fmt.Sprintf("%02d", i)
		folder[i-1] = fmt.Sprintf(config.SourceFolderTemplate, name[i-1])

	}

	// start threads

	var ch = make(chan int, config.MaxComputers) // This number 50 can be anything as long as it's larger than xthreads
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
				arch(name[idx], folder[idx], config.BackupFolder)
			}
		}()
	}

	// Now the jobs can be added to the channel, which is used as a queue
	for i := 0; i < config.MaxComputers; i++ {
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
	fmt.Println("---Start work with " + name + " Folder " + path)
	t := time.Now()
	out, err := exec.Command("7z", "a", "-mx0", backupFolder+"\\"+name+"_"+t.Format(config.DatetimeFormat)+".7z", path).Output()
	fmt.Println("---Done with " + name)
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Printf("%s", out)
	}
}
