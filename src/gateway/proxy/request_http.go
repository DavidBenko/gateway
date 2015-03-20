package proxy

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	aphttp "gateway/http"
)

// HTTPRequest encapsulates a request made over HTTP(s).
type HTTPRequest struct {
	Method  string                 `json:"method"`
	URL     string                 `json:"url"`
	Body    string                 `json:"body"`
	Headers map[string]interface{} `json:"headers"`
	Query   map[string]string      `json:"query"`
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
func (h *HTTPRequest) Perform(s *Server, c chan<- responsePayload, index int) {
	body := bytes.NewReader([]byte(h.Body))

	req, err := http.NewRequest(h.Method, h.CompleteURL(), body)
	if err != nil {
		context := fmt.Errorf("Error creating request from %v: %v\n", h, err)
		c <- responsePayload{index: index, response: NewErrorResponse(context)}
		return
	}
	aphttp.AddHeaders(req.Header, h.Headers)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		c <- responsePayload{index: index, response: NewErrorResponse(err)}
		return
	}

	response, err := ParseResponse(resp)
	if err != nil {
		context := fmt.Errorf("Error parsing response %v: %v\n", resp, err)
		c <- responsePayload{index: index, response: NewErrorResponse(context)}
		return
	}

	c <- responsePayload{index: index, response: response}
}

// Log returns the HTTP request basics, e.g. 'GET http://www.google.com' when in server mode.
// When in dev mode the query parameters, headers, and body are also returned.
func (h *HTTPRequest) Log(s *Server) string {
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("%s %s", h.Method, h.URL))
	if s.devMode {
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
