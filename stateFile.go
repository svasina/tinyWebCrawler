package main

import (
	"bufio"
	"log"
	"os"
)

func checkStateFile(urlStr string) bool {
	file, err := os.Open(stateFileName)
	if err != nil {
		log.Printf("ERROR: Failed to open state file for reading: %v", err)
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line == urlStr {
			return true
		}
	}

	if err = scanner.Err(); err != nil {
		log.Printf("ERROR: Failed to read from state file: %v", err)
	}
	return false
}

func LoadStateFromFile(stateFileName string) {
	log.Printf("check state")
	file, err := os.Open(stateFileName)
	if err != nil {
		log.Println("INFO: State file not found, starting from scratch")
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		urlLine := scanner.Text()
		visitedURLs[urlLine] = true
	}

	if err = scanner.Err(); err != nil {
		log.Printf("ERROR: Failed to read state file: %v", err)
	}
}

func HasBeenDownloaded(urlStr string) bool {
	if *stateFileExistsPtr {
		return checkStateFile(urlStr)
	}
	return false
}

func MarkAsDownloaded(urlStr string) {
	file, err := os.OpenFile(stateFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("ERROR: Failed to open state file %s for writing: %v", stateFileName, err)
		return
	}
	defer file.Close()

	_, err = file.WriteString(urlStr + "\n")
	if err != nil {
		log.Printf("ERROR: Failed to write to state file %s: %v", stateFileName, err)
	}
}
