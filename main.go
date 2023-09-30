package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/cuu/grab"
	"github.com/gocolly/colly"
	"io"
	"log"
	"net/url"
	"os"
	"path"
	"strings"
)

const defaultDepth = 2

var (
	urlString string
	maxDepth  int
)

var visitedURLs = make(map[string]bool)
var stateFileExists = false
var stateFileExistsPtr = &stateFileExists
var stateFileName = "crawler_state.txt"

func main() {
	ParseArgs()

	if err := SetupLogging(); err != nil {
		fmt.Printf("ERROR: %v", err)
	}

	log.Printf("INFO: Starting the crawler for %s...", urlString)

	c := colly.NewCollector(
		colly.MaxDepth(maxDepth),
	)

	if _, err := os.Stat(stateFileName); err == nil {
		loadStateFromFile(stateFileName)
		*stateFileExistsPtr = true
	}

	log.Println(visitedURLs)

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		newURL := buildValidURL(urlString, link)
		if link != "" && !hasBeenVisited(newURL) {
			e.Request.Visit(newURL)
			visitedURLs[newURL] = true
		}
	})
	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	c.OnResponse(func(r *colly.Response) {
		log.Printf("resp")

		contentType := r.Headers.Get("Content-Type")
		if strings.HasPrefix(contentType, "text/html") ||
			strings.HasPrefix(contentType, "application/javascript") ||
			strings.HasPrefix(contentType, "text/css") {

			// Check if the response URL belongs to the same domain
			if isSameDomain(r.Request.URL.String()) {
				dirName := setupDir(urlString)
				fileName := path.Join(dirName, path.Base(r.Request.URL.String()))

				if !hasBeenVisited(r.Request.URL.String()) {
					log.Printf("no visits")
					if !hasBeenDownloaded(r.Request.URL.String()) {
						log.Printf("no downloads")
						err := downloadFile(urlString, dirName, fileName)
						if err != nil {
							log.Printf("ERROR: Error saving file %s: %v", fileName, err)
						} else {
							log.Printf("INFO: Downloaded file %s", fileName)
							markAsDownloaded(r.Request.URL.String())
						}
					}
				}
			}
		}
	})

	err := c.Visit(urlString)
	if err != nil {
		log.Fatalf("INFO: Mission failed, check logs. Msg: %v", err)
	}

	log.Println("INFO: Hooray! Mission completed")
	os.Exit(0)
}

// Create a dir to save web pages
func setupDir(urlString string) string {
	dirName := setupDirNameFormat(urlString)
	err := os.MkdirAll(dirName, os.ModePerm)
	if err != nil {
		log.Fatalf("ERROR: Couldn't create a directory, msg: %v", err)
	}

	return dirName
}

func setupDirNameFormat(urlStr string) string {
	dirName := strings.Replace(urlStr, "://", "_", -1)
	dirName = strings.Replace(dirName, "/", "_", -1)
	dirName = strings.Replace(dirName, ".", "_", -1)
	return dirName
}

func SetupLogging() error {
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

func ParseArgs() {
	log.Println("INFO: Parsing args...")
	flag.StringVar(&urlString, "url", "", "URL to start crawling")
	flag.IntVar(&maxDepth, "depth", defaultDepth, "Maximum depth of crawling")
	flag.Parse()

	if urlString == "" {
		fmt.Println("URL can't be empty")
		os.Exit(1)
	}

	if !strings.HasPrefix(urlString, "http://") && !strings.HasPrefix(urlString, "https://") {
		urlString = fmt.Sprintf("https://%s", urlString)
	}

	if maxDepth <= 0 {
		fmt.Println("max depth should be 0 or above")
		os.Exit(1)
	}
}

func downloadFile(urlString, dirName, fileName string) error {
	req, err := grab.NewRequest(dirName, urlString)
	if err != nil {
		return err
	}

	req.Filename = fileName // Set the filename

	client := grab.NewClient()
	resp := client.Do(req)

	// Wait for the download to complete
	<-resp.Done

	if resp.Err() != nil {
		return resp.Err()
	}

	return nil
}

func buildValidURL(hostname string, link string) string {
	if link == "" {
		return ""
	}

	// check if url already absolute, if yes return it straight away
	parsedURL, err := url.Parse(link)
	if err == nil && parsedURL.IsAbs() {
		return link
	}

	// otherwise join the hostname and the link
	u, err := url.Parse(hostname)
	if err != nil {
		return ""
	}
	u.Path = path.Join(u.Path, link)

	return u.String()
}

// Function to check if two URLs belong to the same domain
func isSameDomain(checkURL string) bool {
	u1, err1 := url.Parse(urlString)
	u2, err2 := url.Parse(checkURL)
	if err1 != nil || err2 != nil {
		return false
	}

	return u1.Hostname() == u2.Hostname()
}

// Load previously downloaded URLs from the state file
func loadStateFromFile(stateFileName string) {
	file, err := os.Open(stateFileName)
	if err != nil {
		log.Println("INFO: State file not found, starting from scratch")
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		urlString = scanner.Text()
		//visitedURLs[urlString] = true
	}

	if err = scanner.Err(); err != nil {
		log.Printf("ERROR: Failed to read state file: %v", err)
	}
}

func updateStateFile(urlStr string) {
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

func hasBeenVisited(urlStr string) bool {
	_, ok := visitedURLs[urlStr]
	log.Printf("ok %v for url %v", ok, urlStr)
	return ok
}

func hasBeenDownloaded(urlStr string) bool {
	if *stateFileExistsPtr {
		log.Printf("check state")
		return checkStateFile(urlStr)
	}
	log.Printf("there is no file ")
	return false
}

func markAsDownloaded(urlStr string) {
	visitedURLs[urlStr] = true
	updateStateFile(urlStr)
}

func checkStateFile(urlStr string) bool {
	file, err := os.Open(stateFileName)
	if err != nil {
		log.Printf("ERROR: Failed to open state file for reading: %v", err)
		log.Printf("downloaded false1")
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line == urlStr {
			log.Printf("downloaded true")
			return true
		}
	}

	if err = scanner.Err(); err != nil {
		log.Printf("ERROR: Failed to read from state file: %v", err)
	}
	log.Printf("downloaded false2")
	return false
}
