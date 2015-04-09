package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"

	"github.com/PuerkitoBio/goquery"
)

func main() {

	domain := "http://watchseriestv.to"
	serie := os.Args[1]
	client := &http.Client{}

	gorillavidIdRegex := regexp.MustCompile(`http:\/\/gorillavid\.in\/(\w+)`)
	gorillavidFileRegex := regexp.MustCompile(`file: "[^"]+"`)

	doc, err := goquery.NewDocument(domain + "/serie/" + serie)
	if err != nil {
		log.Fatal(err)
	}

	doc.Find("div[itemprop=season]").Each(func(i int, _season *goquery.Selection) {
		heading := _season.Find("h2 span").Text()
		fmt.Println("# " + heading)

		_season.Find(".listings li[itemprop=episode]").Each(func(i int, _episode *goquery.Selection) {
			name := _episode.Find("[itemprop=name]").Text()
			episodeLink, _ := _episode.Find("a").Attr("href")

			fmt.Println(" - " + name)

			_episodeLinks, err := goquery.NewDocument(domain + episodeLink)
			if err != nil {
				log.Fatal(err)
			}

			cale, _ := _episodeLinks.Find(".download_link_gorillavid\\.in").First().Find(".buttonlink").Attr("href")

			_cale, err := goquery.NewDocument(domain + cale)
			if err != nil {
				log.Fatal(err)
			}

			gorillaLink, _ := _cale.Find(".push_button.blue").Attr("href")

			match := gorillavidIdRegex.FindStringSubmatch(gorillaLink)
			id := ""
			if len(match) > 1 {
				id = match[1]
			} else {
				return
				// return false
			}

			data := url.Values{}
			data.Add("op", "download1")
			data.Add("id", id)
			data.Add("method_free", "Free Download")

			body := bytes.NewReader([]byte(data.Encode()))
			req, err := http.NewRequest("POST", gorillaLink, body)
			if err != nil {
				log.Fatal(err)
			}

			req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
			req.Header.Set("Accept-Language", "en-US,en;q=0.8")
			req.Header.Set("Cache-Control", "no-cache")
			req.Header.Set("Connection", "keep-alive")
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			req.Header.Set("Cookie", "lang=english; aff=116885; ad_referer=")
			req.Header.Set("Host", "gorillavid.in")
			req.Header.Set("Origin", "http://gorillavid.in")
			req.Header.Set("Pragma", "no-cache")
			// req.Header.Set("Referer", "http://gorillavid.in/a0ftnvspv52c")
			req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_10_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/40.0.2214.91 Safari/537.36")

			res, err := client.Do(req)
			if err != nil {
				log.Fatal(err)
			}

			resBody, err := ioutil.ReadAll(res.Body)
			if err != nil {
				log.Fatal(err)
			}

			fileLink := gorillavidFileRegex.FindStringSubmatch(string(resBody))
			fmt.Println(fileLink)
		})
	})

}
