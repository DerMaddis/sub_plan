package parser

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Substitution struct {
	UpdatedAt time.Time // when this subst was put into the system
	Date      time.Time // which day this subst is concerning
	RoomData  interface{}
	Start     int
	End       int
}

func (s Substitution) String() string {
    return fmt.Sprintf("Start: %d End: %d RoomData: %s", s.Start, s.End, s.RoomData)
}

func Parse(doc *goquery.Document) ([]*Substitution, error) {
	updatedAt, err := GetUpdatedAt(doc)
	if err != nil {
		return nil, fmt.Errorf("UpdatedAt not found %w", err)
	}

	date, err := GetDate(doc)

	entries := doc.Find("tr.list").Slice(1, goquery.ToEnd)
	if entries.Length() == 0 {
		return nil, fmt.Errorf("No entries found %w", parseError)
	}

	var wg sync.WaitGroup
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

	// fill Date field
	for i := 0; i < len(results); i++ {
		results[i].Date = date
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
