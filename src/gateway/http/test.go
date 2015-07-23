package http

type TestResponse struct {
  Method string `json:"method"`
  Status string `json:"status"`
	Body   string `json:"body"`
	Log    string `json:"log"`
}
