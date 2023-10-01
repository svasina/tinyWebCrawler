package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strings"
	"sync"
)

const defaultDepth = 2

var (
	urlString          string
	maxDepth           int
	visitedURLs        = make(map[string]bool)
	visitedURLsLock    = sync.Mutex{}
	stateFileExists    = false
	stateFileExistsPtr = &stateFileExists
	stateFileName      = "crawler_state.txt"
)

var dirName = func() string {
	tmpName := setupDirNameFormat()
	err := os.MkdirAll(tmpName, os.ModePerm)
	if err != nil {
		log.Fatalf("ERROR: Couldn't create a directory, msg: %v", err)
	}

	return tmpName
}

func main() {
	parseArgs()

	if err := setupLogging(); err != nil {
		fmt.Printf("ERROR: %v", err)
		os.Exit(1)
	}

	log.Printf("INFO: Starting the crawler for %s...", urlString)

	if _, err := os.Stat(stateFileName); err == nil {
		LoadStateFromFile(stateFileName)
		*stateFileExistsPtr = true
	}

	Crawl()
}

func setupDirNameFormat() string {
	re := regexp.MustCompile(`[:/.]`)
	nameFormat := re.ReplaceAllString(urlString, "_")
	return nameFormat
}

func setupLogging() error {
	// save logs to file
	logFile, err := os.OpenFile("webCrawlerLog.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	// also print to stdout
	mw := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(mw)
	return nil
}

func parseArgs() {
	log.Println("INFO: Parsing args...")
	flag.StringVar(&urlString, "url", "", "URL to start crawling")
	flag.IntVar(&maxDepth, "depth", defaultDepth, "Maximum depth of crawling")
	flag.Parse()

	if urlString == "" {
		fmt.Println("URL can't be empty")
		printUsage()
		os.Exit(1)
	}

	if !strings.HasPrefix(urlString, "http://") && !strings.HasPrefix(urlString, "https://") {
		urlString = fmt.Sprintf("https://%s", urlString)
	}

	if maxDepth <= 0 {
		fmt.Println("max depth should be above 0")
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  main -url <URL> [-depth <depth>]")
	fmt.Println()
	fmt.Println("Flags:")
	flag.PrintDefaults()
}
