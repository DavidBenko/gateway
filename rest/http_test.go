package rest

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockResource struct {
	index    string
	indexErr error
}

func (m *mockResource) Name() string {
	return ""
}

func (m *mockResource) Index() (resources interface{}, err error) {
	return m.index, m.indexErr
}

func (m *mockResource) Create(data interface{}) (resource interface{}, err error) {
	return nil, nil
}
func (m *mockResource) Show(id interface{}) (resource interface{}, err error) {
	return nil, nil
}
func (m *mockResource) Update(id interface{}, data interface{}) (resource interface{}, err error) {
	return nil, nil
}
func (m *mockResource) Delete(id interface{}) error {
	return nil
}

func TestIndex(t *testing.T) {
	w := testIndex("foo", nil)
	if w.Code != http.StatusOK {
		t.Error("Expected index to have status code 200")
	}
	if w.Body.String() != "foo\n" {
		t.Error("Expected index body to be what was returned by the resource")
	}
}

func TestIndexError(t *testing.T) {
	w := testIndex("", fmt.Errorf("error fetching"))
	if w.Code != http.StatusMethodNotAllowed {
		t.Error("Expected index to have status code 405")
	}
	if w.Body.String() == "An error occurred: error fetching\n" {
		t.Error("Expected index body to contain error")
	}
}

func testIndex(val string, err error) *httptest.ResponseRecorder {
	res := HTTPResource{Resource: &mockResource{
		index:    val,
		indexErr: err,
	}}
	handler := res.IndexHandler()

	req, err := http.NewRequest("GET", "http://example.com/resources", nil)
	if err != nil {
		log.Fatal(err)
	}

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	return w
}
