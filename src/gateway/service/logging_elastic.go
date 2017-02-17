package service

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"gateway/admin"
	"gateway/config"
	"gateway/errors/report"
	"gateway/logreport"

	elasti "github.com/mattbaird/elastigo/lib"
)

const MessageMapping = `{
  "log": {
    "properties": {
      "logDate": {
        "type": "date",
        "format": "YYYY/MM/dd HH:mm:ss.SSSSSS"
      }
    }
  }
}`

type ElasticMessage struct {
	Text     string `json:"text"`
	LogDate  string `json:"logDate"`
	Account  int    `json:"account"`
	API      int    `json:"api"`
	Endpoint int    `json:"endpoint"`
	Timer    int    `json:"timer"`
}

func NewElasticMessage(message string) *ElasticMessage {
	logDate := TimeRegexp.FindString(message)
	if logDate == "" {
		return nil
	}

	var properties [4]int
	for i, re := range []*regexp.Regexp{AccountRegexp, APIRegexp, EndpointRegexp, TimerRegexp} {
		matches := re.FindStringSubmatch(message)
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
		Timer:    properties[3],
	}
}

func ElasticLoggingService(conf config.Configuration) {
	if conf.Elastic.Url == "" {
		return
	}

	logreport.Printf("%s Starting Elastic logging service", config.System)

	go func() {
		logs, unsubscribe := admin.Interceptor.Subscribe("ElasticLoggingService")
		defer unsubscribe()

		c := elasti.NewConn()
		c.SetFromUrl(conf.Elastic.Url)

		_, err := c.CreateIndex("gateway")
		if err == nil {
			err = c.PutMappingFromJSON("gateway", "log", []byte(MessageMapping))
			if err != nil {
				logreport.Fatal(err)
			}
		}
		add := func(message string) {
			elasticMessage := NewElasticMessage(message)
			if elasticMessage == nil {
				return
			}
			_, err = c.Index("gateway", "log", "", nil, elasticMessage)
			if err != nil {
				// print error to stdout, and then report directly to our error reporter.
				// we don't want to log (or use logreport which both logs and reports)
				// because we will potentially compound the problem which is being
				// encountered
				fmt.Printf("[elastic] %v\n", err)
				report.Error(err, nil)
			}
		}
		processLogs(logs, add)
	}()

	if !conf.Jobs {
		return
	}

	deleteTicker := time.NewTicker(24 * time.Hour)
	go func() {
		for _ = range deleteTicker.C {
			days := time.Duration(conf.Elastic.DeleteAfter)
			end := time.Now().Add(-days*24*time.Hour).Format("2006/01/02 15:04:05") + ".000000"
			c := elasti.NewConn()
			c.SetFromUrl(conf.Elastic.Url)
			query := map[string]interface{}{
				"query": map[string]interface{}{
					"range": map[string]interface{}{
						"logDate": map[string]interface{}{
							"lte": end,
						},
					},
				},
			}
			jsonQuery, err := json.Marshal(query)
			if err != nil {
				logreport.Printf("[elastic-delete] %v", err)
				continue
			}
			_, err = c.DeleteByQuery([]string{"gateway"}, []string{"log"}, nil, jsonQuery)
			if err != nil {
				logreport.Printf("[elastic-delete] %v", err)
			}
		}
	}()
}
