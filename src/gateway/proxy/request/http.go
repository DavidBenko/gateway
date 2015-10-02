package request

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	aphttp "gateway/http"
	"gateway/model"
)

// HTTPRequest encapsulates a request made over HTTP(s).
type HTTPRequest struct {
	model.HTTPRequest
	client *http.Client
}

// HTTPResponse encapsulates a response from an HTTPRequest.
type HTTPResponse struct {
	StatusCode int                    `json:"statusCode"`
	Body       string                 `json:"body"`
	Headers    map[string]interface{} `json:"headers"`
}

// CompleteURL returns the full URL including query params
func (h *HTTPRequest) CompleteURL() string {
	params := url.Values{}
	for k, v := range h.Query {
		params.Add(k, v)
	}
	return fmt.Sprintf("%s?%s", h.URL, params.Encode())
}

// Perform executes the HTTP request and returns its response.
func (h *HTTPRequest) Perform() Response {
	body := bytes.NewReader([]byte(h.Body))

	req, err := http.NewRequest(h.Method, h.CompleteURL(), body)
	if err != nil {
		context := fmt.Errorf("Error creating request from %v: %v\n", h, err)
		return NewErrorResponse(context)
	}
	aphttp.AddHeaders(req.Header, h.Headers)

	resp, err := h.client.Do(req)
	if err != nil {
		return NewErrorResponse(err)
	}

	response, err := ParseResponse(resp)
	if err != nil {
		context := fmt.Errorf("Error parsing response %v: %v\n", resp, err)
		return NewErrorResponse(context)
	}

	return response
}

// Log returns the HTTP request basics, e.g. 'GET http://www.google.com' when in server mode.
// When in dev mode the query parameters, headers, and body are also returned.
// TODO(binary132): No more "devMode", use a real logger
func (h *HTTPRequest) Log(devMode bool) string {
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("%s %s", h.Method, h.URL))
	if devMode {
		buffer.WriteString(fmt.Sprintf("\nQuery Parameters:\n"))
		for k, v := range h.Query {
			buffer.WriteString(fmt.Sprintf("    %s: %s\n", k, v))
		}
		buffer.WriteString(fmt.Sprintf("Headers:\n"))
		for k, v := range h.Headers {
			buffer.WriteString(fmt.Sprintf("    %s: %s\n", k, v))
		}
		buffer.WriteString(fmt.Sprintf("Body:\n%s", h.Body))
	}
	return buffer.String()
}

// ParseResponse takes a raw http.Response and creates an HTTPResponse.
func ParseResponse(response *http.Response) (*HTTPResponse, error) {
	r := &HTTPResponse{
		StatusCode: response.StatusCode,
		Headers:    aphttp.DesliceValues(response.Header),
	}
	var err error
	bodyReader := response.Body
	if response.Header.Get("Content-Encoding") == "gzip" {
		bodyReader, err = gzip.NewReader(bodyReader)
		if err != nil {
			return nil, err
		}
	}
	defer bodyReader.Close()

	body, err := ioutil.ReadAll(bodyReader)
	if err != nil {
		return nil, err
	}
	r.Body = string(body)

	return r, nil
}

// JSON converts this response to JSON format.
func (r *HTTPResponse) JSON() ([]byte, error) {
	return json.Marshal(&r)
}

// Log returns the status code
func (r *HTTPResponse) Log() string {
	return fmt.Sprintf("(%d)", r.StatusCode)
}

func NewHTTPRequest(client *http.Client, endpoint *model.RemoteEndpoint, data *json.RawMessage) (Request, error) {
	request := &HTTPRequest{}
	if err := json.Unmarshal(*data, request); err != nil {
		return nil, err
	}

	endpointData := &HTTPRequest{}
	if err := json.Unmarshal(endpoint.Data, endpointData); err != nil {
		return nil, err
	}
	request.updateWith(endpointData)

	if endpoint.SelectedEnvironmentData != nil {
		if err := json.Unmarshal(*endpoint.SelectedEnvironmentData, endpointData); err != nil {
			return nil, err
		}
		request.updateWith(endpointData)
	}

	if client == nil {
		return nil, errors.New("no client defined")
	}

	request.client = client

	return request, nil
}

func (r *HTTPRequest) updateWith(endpointData *HTTPRequest) {
	if endpointData.Method != "" {
		r.Method = endpointData.Method
	}
	if endpointData.URL != "" {
		r.URL = endpointData.URL
	}
	if endpointData.Body != "" {
		r.Body = endpointData.Body
	}
	for name, value := range endpointData.Query {
		if r.Query == nil {
			r.Query = make(map[string]string)
		}
		r.Query[name] = value
	}
	for name, value := range endpointData.Headers {
		if r.Headers == nil {
			r.Headers = make(map[string]interface{})
		}
		r.Headers[name] = value
	}
}
