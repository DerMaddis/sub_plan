package parser

import (
	"fmt"
	"regexp"
	"time"

	"github.com/PuerkitoBio/goquery"
)

func GetUpdatedAt(doc *goquery.Document) (time.Time, error) {
	var zeroTime time.Time

	updateTimeSelection := doc.Find(".mon_head p")
	if updateTimeSelection.Length() != 1 {
		return zeroTime, fmt.Errorf("UpdatedAt element not found %w", parseError)
	}
	updateTimeRegex := regexp.MustCompile(".*Stand: (?P<Date>.*)$")
	matches := updateTimeRegex.FindStringSubmatch(updateTimeSelection.Text())
	if len(matches) != 2 {
		return zeroTime, fmt.Errorf("UpdatedAt not matched %w", parseError)
	}
	updateTimeString := matches[1]
	updatedAt, err := time.Parse("02.01.2006 15:04", updateTimeString)
	if err != nil {
		return zeroTime, fmt.Errorf("UpdateTime not parsed %w", parseError)
	}
	return updatedAt, nil
}

