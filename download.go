package main

import (
	"github.com/cuu/grab"
)

func DownloadFile(urlString, dirName, fileName string) error {
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
