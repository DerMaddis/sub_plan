package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	_ "github.com/dermaddis/sub_plan/util"
	"github.com/joho/godotenv"
)

var myClasses = []string{
	"11E05",
	"12dsp15",
}

var notFoundError = errors.New("page not found")
var parseError = errors.New("Failed parsing")

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	mainUrl := os.Getenv("mainUrl")
	if mainUrl == "" {
		log.Fatalln("Add mainUrl to .env file")
	}
	sessionId := os.Getenv("sessionId")
	if sessionId == "" {
		log.Fatalln("Add sessionId to .env file")
	}

	pages := []string{}
	for i := 1; true; i++ {
		htmlString, err := requestSubst(mainUrl, sessionId, i)
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

func requestSubst(baseUrl, sessionId string, n int) (string, error) {
	client := &http.Client{}

	url := fmt.Sprintf(baseUrl+`subst_%03d.htm`, n)
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

type EntryType int

const (
	Cancelled = 0
	Switch    = 1
	Other     = 2
)

type Entry struct {
	Texts [7]string
	Type  EntryType
	Room  string
	Class string
	Date  time.Time
	Start int
	End   int
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

		log.Println("---", texts, "---")

		hoursStr := texts[2]
		roomStr := texts[3]
		// className := texts[4]

		start, end, err := parseHours(hoursStr)
		if err != nil {
			if errors.Is(err, parseError) {
				start, end = 0, 0
			} else {
				log.Fatalln(err)
			}
		}
		log.Println("start:", start, "end:", end)

		entryType, roomData1, roomData2 := parseRoom(roomStr)
		log.Println("entryType:", entryType, "roomData:", roomData1, roomData2)
	})
}

func parseHours(s string) (int, int, error) {
	switch len(s) {
	case 1:
		res, err := strconv.Atoi(s)
		if err != nil {
			return 0, 0, parseError
		}
		return res, 0, nil
	default:
		split := strings.Split(s, " - ")
		if len(split) != 2 {
			return 0, 0, parseError
		}

		first, err := strconv.Atoi(split[0])
		if err != nil {
			return 0, 0, parseError
		}

		second, err := strconv.Atoi(split[1])
		if err != nil {
			return 0, 0, parseError
		}

		return first, second, nil
	}
}

func parseRoom(s string) (EntryType, string, string) {
	switch {
	case strings.HasSuffix(s, "---"):
		return Cancelled, "", ""
	case strings.Contains(s, "→"):
		rooms := strings.Split(s, "→")
		if len(rooms) != 2 {
			return Other, s, ""
		}
		return Switch, rooms[0], rooms[1]
	default:
		return Other, s, ""
	}
}

func compare(first string) func(other string) bool {
	return func(other string) bool {
		return first == other
	}
}
