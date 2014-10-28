package requests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	aphttp "gateway/http"
)

// HTTPRequest encapsulates a request made over HTTP(s).
type HTTPRequest struct {
	Method  string                 `json:"method"`
	URL     string                 `json:"url"`
	Body    string                 `json:"body"`
	Headers map[string]interface{} `json:"headers"`
}

// HTTPResponse encapsulates a response from an HTTPRequest.
type HTTPResponse struct {
	StatusCode int                    `json:"statusCode"`
	Body       string                 `json:"body"`
	Headers    map[string]interface{} `json:"headers"`
}

// NewHTTPRequest decodes a JSON string and returns the request.
func NewHTTPRequest(jsonReq string) (*HTTPRequest, error) {
	req := HTTPRequest{}
	err := json.Unmarshal([]byte(jsonReq), &req)
	return &req, err
}

// Perform executes the HTTP request and returns its response.
func (h *HTTPRequest) Perform(c chan<- responsePayload, index int) {
	body := bytes.NewReader([]byte(h.Body))

	req, err := http.NewRequest(h.Method, h.URL, body)
	if err != nil {
		context := fmt.Errorf("Error creating request from %v: %v\n", h, err)
		c <- responsePayload{index: index, response: NewErrorResponse(context)}
		return
	}
	aphttp.AddHeaders(req.Header, h.Headers)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		context := fmt.Errorf("Error performing request %v: %v\n", h, err)
		c <- responsePayload{index: index, response: NewErrorResponse(context)}
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

// Log returns the HTTP request basics, e.g. 'GET http://www.google.com'
func (h *HTTPRequest) Log() string {
	return fmt.Sprintf("%s %s", h.Method, h.URL)
}

// ParseResponse takes a raw http.Response and creates an HTTPResponse.
func ParseResponse(response *http.Response) (*HTTPResponse, error) {
	r := &HTTPResponse{
		StatusCode: response.StatusCode,
		Headers:    aphttp.DesliceValues(response.Header),
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
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
