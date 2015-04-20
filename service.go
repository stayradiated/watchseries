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

func makePostRequest(url, host string, data io.Reader) (body []byte, err error) {
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

func makeGetRequest(url, host string) (body []byte, err error) {
	req, err := http.NewRequest("GET", url, nil)
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

	body, err := makePostRequest(pageUrl, g.host, strings.NewReader(data))
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

type KeyBasedVideoHost struct {
	className string
	host      string
	api       string
}

var KEYBASED_FILE_REGEXP = regexp.MustCompile(`flashvars\.file="(\w+)";`)
var KEYBASED_FILEKEY_REGEXP = regexp.MustCompile(`flashvars\.filekey="([\w\.-]+)"`)
var KEYBASED_VIDEO_REGEXP = regexp.MustCompile(`http:\/\/[\w\.]+\/dl\/\w+\/\w+\/\w+\.(?:mp4|flv)`)

func (k *KeyBasedVideoHost) ClassName() string {
	return k.className
}

func (k *KeyBasedVideoHost) ExtractFileFromPage(pageUrl string) (fileUrl string, err error) {

	pageBody, err := makePostRequest(pageUrl, k.host, nil)
	if err != nil {
		return "", err
	}

	matchFile := KEYBASED_FILE_REGEXP.FindStringSubmatch(string(pageBody))
	if len(matchFile) <= 0 {
		return "", ERR_FILE_NOT_FOUND
	}
	file := matchFile[1]

	matchFileKey := KEYBASED_FILEKEY_REGEXP.FindStringSubmatch(string(pageBody))
	if len(matchFileKey) <= 0 {
		return "", ERR_FILE_NOT_FOUND
	}
	fileKey := matchFileKey[1]

	params := url.Values{}
	params.Add("cid3", "watch-series-tv.to")
	params.Add("numOfErrors", "0")
	params.Add("key", fileKey)
	params.Add("file", file)
	params.Add("cid", "1")
	data := params.Encode()

	body, err := makeGetRequest(k.api+"?"+data, k.host)
	if err != nil {
		return "", err
	}

	fileLink := KEYBASED_VIDEO_REGEXP.FindString(string(body))
	if len(fileLink) == 0 {
		return "", ERR_FILE_NOT_FOUND
	}

	return fileLink, nil
}

var NovaMov = &KeyBasedVideoHost{
	host:      "novamov.com",
	className: ".download_link_novamov\\.com",
	api:       "http://www.novamov.com/api/player.api.php",
}
