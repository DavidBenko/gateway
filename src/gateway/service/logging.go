package service

import (
	"bytes"
	"regexp"

	"gateway/admin"
	"gateway/config"
	"gateway/logreport"
	"gateway/queue"
	"gateway/queue/mangos"
)

const LogTimeFormat = "2006/01/02 15:04:05"

var (
	TimeRegexp     = regexp.MustCompile("^([0-9]{4}/[0-9]{2}/[0-9]{2} [0-9]{2}:[0-9]{2}:[0-9]{2})[.]([0-9]{6})")
	AccountRegexp  = regexp.MustCompile(".*\\[act ([0-9]{1,})\\].*")
	APIRegexp      = regexp.MustCompile(".*\\[api ([0-9]{1,})\\].*")
	EndpointRegexp = regexp.MustCompile(".*\\[end ([0-9]{1,})\\].*")
	TimerRegexp    = regexp.MustCompile(".*\\[timer ([0-9]{1,})\\].*")
)

func processLogs(logs <-chan []byte, add func(message string)) {
	buffer := &bytes.Buffer{}
	for input := range logs {
		for _, b := range input {
			buffer.WriteByte(b)
			if b == '\n' {
				add(buffer.String())
				buffer.Reset()
			}
		}
	}
}

func LogPublishingService(conf config.ProxyAdmin) {

	logreport.Printf("%s Starting log publisher", config.System)

	go func() {
		logs, unsubscribe := admin.Interceptor.Subscribe("LogPublishingService")
		defer unsubscribe()

		send, err := queue.Publish(
			conf.XSub(),
			mangos.Pub(true),
			mangos.PubTCP,
		)
		if err != nil {
			logreport.Fatal(err)
		}
		c, e := send.Channels()
		defer func() {
			send.Close()
		}()
		go func() {
			for err := range e {
				logreport.Printf("[logging] %v", err)
			}
		}()

		add := func(message string) {
			c <- []byte(message)
		}
		processLogs(logs, add)
	}()
}
