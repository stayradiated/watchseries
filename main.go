package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

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

type Episode struct {
	Id   int
	Name string
	Path string
	Link string
}

type EpisodeSlice []*Episode

func (e EpisodeSlice) Len() int           { return len(e) }
func (e EpisodeSlice) Less(i, j int) bool { return e[i].Id < e[j].Id }
func (e EpisodeSlice) Swap(i, j int)      { e[i], e[j] = e[j], e[i] }

type Season struct {
	Id       int
	Path     string
	Episodes []*Episode
}

type SeasonSlice []*Season

func (s SeasonSlice) Len() int           { return len(s) }
func (s SeasonSlice) Less(i, j int) bool { return s[i].Id < s[j].Id }
func (s SeasonSlice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

func main() {

	rand.Seed(time.Now().UTC().UnixNano())

	domain := "http://watch-series-tv.to"

	var reverseSort bool
	flag.BoolVar(&reverseSort, "r", false, "Reverse sort")

	var serie string
	flag.StringVar(&serie, "s", "", "Series ID ")

	flag.Parse()

	if serie == "" {
		log.Fatal("Must specify serie id")
		return
	}

	EPISODE_REGEXP := regexp.MustCompile(`Episode\s+(\d+)([^$]+)`)
	SEASON_REGEXP := regexp.MustCompile(`Season\s+(\d+)`)

	services := []ServiceScraper{
		NovaMov,
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

	seasons := make([]*Season, 0)

	doc.Find("div[itemprop=season]").Each(func(i int, _season *goquery.Selection) {
		season := &Season{
			Id:       0,
			Episodes: make([]*Episode, 0),
		}
		seasons = append(seasons, season)

		heading := _season.Find("h2 span").Text()

		match := SEASON_REGEXP.FindStringSubmatch(heading)
		if len(match) < 2 {
			log.Println("Could not find season ID for " + heading)
		} else {
			season.Id, _ = strconv.Atoi(match[1])
		}

		season.Path = fmt.Sprintf("Season %d", season.Id)

		_season.Find(".listings li[itemprop=episode]").Each(func(i int, _episode *goquery.Selection) {
			episode := &Episode{
				Id: 0,
			}
			season.Episodes = append(season.Episodes, episode)

			text := _episode.Find("[itemprop=name]").Text()

			match := EPISODE_REGEXP.FindStringSubmatch(text)
			episode.Name = "Unknown"
			if len(match) < 3 {
				log.Println("Could not find episode details for " + text)
			} else {
				episode.Id, _ = strconv.Atoi(match[1])
				episode.Name = sanitize.Path(strings.TrimSpace(match[2]))
			}

			episode.Path = fmt.Sprintf("%s s%de%d - %s",
				seriesTitle, season.Id, episode.Id, episode.Name)

			episode.Link, _ = _episode.Find("a").Attr("href")

		})
	})

	if reverseSort {
		sort.Sort(sort.Reverse(SeasonSlice(seasons)))
	} else {
		sort.Sort(SeasonSlice(seasons))
	}

	for _, season := range seasons {
		if reverseSort {
			sort.Sort(sort.Reverse(SeasonSlice(seasons)))
			sort.Sort(sort.Reverse(EpisodeSlice(season.Episodes)))
		} else {
			sort.Sort(SeasonSlice(seasons))
			sort.Sort(EpisodeSlice(season.Episodes))
		}

		for _, episode := range season.Episodes {

			if matches, _ := filepath.Glob(season.Path + "/" + episode.Path + ".*"); matches != nil {
				continue
			}

			fmt.Println("- " + season.Path + " / " + episode.Path)

			_episodeLinks, err := goquery.NewDocument(domain + episode.Link)
			if err != nil {
				log.Println(err)
				continue
			}

			Shuffle(services)

			for _, service := range services {

				fmt.Println(service.ClassName())

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

				fmt.Println(episode.Path)

				params := make(map[string]interface{})
				params["max-connection-per-server"] = "5"
				params["max-concurrent-downloads"] = "5"
				params["out"] = season.Path + "/" + episode.Path + filepath.Ext(fileUrl)
				aria2rpc.AddUri(fileUrl, params)
				break
			}
		}
	}

}
