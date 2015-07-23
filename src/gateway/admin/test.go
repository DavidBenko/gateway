package admin

import (
  "bytes"
  "errors"
  "encoding/json"
	"fmt"
  "io/ioutil"
  "net/http"
  "net/url"

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

  endpoint, err := model.FindProxyEndpointForAPIIDAndAccountID(db, endpointID, apiID, accountID)
  if err != nil {
    return aphttp.NewError(err, http.StatusBadRequest)
  }

  hosts, err := model.AllHostsForAPIIDAndAccountID(db, apiID, accountID)
  if err != nil {
    return aphttp.NewError(err, http.StatusBadRequest)
  }

  if len(hosts) == 0 {
    return aphttp.NewError(errors.New("A host needs to be defined."), http.StatusBadRequest)
  }

  var responses []*aphttp.TestResponse
  addResponse := func(method string, response *http.Response) aphttp.Error {
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
    testResponse.Status = response.Status

    responses = append(responses, testResponse)

    return nil
  }

  for _, test := range endpoint.Tests {
    if test.ID == testID {
      methods, err := test.GetMethods()
      if err != nil {
        return aphttp.NewError(err, http.StatusBadRequest)
      }

      testUrl := fmt.Sprintf("http://%v:%v/justapis/test%v",
        hosts[0].Hostname, c.ProxyServer.Port, test.Route)
      for _, method := range methods {
        switch method {
        case "GET":
          testUrl := testUrl
          if len(test.Pairs) > 0 {
            values := url.Values{}
            for _, pair := range test.Pairs {
              values.Add(pair.Key, pair.Value)
            }
            testUrl += "?" + values.Encode()
          }

          response, err := http.Get(testUrl)
          if err != nil {
            return aphttp.NewError(err, http.StatusBadRequest)
          }

          if err := addResponse(method, response); err != nil {
            return err
          }
        case "POST":
          if len(test.Body) > 0 {
            body := bytes.NewBufferString(test.Body)

            response, err := http.Post(testUrl, "application/json", body)
            if err != nil {
              return aphttp.NewError(err, http.StatusBadRequest)
            }

            if err := addResponse(method, response); err != nil {
              return err
            }
          } else {
            values := url.Values{}
            for _, pair := range test.Pairs {
              values.Add(pair.Key, pair.Value)
            }

            response, err := http.PostForm(testUrl, values)
            if err != nil {
              return aphttp.NewError(err, http.StatusBadRequest)
            }

            if err := addResponse(method, response); err != nil {
              return err
            }
          }
        case "PUT":
          client, body := &http.Client{}, bytes.NewBufferString(test.Body)
          request, err := http.NewRequest(method, testUrl, body)
          if err != nil {
            return aphttp.NewError(err, http.StatusBadRequest)
          }

          response, err := client.Do(request)
          fmt.Println(response)
          if err != nil {
            return aphttp.NewError(err, http.StatusBadRequest)
          }

          if err := addResponse(method, response); err != nil {
            return err
          }
        case "DELETE":
          client := &http.Client{}
          request, err := http.NewRequest(method, testUrl, nil)
          if err != nil {
            return aphttp.NewError(err, http.StatusBadRequest)
          }

          response, err := client.Do(request)
          if err != nil {
            return aphttp.NewError(err, http.StatusBadRequest)
          }

          if err := addResponse(method, response); err != nil {
            return err
          }
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
