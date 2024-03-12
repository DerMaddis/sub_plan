package parser

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Substitution struct {
}

func Parse(doc *goquery.Document) (*Substitution, error) {
    updateTimeSelection := doc.Find(".mon_head p")
    if updateTimeSelection.Length() != 1 {
        return nil, fmt.Errorf("UpdatedAt element not found %w", parseError)
    }
    updateTimeRegex := regexp.MustCompile(".*Stand: (?P<Date>.*)$")
    matches := updateTimeRegex.FindStringSubmatch(updateTimeSelection.Text())
    if len(matches) != 2 {
        return nil, fmt.Errorf("UpdatedAt not matched %w", parseError)
    }
    updateTimeString := matches[1]
    updateTime, err := time.Parse("02.01.2006 15:04", updateTimeString)
    if err != nil {
        return nil, fmt.Errorf("UpdateTime not parsed %w", parseError)
    }
    log.Println(updateTime)

	titleSelection := doc.Find(".mon_title")
	if titleSelection.Length() == 0 {
		return nil, fmt.Errorf("Title not found %w", parseError)
	}

	var _err *error

	doc.Find("tr.list").Slice(1, goquery.ToEnd).Each(func(i int, s *goquery.Selection) {
		if _err != nil {
			// skip iteration since one of the last ones errored
			return
		}

		data := s.Find("td.list")
		if data.Length() != 7 {
            err := fmt.Errorf("Not 7 entries in table row %w", parseError)
            _err = &err
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
				_err = &err
				return
			}
		}
		log.Println("start:", start, "end:", end)

		roomData := parseRoom(roomStr)
		switch d := roomData.(type) {
		case RoomCancelledData:
			log.Println("cancelled")
		case RoomSwitchData:
			log.Println("switch from", d.Room1, "to", d.Room2)
		case RoomOtherData:
			log.Println("other", d.Data)
		}
	})

	if _err != nil {
		return nil, fmt.Errorf("Parse: %w", *_err)
	}

	return &Substitution{}, nil
}

