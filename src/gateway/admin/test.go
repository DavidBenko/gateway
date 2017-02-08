package admin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"gateway/config"
	"gateway/core"
	aphttp "gateway/http"
	"gateway/logreport"
	"gateway/model"
	"gateway/push"
	apsql "gateway/sql"

	"github.com/gorilla/handlers"
	"github.com/nanoscaleio/surgemq/message"
)

func RouteTest(controller *TestController, path string,
	router aphttp.Router, db *apsql.DB, conf config.ProxyAdmin) {

	routes := map[string]http.Handler{
		"GET":  read(db, controller.Test),
		"POST": read(db, controller.Test),
	}
	if conf.CORSEnabled {
		routes["OPTIONS"] = aphttp.CORSOptionsHandler([]string{"GET", "POST", "OPTIONS"})
	}

	router.Handle(path, handlers.MethodHandler(routes))
}

type TestController struct {
	BaseController
	config.ProxyServer
	*core.Core
}

type TestResults struct {
	Results []*aphttp.TestResponse `json:"results"`
}

func (c *TestController) Test(w http.ResponseWriter, r *http.Request, db *apsql.DB) aphttp.Error {
	accountID, apiID, endpointID, testID := c.accountID(r), apiIDFromPath(r), endpointIDFromPath(r), testIDFromPath(r)

	proxyEndpoint := model.ProxyEndpoint{
		AccountID: accountID,
		APIID:     apiID,
		ID:        endpointID,
		Type:      model.ProxyEndpointTypeHTTP,
	}
	endpoint, err := proxyEndpoint.Find(db)
	if err != nil {
		return aphttp.NewError(err, http.StatusBadRequest)
	}

	for _, test := range endpoint.Tests {
		if test.ID == testID && test.Channels {
			return c.TestChannel(w, r, db, endpoint, test)
		}
	}

	selectedHost := "127.0.0.1"
	if c.Host != "" {
		selectedHost = c.Host
	}

	var responses []*aphttp.TestResponse
	addResponse := func(time int64, method string, response *http.Response) aphttp.Error {
		defer response.Body.Close()
		testResponse := &aphttp.TestResponse{}

		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return aphttp.NewError(err, http.StatusBadRequest)
		}

		if len(body) > 0 {
			err = json.Unmarshal(body, testResponse)
			if err != nil {
				testResponse.Body = string(body)
			}
		}

		testResponse.Method = method
		for name, values := range response.Header {
			for _, value := range values {
				if name == "Content-Length" {
					value = fmt.Sprintf("%v", len(testResponse.Body))
				}
				header := &aphttp.TestHeader{Name: name, Value: value}
				testResponse.Headers = append(testResponse.Headers, header)
			}
		}
		testResponse.Time = (time + 5e5) / 1e6

		responses = append(responses, testResponse)

		return nil
	}

	for _, test := range endpoint.Tests {
		if test.ID == testID {
			methods, err := test.GetMethods()
			if err != nil {
				return aphttp.NewError(err, http.StatusBadRequest)
			}

			paths := make(map[string]string)
			for _, pair := range test.Pairs {
				if pair.Type == model.PairTypePath {
					paths[pair.Key] = pair.Value
				}
			}
			route, path := bytes.Buffer{}, []rune(test.Route)
			for i := 0; i < len(path); i++ {
				if path[i] != '{' {
					route.WriteRune(path[i])
				} else {
					key := bytes.Buffer{}
					for i++; path[i] != '}'; i++ {
						if path[i] != ' ' && path[i] != '\t' {
							key.WriteRune(path[i])
						}
					}
					route.WriteString(paths[key.String()])
				}
			}

			testUrl := fmt.Sprintf("http://%v:%v/nanoscale/test%v",
				selectedHost, c.ProxyServer.Port, route.String())
			for _, method := range methods {
				client, values := &http.Client{}, url.Values{}
				request, err := http.NewRequest(method, testUrl, nil)
				if err != nil {
					return aphttp.NewError(err, http.StatusBadRequest)
				}

				request.Host = fmt.Sprintf("%v.example.com", apiID)

				content_type := ""
				for _, pair := range test.Pairs {
					switch pair.Type {
					case model.PairTypeGet:
						values.Add(pair.Key, pair.Value)
					case model.PairTypePost:
					case model.PairTypeHeader:
						request.Header.Set(pair.Key, pair.Value)
						if pair.Key == "Content-Type" {
							content_type = pair.Value
						}
					}
				}

				switch method {
				case model.ProxyEndpointTestMethodGet:
					request.URL.RawQuery = values.Encode()
				case model.ProxyEndpointTestMethodPost:
					if content_type == "application/x-www-form-urlencoded" {
						request.Body = ioutil.NopCloser(bytes.NewBufferString(values.Encode()))
					} else {
						request.Body = ioutil.NopCloser(bytes.NewBufferString(test.Body))
					}
				case model.ProxyEndpointTestMethodPut:
					request.Body = ioutil.NopCloser(bytes.NewBufferString(test.Body))
				case model.ProxyEndpointTestMethodDelete:
					// empty
				}

				start := time.Now()
				response, err := client.Do(request)
				elapsed := time.Since(start)
				if err != nil {
					return aphttp.NewError(err, http.StatusBadRequest)
				}

				if err := addResponse(elapsed.Nanoseconds(), method, response); err != nil {
					return err
				}
			}

			break
		}
	}

	body, err := json.MarshalIndent(&TestResults{responses}, "", "    ")
	if err != nil {
		return aphttp.NewError(err, http.StatusBadRequest)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(body)

	return nil
}

