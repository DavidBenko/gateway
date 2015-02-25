package proxy

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"

	aphttp "gateway/http"
)

type proxyRequest struct {
	Method   string `json:"method"`
	Host     string `json:"host"`
	URI      string `json:"uri"`
	Path     string `json:"path"`
	RawQuery string `json:"rawQuery"`
	Body     string `json:"body"`

	RemoteAddress string `json:"remoteAddress"`
	ContentLength int64  `json:"contentLength"`

	Headers map[string]interface{} `json:"headers"`
	Form    map[string]interface{} `json:"form"`
	Query   map[string]interface{} `json:"query"`
	Vars    map[string]string      `json:"vars"`
	Params  map[string]interface{} `json:"params"`

	ID string `json:"id"`
}

type proxyResponse struct {
	StatusCode int                    `json:"statusCode"`
	Body       string                 `json:"body"`
	Headers    map[string]interface{} `json:"headers"`
}

func proxyRequestJSON(r *http.Request, id string, vars map[string]string) (string, error) {
	request := proxyRequest{
		Method:        r.Method,
		Host:          r.Host,
		URI:           r.RequestURI,
		Path:          r.URL.Path,
		RawQuery:      r.URL.RawQuery,
		RemoteAddress: r.RemoteAddr,
		ContentLength: r.ContentLength,
		Headers:       aphttp.DesliceValues(r.Header),
		Vars:          vars,
		ID:            id,
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return "", err
	}
	request.Body = string(body)

	r.Body = ioutil.NopCloser(bytes.NewReader(body))
	if err = r.ParseForm(); err != nil {
		return "", err
	}
	request.Form = aphttp.DesliceValues(r.PostForm)

	query, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return "", err
	}
	request.Query = aphttp.DesliceValues(query)

	params := joinSlices(r.PostForm, query, aphttp.ResliceValues(vars))
	request.Params = aphttp.DesliceValues(params)

	json, err := json.Marshal(request)
	if err != nil {
		return "", err
	}
	return string(json), nil
}

func proxyResponseFromJSON(responseJSON string) (proxyResponse, error) {
	response := proxyResponse{}
	err := json.Unmarshal([]byte(responseJSON), &response)
	return response, err
}

func joinSlices(slices ...map[string][]string) map[string][]string {
	joined := make(map[string][]string)
	for _, slice := range slices {
		for key, values := range slice {
			for _, value := range values {
				if !valueInSlice(value, joined[key]) {
					joined[key] = append(joined[key], value)
				}
			}
		}
	}
	return joined
}

func valueInSlice(value string, slice []string) bool {
	for _, sliceValue := range slice {
		if value == sliceValue {
			return true
		}
	}
	return false
}
