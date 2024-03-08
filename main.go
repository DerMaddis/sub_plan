package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/dermaddis/sub_plan/util"
)

const mainUrl = "https://dsbmobile.de/data/a1e4f6c9-6f35-4f41-aafb-1b41c3843ade/44541846-bcd9-4247-bb6e-2a9475bb5ef7/"
const sessionId = "qunfqjjmdlhelkmcomik4h21"

var myClasses = []string{
	"11E05",
	"12dsp15",
}

var notFoundError = errors.New("page not found")

func main() {
	pages := []string{}
	for i := 1; true; i++ {
		htmlString, err := requestSubst(i)
		if err != nil {
			if errors.Is(err, notFoundError) {
				break
			}
			log.Fatalln(err)
		}
		pages = append(pages, htmlString)
	}

	for _, page := range pages {

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(page))
		if err != nil {
			log.Fatalln(err)
		}

		parseHtml(doc)
	}
}

func requestSubst(n int) (string, error) {
	client := &http.Client{}

	url := fmt.Sprintf(mainUrl+`subst_%03d.htm`, n)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("ASP.NET_SessionId", sessionId)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	htmlBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	htmlString := string(htmlBytes)

	if htmlString == "" {
		return "", notFoundError
	}

	return htmlString, nil
}

func parseHtml(doc *goquery.Document) {
	titleSelection := doc.Find(".mon_title")
	if titleSelection.Length() == 0 {
		return
	}

	doc.Find("tr.list").Slice(1, goquery.ToEnd).Each(func(i int, s *goquery.Selection) {
		data := s.Find("td.list")
		if data.Length() != 7 {
			return
		}

		texts := make([]string, 0, 7)
		data.Each(func(i int, s *goquery.Selection) {
			texts = append(texts, s.Text())
		})

		hours := texts[2]
		room := texts[3]
		className := texts[4]

		compare := func(first string) func(other string) bool {
			return func(other string) bool {
				return first == other
			}
		}

		if util.Some(myClasses, compare(className)) {
			log.Println(className, "in", room, "at", hours)
		}
	})
}