type mqttRequest struct {
	Method        string `json:"method"`
	Body          string `json:"body"`
	ContentLength int64  `json:"contentLength"`

	Headers map[string]interface{} `json:"headers"`
	Form    map[string]interface{} `json:"form"`
	Query   map[string]interface{} `json:"query"`
	Vars    map[string]string      `json:"vars"`
	Params  map[string]interface{} `json:"params"`
}

type dummyAddr struct {
}

func (a *dummyAddr) Network() string {
	return "localhost"
}

func (a *dummyAddr) String() string {
	return "localhost"
}

func (c *TestController) TestChannel(w http.ResponseWriter, r *http.Request, db *apsql.DB,
	endpoint *model.ProxyEndpoint, test *model.ProxyEndpointTest) aphttp.Error {
	channel := &model.ProxyEndpointChannel{
		AccountID:       endpoint.AccountID,
		APIID:           endpoint.APIID,
		ProxyEndpointID: endpoint.ID,
		ID:              *test.ChannelID,
	}
	var err error
	channel, err = channel.Find(db)
	if err != nil {
		return aphttp.NewError(err, http.StatusBadRequest)
	}

	remoteEndpoint, err := model.FindRemoteEndpointForAPIIDAndAccountID(db, channel.RemoteEndpointID,
		channel.APIID, channel.AccountID)
	if err != nil {
		return aphttp.NewError(err, http.StatusBadRequest)
	}

	context := &push.Context{
		RemoteEndpoint: remoteEndpoint,
	}

	logs := &bytes.Buffer{}
	logPrint := logreport.PrintfCopier(logs)

	methods, err := test.GetMethods()
	if err != nil {
		return aphttp.NewError(err, http.StatusBadRequest)
	}
	request := &mqttRequest{
		Method:        methods[0],
		Body:          test.Body,
		ContentLength: int64(len(test.Body)),
		Headers:       make(map[string]interface{}),
		Form:          make(map[string]interface{}),
		Query:         make(map[string]interface{}),
		Vars:          make(map[string]string),
		Params:        make(map[string]interface{}),
	}
	for _, pair := range test.Pairs {
		switch pair.Type {
		case model.PairTypeGet:
			request.Query[pair.Key] = pair.Value
			request.Params[pair.Key] = pair.Value
		case model.PairTypePost:
			request.Form[pair.Key] = pair.Value
			request.Params[pair.Key] = pair.Value
		case model.PairTypeHeader:
			request.Headers[pair.Key] = pair.Value
		case model.PairTypePath:
			request.Vars[pair.Key] = pair.Value
			request.Params[pair.Key] = pair.Value
		}
	}

	requestBytes, err := json.Marshal(request)
	if err != nil {
		return aphttp.NewError(err, http.StatusBadRequest)
	}

	requestMessage := message.NewPublishMessage()
	requestMessage.SetTopic([]byte("/" + channel.Name))
	requestMessage.SetPayload(requestBytes)

	start := time.Now()
	var responseMessage *message.PublishMessage
	onpub := func(msg *message.PublishMessage) error {
		responseMessage = msg
		return nil
	}
	err = c.ExecuteMQTT(context, logPrint, requestMessage, &dummyAddr{}, onpub)
	if err != nil {
		return aphttp.NewError(err, http.StatusBadRequest)
	}
	elapsed := time.Since(start)

	response := &core.ProxyResponse{}
	err = json.Unmarshal(responseMessage.Payload(), response)
	if err != nil {
		return aphttp.NewError(err, http.StatusBadRequest)
	}

	testResponse := &aphttp.TestResponse{
		Method: methods[0],
		Status: fmt.Sprintf("%v", response.StatusCode),
		Body:   response.Body,
		Log:    logs.String(),
		Time:   (elapsed.Nanoseconds() + 5e5) / 1e6,
	}
	for name, value := range response.Headers {
		header := &aphttp.TestHeader{Name: name, Value: fmt.Sprintf("%v", value)}
		testResponse.Headers = append(testResponse.Headers, header)
	}

	body, err := json.MarshalIndent(&TestResults{[]*aphttp.TestResponse{testResponse}}, "", "    ")
	if err != nil {
		return aphttp.NewError(err, http.StatusBadRequest)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(body)

	return nil
}
