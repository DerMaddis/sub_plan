package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/dermaddis/sub_plan/parser"

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

	if len(pages) == 0 {
		log.Fatalln("No pages found")
	}

	for _, page := range pages {

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(page))
		if err != nil {
			log.Fatalln(err)
		}

		substitution, err := parser.Parse(doc)
		if err != nil {
			log.Fatalln(err)
		}

		log.Println(substitution)
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
