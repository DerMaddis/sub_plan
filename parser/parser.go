package parser

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Substitution struct {
	UpdatedAt time.Time
	RoomData  interface{}
	Start     int
	End       int
}

func Parse(doc *goquery.Document) ([]*Substitution, error) {
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
	updatedAt, err := time.Parse("02.01.2006 15:04", updateTimeString)
	if err != nil {
		return nil, fmt.Errorf("UpdateTime not parsed %w", parseError)
	}

	titleSelection := doc.Find(".mon_title")
	if titleSelection.Length() == 0 {
		return nil, fmt.Errorf("Title not found %w", parseError)
	}

	var wg sync.WaitGroup

	entries := doc.Find("tr.list").Slice(1, goquery.ToEnd)
	if entries.Length() == 0 {
		return nil, fmt.Errorf("No entries found %w", parseError)
	}
	substChan := make(chan *Substitution, entries.Length())
	errChan := make(chan error, entries.Length())

	doc.Find("tr.list").Slice(1, goquery.ToEnd).Each(func(i int, s *goquery.Selection) {
		wg.Add(1)
		go parseEntry(s, substChan, errChan, &wg)
	})

	go func() {
		wg.Wait()
		close(substChan)
		close(errChan)
	}()

	results := []*Substitution{}
	for subst := range substChan {
		results = append(results, subst)
	}

	// fill UpdatedAt field
	for i := 0; i < len(results); i++ {
		results[i].UpdatedAt = updatedAt
	}

	if len(errChan) != 0 {
		log.Println("--- Errors")
		for err := range errChan {
			log.Println(err)
		}
		log.Println("--- Errors end")
	}

	return results, nil
}

func parseEntry(s *goquery.Selection, substChan chan<- *Substitution, errChan chan<- error, wg *sync.WaitGroup) {
	defer wg.Done()
	data := s.Find("td.list")
	if data.Length() != 7 {
		err := fmt.Errorf("Not 7 entries in table row %w", parseError)
		errChan <- err
		return
	}

	texts := make([]string, 0, 7)
	data.Each(func(i int, s *goquery.Selection) {
		texts = append(texts, s.Text())
	})

	hoursStr := texts[2]
	roomStr := texts[3]
	// className := texts[4]

	start, end, err := parseHours(hoursStr)
	if err != nil {
		if errors.Is(err, parseError) {
			start, end = 0, 0
		} else {
			errChan <- err
			return
		}
	}

	roomData := parseRoom(roomStr)

	subst := &Substitution{
		Start:    start,
		End:      end,
		RoomData: roomData,
	}
	substChan <- subst
}
