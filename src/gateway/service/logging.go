package service

import (
	"bytes"
	"log"
	"regexp"
	"strconv"

	"gateway/admin"
	"gateway/config"

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
	TIME_REGEXP     = "^[0-9]{4}/[0-9]{2}/[0-9]{2} [0-9]{2}:[0-9]{2}:[0-9]{2}[.][0-9]{6}"
	ACCOUNT_REGEXP  = ".*\\[act ([0-9]{1,})\\].*"
	API_REGEXP      = ".*\\[api ([0-9]{1,})\\].*"
	ENDPOINT_REGEXP = ".*\\[end ([0-9]{1,})\\].*"
)

type Message struct {
	Text     string `json:"text"`
	LogDate  string `json:"logDate"`
	Account  int    `json:"account"`
	API      int    `json:"api"`
	Endpoint int    `json:"endpoint"`
}

func NewMessage(message string) *Message {
	logDate := regexp.MustCompile(TIME_REGEXP).FindString(message)
	var properties [3]int
	for i, re := range []string{ACCOUNT_REGEXP, API_REGEXP, ENDPOINT_REGEXP} {
		matches := regexp.MustCompile(re).FindStringSubmatch(message)
		if len(matches) == 2 {
			properties[i], _ = strconv.Atoi(matches[1])
		} else {
			properties[i] = -1
		}
	}

	return &Message{
		Text:     message,
		LogDate:  logDate,
		Account:  properties[0],
		API:      properties[1],
		Endpoint: properties[2],
	}
}

func ElasticLoggingService(conf config.ElasticLogging) {
	if conf.Domain == "" {
		return
	}

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
				log.Fatal(err)
			}
		}
		buffer, newline := &bytes.Buffer{}, false
		elastic := func(b byte) {
			buffer.WriteByte(b)
			if b == '\n' {
				_, err = c.Index("gateway", "log", "", nil, NewMessage(buffer.String()))
				if err != nil {
					log.Printf("[elastic] %v", err)
				}
				buffer.Reset()
			}
		}

		for _, b := range <-logs {
			if newline {
				elastic(b)
			} else if b == '\n' {
				newline = true
			}
		}
		for input := range logs {
			for _, b := range input {
				elastic(b)
			}
		}
	}()
}
