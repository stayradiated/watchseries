package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/NoahShen/aria2rpc"
	"github.com/PuerkitoBio/goquery"
	"github.com/kennygrant/sanitize"
)

func Shuffle(a []ServiceScraper) {
	for i := range a {
		j := rand.Intn(i + 1)
		a[i], a[j] = a[j], a[i]
	}
}

func main() {

	domain := "http://watch-series-tv.to"
	serie := os.Args[1]

	EPISODE_REGEXP := regexp.MustCompile(`Episode\s+(\d+)([^$]+)`)
	SEASON_REGEXP := regexp.MustCompile(`Season\s+(\d+)`)

	services := []ServiceScraper{
		VodLocker,
		GorillaVid,
		DaClips,
		MovPod,
	}

	doc, err := goquery.NewDocument(domain + "/serie/" + serie)
	if err != nil {
		log.Fatal(err)
	}

	seriesTitle := doc.Find(".channel-title [itemprop=name]").Text()

	doc.Find("div[itemprop=season]").Each(func(i int, _season *goquery.Selection) {
		heading := _season.Find("h2 span").Text()

		match := SEASON_REGEXP.FindStringSubmatch(heading)
		seasonId := "0"
		if len(match) < 2 {
			log.Println("Could not find season ID for " + heading)
		} else {
			seasonId = match[1]
		}

		fmt.Println("# " + heading)

		_season.Find(".listings li[itemprop=episode]").Each(func(i int, _episode *goquery.Selection) {
			text := _episode.Find("[itemprop=name]").Text()

			match := EPISODE_REGEXP.FindStringSubmatch(text)
			episodeId := "0"
			episodeTitle := "Unknown"
			if len(match) < 3 {
				log.Println("Could not find episode details for " + text)
			} else {
				episodeId = match[1]
				episodeTitle = sanitize.Path(strings.TrimSpace(match[2]))
			}

			filename := fmt.Sprintf("%s s%se%s - %s",
				seriesTitle, seasonId, episodeId, episodeTitle)

			fmt.Println("- " + filename)

			if matches, _ := filepath.Glob(filename + ".*"); matches != nil {
				fmt.Println("Already downloaded")
				return
			}

			episodeLink, _ := _episode.Find("a").Attr("href")

			_episodeLinks, err := goquery.NewDocument(domain + episodeLink)
			if err != nil {
				log.Println(err)
				return
			}

			Shuffle(services)

			for _, service := range services {

				_serviceLinks := _episodeLinks.Find(service.ClassName())
				if _serviceLinks.Length() < 1 {
					continue
				}

				fileUrl := ""

				_serviceLinks.Each(func(i int, _episode *goquery.Selection) {

					if fileUrl != "" {
						return
					}

					caleLink, _ := _serviceLinks.First().Find(".buttonlink").Attr("href")

					_cale, err := goquery.NewDocument(domain + caleLink)
					if err != nil {
						return
					}

					pageUrl, _ := _cale.Find(".push_button.blue").Attr("href")

					fileUrl, err = service.ExtractFileFromPage(pageUrl)
					if err != nil {
						log.Println(err)
						return
					}
				})

				if fileUrl == "" {
					continue
				}

				fmt.Println(filename)

				fmt.Println(fileUrl)
				params := make(map[string]interface{})
				params["max-connection-per-server"] = "5"
				params["max-concurrent-downloads"] = "25"
				params["out"] = filename + filepath.Ext(fileUrl)
				aria2rpc.AddUri(fileUrl, params)
				break
			}

		})
	})

}
