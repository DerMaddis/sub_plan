package parser

import (
	"fmt"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

func GetDate(doc *goquery.Document) (time.Time, error) {
	var zeroTime time.Time

	titleSelection := doc.Find(".mon_title")
	if titleSelection.Length() == 0 {
		return zeroTime, fmt.Errorf("GetDate %w", parseError)
	}
	split := strings.Split(titleSelection.Text(), " ")
	if len(split) != 4 {
		return zeroTime, fmt.Errorf("GetDate %w", parseError)
	}
    
	dateStr := split[0]
	date, err := time.Parse("2.1.2006", dateStr)
	if err != nil {
		return zeroTime, fmt.Errorf("GetDate %w %w", err, parseError)
	}

	return date, nil
}

