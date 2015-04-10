package main

import (
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type ServiceScraper interface {
	ClassName() string
	ExtractFileFromPage(pageUrl string) (fileUrl string, err error)
}

var httpClient = &http.Client{}

// errors

var ERR_FILE_NOT_FOUND = errors.New("Could not find file url")
var ERR_INVALID_URL = errors.New("Invalid page url")

// regexs
var GENERIC_FILE_REGEXP = regexp.MustCompile(`http:\/\/[\w\.]+(?::\d{1,4})?\/\w+\/\w+.(:?mp4|flv)`)

// utilities

func makeRequest(url, host string, data io.Reader) (body []byte, err error) {
	req, err := http.NewRequest("POST", url, data)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept-Language", "en-US,en;q=0.8")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Host", host)
	req.Header.Set("Origin", "http:// + host")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_10_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/40.0.2214.91 Safari/537.36")

	res, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	body, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

type GenericVideoHost struct {
	className string
	host      string
	params    map[string]string
}

func (g *GenericVideoHost) ClassName() string {
	return g.className
}

func (g *GenericVideoHost) ExtractFileFromPage(pageUrl string) (fileUrl string, err error) {
	_waitPage, err := goquery.NewDocument(pageUrl)
	if err != nil {
		return "", err
	}

	params := url.Values{}
	for name, selector := range g.params {
		value, _ := _waitPage.Find(selector).Attr("value")
		params.Add(name, value)
	}
	data := params.Encode()

	body, err := makeRequest(pageUrl, g.host, strings.NewReader(data))
	if err != nil {
		return "", err
	}

	fileLink := GENERIC_FILE_REGEXP.FindString(string(body))
	if len(fileLink) == 0 {
		return "", ERR_FILE_NOT_FOUND
	}

	return fileLink, nil
}

var GorillaVid = &GenericVideoHost{
	host:      "gorillavid.in",
	className: ".download_link_gorillavid\\.in",
	params: map[string]string{
		"op":          "input[type=hidden][name=op]",
		"id":          "input[type=hidden][name=id]",
		"fname":       "input[type=hidden][name=fname]",
		"method_free": "input[type=hidden][name=method_free]",
	},
}

var DaClips = &GenericVideoHost{
	host:      "daclips.in",
	className: ".download_link_daclips\\.in",
	params: map[string]string{
		"op":          "input[type=hidden][name=op]",
		"id":          "input[type=hidden][name=id]",
		"fname":       "input[type=hidden][name=fname]",
		"method_free": "input[type=hidden][name=method_free]",
	},
}

var VodLocker = &GenericVideoHost{
	host:      "vodlocker.com",
	className: ".download_link_vodlocker\\.com",
	params: map[string]string{
		"op":      "form[method=POST] input[type=hidden][name=op]",
		"id":      "input[type=hidden][name=id]",
		"fname":   "input[type=hidden][name=fname]",
		"hash":    "input[type=hidden][name=hash]",
		"imhuman": "input[type=submit][name=imhuman]",
	},
}

var MovPod = &GenericVideoHost{
	host:      "movpod.in",
	className: ".download_link_movpod\\.in",
	params: map[string]string{
		"op":          "input[type=hidden][name=op]",
		"id":          "input[type=hidden][name=id]",
		"fname":       "input[type=hidden][name=fname]",
		"method_free": "input[type=hidden][name=method_free]",
	},
}
