package service

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"gateway/admin"
	"gateway/config"
	"gateway/logreport"

	"github.com/blevesearch/bleve"
)

type BleveMessage struct {
	Text     string    `json:"text"`
	LogDate  time.Time `json:"logDate"`
	Account  float64   `json:"account"`
	API      float64   `json:"api"`
	Endpoint float64   `json:"endpoint"`
	Timer    float64   `json:"timer"`
}

func NewBleveMessage(message string) *BleveMessage {
	logDate := TimeRegexp.FindStringSubmatch(message)
	var date time.Time
	if len(logDate) == 3 {
		var err error
		date, err = time.Parse(LogTimeFormat, logDate[1])
		if err != nil {
			logreport.Fatal(err)
		}
		seconds, err := strconv.Atoi(logDate[2])
		if err != nil {
			logreport.Fatal(err)
		}
		date = date.Add(time.Duration(seconds) * time.Microsecond)
	} else {
		return nil
	}
	var properties [4]float64
	for i, re := range []*regexp.Regexp{AccountRegexp, APIRegexp, EndpointRegexp, TimerRegexp} {
		matches := re.FindStringSubmatch(message)
		if len(matches) == 2 {
			val, _ := strconv.Atoi(matches[1])
			properties[i] = float64(val)
		} else {
			properties[i] = float64(-1)
		}
	}

	return &BleveMessage{
		Text:     message,
		LogDate:  date,
		Account:  properties[0],
		API:      properties[1],
		Endpoint: properties[2],
		Timer:    properties[3],
	}
}

func (m *BleveMessage) Type() string {
	return "message"
}

func (m *BleveMessage) Id() string {
	data, err := json.Marshal(m)
	if err != nil {
		logreport.Fatal(err)
	}
	return fmt.Sprintf("%x", sha1.Sum(data))
}

func BleveLoggingService(conf config.BleveLogging) {
	if conf.File == "" {
		return
	}

	logreport.Printf("%s Starting Bleve logging service", config.System)

	mapping := bleve.NewIndexMapping()
	index, err := bleve.New(conf.File, mapping)
	if err != nil {
		index, err = bleve.Open(conf.File)
		if err != nil {
			logreport.Fatal(err)
		}
	}
	admin.Bleve = index

	indexer := make(chan *BleveMessage, 8192)
	go func() {
		for bleveMessage := range indexer {
			err := index.Index(bleveMessage.Id(), bleveMessage)
			if err != nil {
				logreport.Printf("[bleve] %v", err)
			}
		}
	}()

	go func() {
		logs, unsubscribe := admin.Interceptor.Subscribe("BleveLoggingService")
		defer unsubscribe()

		add := func(message string) {
			bleveMessage := NewBleveMessage(message)
			if bleveMessage == nil {
				return
			}
			indexer <- bleveMessage
		}
		processLogs(logs, add)
	}()

	deleteTicker := time.NewTicker(24 * time.Hour)
	go func() {
		for _ = range deleteTicker.C {
			days := time.Duration(conf.DeleteAfter)
			end := time.Now().Add(-days * 24 * time.Hour).Format("2006-01-02T15:04:05Z")
			for {
				query := bleve.NewDateRangeQuery(nil, &end)
				query.SetField("logDate")
				search := bleve.NewSearchRequest(query)
				search.Size = 1024
				searchResults, err := index.Search(search)
				if err != nil {
					logreport.Printf("[bleve-delete] %v", err)
				}
				if len(searchResults.Hits) == 0 {
					break
				}
				batch := index.NewBatch()
				for _, hit := range searchResults.Hits {
					batch.Delete(hit.ID)
				}
				index.Batch(batch)
			}
		}
	}()
}
