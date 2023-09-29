package main

import (
	"bytes"
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"github.com/cuu/grab"
	"golang.org/x/net/html"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"
)

const defaultDepth = 3

var (
	urlString string // url should be set without https://
	maxDepth  int
)

var visitedURLs = make(map[string]bool)

func main() {
	ParseArgs()

	if err := SetupLogging(); err != nil {
		fmt.Printf("ERROR: %v", err)
	}

	log.Printf("INFO: Starting the crawler for %s...", urlString)

	newClient := InitHTTPClient()
	dirName := setupDir(urlString)

	if err := Crawler(urlString, dirName, &newClient, 0); err != nil {
		log.Fatalf("INFO: Mission failed, check logs. Msg: %v", err)
	}

	log.Println("INFO: Hooray! Mission completed")
	os.Exit(0)
}

// Create a dir to save web pages
func setupDir(urlString string) string {
	dirName := setupFileName(urlString)
	err := os.MkdirAll(dirName, os.ModePerm)
	if err != nil {
		log.Fatalf("ERROR: Couldn't create a directory, msg: %v", err)
	}
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

func InitHTTPClient() http.Client {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}
	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	return http.Client{
		Timeout:   20 * time.Second,
		Transport: transport,
	}
}

func Crawler(urlString, dirName string, client *http.Client, currentDepth int) error {
	if currentDepth > maxDepth {
		log.Printf("INFO: Stopping crawling, max depth reached (depth %d)...", maxDepth)
		return nil
	}

	visitedURLs[urlString] = true

	body, err := ConnectToURL(urlString, client)
	if err != nil {
		return err
	}

	// Check if the URL ends with a known file extension
	if shouldDownload(urlString) {
		// Download the file using grab
		err = downloadFile(urlString, dirName)
		if err != nil {
			return errors.New(fmt.Sprintf("error downloading file: %v", err))
		}
		log.Printf("INFO: Downloaded file %s, depth is %d, max depth is %d", urlString, currentDepth, maxDepth)
		return nil
	}

	// Parse the response body as HTML and find links to continue crawling
	links := findLinks(body)
	for _, link := range links {
		newUrlString := link
		// Check if URL has already been visited
		if visitedURLs[newUrlString] {
			log.Printf("INFO: Skipping already visited URL: %s", newUrlString)
			continue
		}
		err = Crawler(newUrlString, dirName, client, currentDepth+1)
		if err != nil {
			return err
		}
	}
	return nil
}

func ConnectToURL(urlString string, client *http.Client) ([]byte, error) {
	resp, err := client.Get(urlString)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("website responded with %v", resp.StatusCode))
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func setupFileName(url string) string {
	fileName := strings.Replace(url, "://", "_", -1)
	fileName = strings.Replace(fileName, "/", "_", -1)
	fileName = strings.Replace(fileName, ".", "_", -1)
	return fileName + "_" + time.Now().Format("20060102150405.000")
}

func saveResponseBodyToFile(fileName string, body []byte) error {
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(body)
	if err != nil {
		return err
	}

	return nil
}

func findLinks(body []byte) []string {
	links := make([]string, 0)

	reader := bytes.NewReader(body)
	tokenizer := html.NewTokenizer(reader)

	for {
		tokenType := tokenizer.Next()

		switch tokenType {
		case html.ErrorToken:
			return links // End of document
		case html.StartTagToken, html.SelfClosingTagToken:
			token := tokenizer.Token()

			if token.Data == "a" {
				for _, attr := range token.Attr {
					if attr.Key == "href" {
						link := attr.Val
						// Ensure the link is not empty and is a valid URL
						newLink := buildValidURL(urlString, link)
						if newLink != "" {
							links = append(links, newLink)
						}
					}
				}
			}
		}
	}
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

	// otherwise join the hostname and the linl
	u, err := url.Parse(hostname)
	if err != nil {
		return ""
	}
	u.Path = path.Join(u.Path, link)

	return u.String()
}

func shouldDownload(urlString string) bool {
	// Define a list of file extensions you want to download
	allowedExtensions := []string{".html", ".htm", ".js", ".css"}
	for _, ext := range allowedExtensions {
		if strings.HasSuffix(urlString, ext) {
			return true
		}
	}

	return false
}

func downloadFile(urlString, dirName string) error {
	req, err := grab.NewRequest(dirName, urlString)
	if err != nil {
		return err
	}

	client := grab.NewClient()
	resp := client.Do(req)

	// Wait for the download to complete
	<-resp.Done

	if resp.Err() != nil {
		return resp.Err()
	}

	return nil
}
