package integration

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type HttpHelper struct {
	httpClient *http.Client
	cookies    map[string]*http.Cookie
}

func NewHTTPHelper() *HttpHelper {
	helper := new(HttpHelper)
	helper.httpClient = &http.Client{}
	helper.cookies = map[string]*http.Cookie{}
	return helper
}

func (h *HttpHelper) Get(url string) (int, map[string][]string, string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, nil, "", err
	}
	return h.Do(req)
}

func (h *HttpHelper) Post(url, body string) (int, map[string][]string, string, error) {
	req, err := http.NewRequest("POST", url, strings.NewReader(body))
	if err != nil {
		return 0, nil, "", err
	}
	return h.Do(req)
}

func (h *HttpHelper) Do(req *http.Request) (int, map[string][]string, string, error) {
	for _, c := range h.cookies {
		req.AddCookie(c)
	}
	resp, err := h.httpClient.Do(req)
	if err != nil {
		return 0, nil, "", err
	}

	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, "", fmt.Errorf("Error reading response body due to: %v", err)
	}

	for _, c := range resp.Cookies() {
		h.cookies[c.Name] = c
	}

	return resp.StatusCode, resp.Header, string(respBody), nil
}
