package service

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	logger "log"
	"regexp"
	"strconv"
	"time"

	"gateway/admin"
	"gateway/config"
	"gateway/queue"
	"gateway/queue/mangos"

	"github.com/blevesearch/bleve"
	elasti "github.com/mattbaird/elastigo/lib"
)

const (
	MESSAGE_MAPPING = `{
    "log": {
      "properties": {
        "logDate": {
          "type": "date",
          "format": "YYYY/MM/dd HH:mm:ss.SSSSSS"
        }
      }
    }
  }`
	LOG_TIME_FORMAT = "2006/01/02 15:04:05"
	TIME_REGEXP     = "^([0-9]{4}/[0-9]{2}/[0-9]{2} [0-9]{2}:[0-9]{2}:[0-9]{2})[.]([0-9]{6})"
	ACCOUNT_REGEXP  = ".*\\[act ([0-9]{1,})\\].*"
	API_REGEXP      = ".*\\[api ([0-9]{1,})\\].*"
	ENDPOINT_REGEXP = ".*\\[end ([0-9]{1,})\\].*"
)

type ElasticMessage struct {
	Text     string `json:"text"`
	LogDate  string `json:"logDate"`
	Account  int    `json:"account"`
	API      int    `json:"api"`
	Endpoint int    `json:"endpoint"`
}

func NewElasticMessage(message string) *ElasticMessage {
	logDate := regexp.MustCompile(TIME_REGEXP).FindString(message)
	if logDate == "" {
		return nil
	}

	var properties [3]int
	for i, re := range []string{ACCOUNT_REGEXP, API_REGEXP, ENDPOINT_REGEXP} {
		matches := regexp.MustCompile(re).FindStringSubmatch(message)
		if len(matches) == 2 {
			properties[i], _ = strconv.Atoi(matches[1])
		} else {
			properties[i] = -1
		}
	}

	return &ElasticMessage{
		Text:     message,
		LogDate:  logDate,
		Account:  properties[0],
		API:      properties[1],
		Endpoint: properties[2],
	}
}

func processLogs(logs <-chan []byte, add func(message string)) {
	buffer, newline := &bytes.Buffer{}, false
	process := func(b byte) {
		buffer.WriteByte(b)
		if b == '\n' {
			add(buffer.String())
			buffer.Reset()
		}
	}

	for _, b := range <-logs {
		if newline {
			process(b)
		} else if b == '\n' {
			newline = true
		}
	}
	for input := range logs {
		for _, b := range input {
			process(b)
		}
	}
}

func ElasticLoggingService(conf config.ElasticLogging) {
	if conf.Domain == "" {
		return
	}

	logger.Printf("%s Starting Elastic logging service", config.System)

	go func() {
		logs, unsubscribe := admin.Interceptor.Subscribe()
		defer unsubscribe()

		c := elasti.NewConn()
		c.Domain = conf.Domain
		c.Username = conf.Username
		c.Password = conf.Password

		_, err := c.CreateIndex("gateway")
		if err == nil {
			err = c.PutMappingFromJSON("gateway", "log", []byte(MESSAGE_MAPPING))
			if err != nil {
				logger.Fatal(err)
			}
		}
		add := func(message string) {
			elasticMessage := NewElasticMessage(message)
			if elasticMessage == nil {
				return
			}
			_, err = c.Index("gateway", "log", "", nil, elasticMessage)
			if err != nil {
				logger.Printf("[elastic] %v", err)
			}
		}
		processLogs(logs, add)
	}()
}

type BleveMessage struct {
	Text     string    `json:"text"`
	LogDate  time.Time `json:"logDate"`
	Account  float64   `json:"account"`
	API      float64   `json:"api"`
	Endpoint float64   `json:"endpoint"`
}

func NewBleveMessage(message string) *BleveMessage {
	logDate := regexp.MustCompile(TIME_REGEXP).FindStringSubmatch(message)
	var date time.Time
	if len(logDate) == 3 {
		var err error
		date, err = time.Parse(LOG_TIME_FORMAT, logDate[1])
		if err != nil {
			logger.Fatal(err)
		}
		seconds, err := strconv.Atoi(logDate[2])
		if err != nil {
			logger.Fatal(err)
		}
		date = date.Add(time.Duration(seconds) * time.Microsecond)
	} else {
		return nil
	}
	var properties [3]float64
	for i, re := range []string{ACCOUNT_REGEXP, API_REGEXP, ENDPOINT_REGEXP} {
		matches := regexp.MustCompile(re).FindStringSubmatch(message)
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
	}
}

func (m *BleveMessage) Type() string {
	return "message"
}

func (m *BleveMessage) Id() string {
	data, err := json.Marshal(m)
	if err != nil {
		logger.Fatal(err)
	}
	return fmt.Sprintf("%x", sha1.Sum(data))
}

func BleveLoggingService(conf config.BleveLogging) {
	if conf.File == "" {
		return
	}

	logger.Printf("%s Starting Bleve logging service", config.System)

	mapping := bleve.NewIndexMapping()
	index, err := bleve.New(conf.File, mapping)
	if err != nil {
		index, err = bleve.Open(conf.File)
		if err != nil {
			logger.Fatal(err)
		}
	}
	admin.Bleve = index

	indexer := make(chan *BleveMessage, 8192)
	go func() {
		for bleveMessage := range indexer {
			err := index.Index(bleveMessage.Id(), bleveMessage)
			if err != nil {
				logger.Printf("[bleve] %v", err)
			}
		}
	}()

	go func() {
		logs, unsubscribe := admin.Interceptor.Subscribe()
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
					logger.Printf("[bleve-delete] %v", err)
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

func LogPublishingService(conf config.ProxyAdmin) {

	logger.Printf("%s Starting log publisher", config.System)

	go func() {
		logs, unsubscribe := admin.Interceptor.Subscribe()
		defer unsubscribe()

		send, err := queue.Publish(
			conf.XSub(),
			mangos.Pub(true),
			mangos.PubTCP,
		)
		if err != nil {
			logger.Fatal(err)
		}
		c, e := send.Channels()
		defer func() {
			send.Close()
		}()
		go func() {
			for err := range e {
				logger.Printf("[logging] %v", err)
			}
		}()

		add := func(message string) {
			c <- []byte(message)
		}
		processLogs(logs, add)
	}()
}
