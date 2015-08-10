package http

type TestResponse struct {
  Method  string        `json:"method"`
  Status  string        `json:"status"`
  Headers []*TestHeader `json:"headers"`
	Body    string        `json:"body"`
	Log     string        `json:"log"`
  Time    int64         `json:"time"`
}

type TestHeader struct {
  Name  string `json:"name"`
  Value string `json:"value"`
}
