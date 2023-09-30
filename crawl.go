package main

import (
	"fmt"
	"github.com/gocolly/colly"
	"log"
	"os"
	"path"
	"strings"
)

func Crawl() {
	c := colly.NewCollector(
		colly.MaxDepth(maxDepth),
	)

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		newURL := BuildValidURL(urlString, link)
		if link != "" && !visitedURLs[newURL] {
			visitedURLs[newURL] = true
			e.Request.Visit(newURL)
		}
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	c.OnResponse(func(r *colly.Response) {
		contentType := r.Headers.Get("Content-Type")
		if strings.HasPrefix(contentType, "text/html") ||
			strings.HasPrefix(contentType, "application/javascript") ||
			strings.HasPrefix(contentType, "text/css") {
			handleTxtContent(r.Request.URL.String())
		}
	})

	err := c.Visit(urlString)
	if err != nil {
		log.Fatalf("INFO: Mission failed, check logs. Msg: %v", err)
	}

	log.Println("INFO: Hooray! Mission completed")
	os.Exit(0)
}

func handleTxtContent(urlStr string) {
	if IsSameDomain(urlStr) {
		fileName := path.Join(dirName(), path.Base(urlStr))
		if !HasBeenDownloaded(urlStr) {
			err := DownloadFile(urlStr, dirName(), fileName)
			if err != nil {
				log.Printf("INFO: Error saving file %s: %v", fileName, err)
			} else {
				log.Printf("INFO: Downloaded file %s", fileName)
				MarkAsDownloaded(urlStr)
			}
		}
	}
}
