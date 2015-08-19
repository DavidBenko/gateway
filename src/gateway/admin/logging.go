package admin

import (
	"bytes"
	"fmt"
	"os"
	"regexp"

	aphttp "gateway/http"

	"golang.org/x/net/websocket"
)

var Interceptor = newInterceptor()

func RouteLogging(path string, router aphttp.Router) {
	router.Handle(path, websocket.Handler(logHandler))
}

type logInterceptor struct {
	subscribers            []chan []byte
	subscribe, unsubscribe chan chan []byte
	write                  chan []byte
}

func newInterceptor() *logInterceptor {
	l := &logInterceptor{
		subscribers: make([]chan []byte, 8),
		subscribe:   make(chan chan []byte, 8),
		unsubscribe: make(chan chan []byte, 8),
		write:       make(chan []byte, 8),
	}
	go func() {
		for {
			select {
			case s := <-l.subscribe:
				found := false
				for i, j := range l.subscribers {
					if j == nil {
						l.subscribers[i] = s
						found = true
						break
					}
				}
				if !found {
					l.subscribers = append(l.subscribers, s)
				}
			case u := <-l.unsubscribe:
				for i, j := range l.subscribers {
					if j == u {
						l.subscribers[i] = nil
						close(u)
						break
					}
				}
			case buffer := <-l.write:
				for _, j := range l.subscribers {
					if j != nil {
						j <- buffer
					}
				}
			}
		}
	}()
	return l
}

func (l *logInterceptor) Write(p []byte) (n int, err error) {
	cp := make([]byte, len(p))
	copy(cp, p)
	l.write <- cp
	return os.Stdout.Write(p)
}

func (l *logInterceptor) Subscribe() (logs <-chan []byte, unsubscribe func()) {
	_logs := make(chan []byte, 8)
	l.subscribe <- _logs
	logs = _logs
	unsubscribe = func() {
		go func() {
			for _ = range logs {
				//noop
			}
		}()
		l.unsubscribe <- _logs
	}
	return
}

func makeFilter(ws *websocket.Conn) func(b byte) bool {
	request := ws.Request()
	exps := []*regexp.Regexp{regexp.MustCompile(".*\\[proxy\\].*")}
	act := int64(-1)
	if requestSession != nil {
		session := requestSession(request)
		accountID := session.Values[accountIDKey]
		if accountID, valid := accountID.(int64); valid {
			act = accountID
		}
	}
	if act >= 0 {
		exps = append(exps, regexp.MustCompile(fmt.Sprintf(".*\\[act %d\\].*", act)))
	}
	request.ParseForm()
	if api, valid := request.Form["api"]; valid {
		exps = append(exps, regexp.MustCompile(fmt.Sprintf(".*\\[api %v\\].*", api)))
	}
	if end, valid := request.Form["end"]; valid {
		exps = append(exps, regexp.MustCompile(fmt.Sprintf(".*\\[end %v\\].*", end)))
	}
	buffer := &bytes.Buffer{}
	return func(b byte) bool {
		buffer.WriteByte(b)
		if b == '\n' {
			defer buffer.Reset()
			matches, b := true, buffer.Bytes()
			for _, exp := range exps {
				if !exp.Match(b) {
					matches = false
					break
				}
			}
			if matches {
				_, err := ws.Write(b)
				return err != nil
			}
		}
		return false
	}
}

func logHandler(ws *websocket.Conn) {
	logs, unsubscribe := Interceptor.Subscribe()
	defer unsubscribe()
	filter, newline := makeFilter(ws), false
	for _, b := range <-logs {
		if newline {
			if filter(b) {
				return
			}
		} else if b == '\n' {
			newline = true
		}
	}
	for input := range logs {
		for _, b := range input {
			if filter(b) {
				return
			}
		}
	}
}
