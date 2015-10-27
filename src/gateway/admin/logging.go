package admin

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"

	"gateway/config"
	aphttp "gateway/http"
	"gateway/queue"
	"gateway/queue/mangos"
	apsql "gateway/sql"

	"github.com/blevesearch/bleve"
	"github.com/gorilla/handlers"
	elasti "github.com/mattbaird/elastigo/lib"
	"golang.org/x/net/websocket"
)

var Interceptor = newInterceptor()

type LogStreamController struct {
	BaseController
	aggregator *mangos.Broker
}

func RouteLogStream(c *LogStreamController, path string, router aphttp.Router) {
	router.Handle(path, websocket.Handler(c.logHandler))
}

type logPublisher struct {
	subscribers            []chan []byte
	subscribe, unsubscribe chan chan []byte
	write                  chan []byte
}

func newPublisher(in chan []byte) *logPublisher {
	l := &logPublisher{
		subscribers: make([]chan []byte, 8),
		subscribe:   make(chan chan []byte, 8),
		unsubscribe: make(chan chan []byte, 8),
		write:       in,
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

func (l *logPublisher) Write(p []byte) (n int, err error) {
	cp := make([]byte, len(p))
	copy(cp, p)
	l.write <- cp
	return os.Stdout.Write(p)
}

func (l *logPublisher) Subscribe() (logs <-chan []byte, unsubscribe func()) {
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

func newInterceptor() *logPublisher {
	return newPublisher(make(chan []byte, 8))
}

func newAggregator(conf config.ProxyAdmin) (*mangos.Broker, error) {
	return mangos.NewBroker(mangos.XPubXSub, mangos.TCP, conf.XPub(), conf.XSub())
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
	if api := apiIDFromPath(request); api != -1 {
		exps = append(exps, regexp.MustCompile(fmt.Sprintf(".*\\[api %v\\].*", api)))
	}
	if end := endpointIDFromPath(request); end != -1 {
		exps = append(exps, regexp.MustCompile(fmt.Sprintf(".*\\[end %v\\].*", end)))
	}
	request.ParseForm()
	if query, valid := request.Form["query"]; valid {
		exps = append(exps, regexp.MustCompile(fmt.Sprintf(".*%v.*", regexp.QuoteMeta(query[0]))))
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

func (c *LogStreamController) logHandler(ws *websocket.Conn) {
	receive, err := queue.Subscribe(
		c.conf.XPub(),
		mangos.Sub,
		mangos.SubTCP,
	)
	if err != nil {
		log.Fatal(err)
	}
	logs, e := receive.Channels()
	defer func() {
		receive.Close()
	}()
	go func() {
		for err := range e {
			log.Printf("[logging] %v", err)
		}
	}()

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

var Bleve bleve.Index

func RouteLogSearch(controller *LogSearchController, path string,
	router aphttp.Router, db *apsql.DB, conf config.ProxyAdmin) {

	routes := map[string]http.Handler{
		"GET":  read(db, controller.Search),
		"POST": read(db, controller.Search),
	}
	if conf.CORSEnabled {
		routes["OPTIONS"] = aphttp.CORSOptionsHandler([]string{"GET", "POST", "OPTIONS"})
	}

	router.Handle(path, handlers.MethodHandler(routes))
}

type LogSearchController struct {
	config.ElasticLogging
	BaseController
}

type LogSearchResult struct {
	Text string `json:"text"`
}

func (c *LogSearchController) ElasticSearch(r *http.Request) (results []LogSearchResult, httperr aphttp.Error) {
	e := elasti.NewConn()
	e.Domain = c.Domain

	queryMust := []interface{}{}
	convert := func(t string) string {
		tt, _ := time.Parse("2006-01-02T15:04:05Z", t)
		return tt.Format("2006/01/02 15:04:05") + ".000000"
	}

	if len(r.Form["start"]) == 1 || len(r.Form["end"]) == 1 {
		queryLogDate := map[string]interface{}{}
		if len(r.Form["start"]) == 1 {
			queryLogDate["gte"] = convert(r.Form["start"][0])
		}
		if len(r.Form["end"]) == 1 {
			queryLogDate["lte"] = convert(r.Form["end"][0])
		}
		queryLogDate = map[string]interface{}{
			"range": map[string]interface{}{
				"logDate": queryLogDate,
			},
		}
		queryMust = append(queryMust, queryLogDate)
	}
	account := c.accountID(r)
	queryAccount := map[string]interface{}{
		"term": map[string]interface{}{
			"account": float64(account),
		},
	}
	queryMust = append(queryMust, queryAccount)
	if api := apiIDFromPath(r); api != -1 {
		queryAPI := map[string]interface{}{
			"term": map[string]interface{}{
				"api": float64(api),
			},
		}
		queryMust = append(queryMust, queryAPI)
	}
	if endpoint := endpointIDFromPath(r); endpoint != -1 {
		queryEndpoint := map[string]interface{}{
			"term": map[string]interface{}{
				"endpoint": float64(endpoint),
			},
		}
		queryMust = append(queryMust, queryEndpoint)
	}
	if len(r.Form["query"]) == 1 {
		queryQuery := map[string]interface{}{
			"term": map[string]interface{}{
				"text": r.Form["query"][0],
			},
		}
		queryMust = append(queryMust, queryQuery)
	}
	size := 100
	if len(r.Form["limit"]) == 1 {
		sz, err := strconv.Atoi(r.Form["limit"][0])
		if err != nil {
			size = sz
		}
	}
	query := map[string]interface{}{
		"size": float64(size),
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": queryMust,
			},
		},
	}
	jsonQuery, err := json.Marshal(query)
	if err != nil {
		httperr = aphttp.NewError(err, http.StatusBadRequest)
		return
	}
	out, err := e.Search("gateway", "log", nil, jsonQuery)
	results = make([]LogSearchResult, len(out.Hits.Hits))
	for i, hit := range out.Hits.Hits {
		json.Unmarshal(*hit.Source, &results[i])
	}
	if err != nil {
		httperr = aphttp.NewError(err, http.StatusBadRequest)
		return
	}

	return
}

func (c *LogSearchController) BleveSearch(r *http.Request) (results []LogSearchResult, httperr aphttp.Error) {
	if Bleve == nil {
		httperr = aphttp.NewError(errors.New("Can't find bleve index."), http.StatusBadRequest)
		return
	}

	query := []bleve.Query{}
	if len(r.Form["start"]) == 1 || len(r.Form["end"]) == 1 {
		var start *string
		if len(r.Form["start"]) == 1 {
			start = &r.Form["start"][0]
		}
		var end *string
		if len(r.Form["end"]) == 1 {
			end = &r.Form["end"][0]
		}
		queryDate := bleve.NewDateRangeQuery(start, end)
		queryDate.SetField("logDate")
		query = append(query, queryDate)
	}
	account := c.accountID(r)
	minAccount, maxAccount := float64(account), float64(account+1)
	queryAccount := bleve.NewNumericRangeQuery(&minAccount, &maxAccount)
	queryAccount.SetField("account")
	query = append(query, queryAccount)
	if api := apiIDFromPath(r); api != -1 {
		minAPI, maxAPI := float64(api), float64(api+1)
		queryAPI := bleve.NewNumericRangeQuery(&minAPI, &maxAPI)
		queryAPI.SetField("api")
		query = append(query, queryAPI)
	}
	if endpoint := endpointIDFromPath(r); endpoint != -1 {
		minEndpoint, maxEndpoint := float64(endpoint), float64(endpoint+1)
		queryEndpoint := bleve.NewNumericRangeQuery(&minEndpoint, &maxEndpoint)
		queryEndpoint.SetField("endpoint")
		query = append(query, queryEndpoint)
	}
	if len(r.Form["query"]) == 1 {
		queryQuery := bleve.NewMatchQuery(r.Form["query"][0])
		query = append(query, queryQuery)
	}
	size := 100
	if len(r.Form["limit"]) == 1 {
		sz, err := strconv.Atoi(r.Form["limit"][0])
		if err != nil {
			size = sz
		}
	}

	search := bleve.NewSearchRequest(bleve.NewConjunctionQuery(query))
	search.Size = size
	search.Fields = []string{"text"}
	searchResults, err := Bleve.Search(search)
	if err != nil {
		httperr = aphttp.NewError(err, http.StatusBadRequest)
		return
	}
	results = make([]LogSearchResult, len(searchResults.Hits))
	for i, hit := range searchResults.Hits {
		results[i].Text = hit.Fields["text"].(string)
	}

	return
}

func (c *LogSearchController) Search(w http.ResponseWriter, r *http.Request, db *apsql.DB) aphttp.Error {
	var results []LogSearchResult
	r.ParseForm()
	if c.Domain == "" {
		var httperr aphttp.Error
		results, httperr = c.BleveSearch(r)
		if httperr != nil {
			return httperr
		}
	} else {
		var httperr aphttp.Error
		results, httperr = c.ElasticSearch(r)
		if httperr != nil {
			return httperr
		}
	}

	logs := ""
	for _, result := range results {
		logs += result.Text
	}
	result := struct {
		Logs string `json:"logs"`
	}{
		logs,
	}
	body, err := json.MarshalIndent(&result, "", "    ")
	if err != nil {
		return aphttp.NewError(err, http.StatusBadRequest)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
	return nil
}
