package main

import (
	"net/url"
	"path"
)

func BuildValidURL(hostname string, link string) string {
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

func IsSameDomain(checkURL string) bool {
	u1, err1 := url.Parse(urlString)
	u2, err2 := url.Parse(checkURL)
	if err1 != nil || err2 != nil {
		return false
	}

	return u1.Hostname() == u2.Hostname()
}
