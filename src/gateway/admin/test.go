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
	aphttp "gateway/http"
	"gateway/model"
	apsql "gateway/sql"

	"github.com/gorilla/handlers"
)

func RouteTest(controller *TestController, path string,
	router aphttp.Router, db *apsql.DB, conf config.ProxyAdmin) {

	routes := map[string]http.Handler{
		"GET": read(db, controller.Test),
	}
	if conf.CORSEnabled {
		routes["OPTIONS"] = aphttp.CORSOptionsHandler([]string{"GET", "OPTIONS"})
	}

	router.Handle(path, handlers.MethodHandler(routes))
}

type TestController struct {
	BaseController
	config.ProxyServer
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
	}
	endpoint, err := proxyEndpoint.Find(db)
	if err != nil {
		return aphttp.NewError(err, http.StatusBadRequest)
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

			testUrl := fmt.Sprintf("http://%v:%v/justapis/test%v",
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
				case "GET":
					request.URL.RawQuery = values.Encode()
				case "POST":
					if content_type == "application/x-www-form-urlencoded" {
						request.Body = ioutil.NopCloser(bytes.NewBufferString(values.Encode()))
					} else {
						request.Body = ioutil.NopCloser(bytes.NewBufferString(test.Body))
					}
				case "PUT":
					request.Body = ioutil.NopCloser(bytes.NewBufferString(test.Body))
				case "DELETE":
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
