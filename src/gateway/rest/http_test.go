package rest

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type mockResource struct {
	result interface{}
	err    error
}

func (m *mockResource) Name() string {
	return ""
}
func (m *mockResource) Index() (resources interface{}, err error) {
	return m.result, m.err
}
func (m *mockResource) Create(data interface{}) (resource interface{}, err error) {
	return m.result, m.err
}
func (m *mockResource) Show(id interface{}) (resource interface{}, err error) {
	return m.result, m.err
}
func (m *mockResource) Update(id interface{}, data interface{}) (resource interface{}, err error) {
	return m.result, m.err
}
func (m *mockResource) Delete(id interface{}) error {
	return m.err
}

func TestIndex(t *testing.T) {
	w := testAction("foo", nil, func(res HTTPResource) http.Handler {
		return res.IndexHandler()
	})
	checkResponse(t, w, http.StatusOK, "foo\n")
}

func TestIndexError(t *testing.T) {
	w := testAction("", fmt.Errorf("error fetching"), func(res HTTPResource) http.Handler {
		return res.IndexHandler()
	})
	checkResponse(t, w, http.StatusInternalServerError, "error fetching\n")
}

func TestCreate(t *testing.T) {
	w := testAction("foo", nil, func(res HTTPResource) http.Handler {
		return res.CreateHandler()
	})
	checkResponse(t, w, http.StatusOK, "foo\n")
}

func TestCreateError(t *testing.T) {
	w := testAction("", fmt.Errorf("error fetching"), func(res HTTPResource) http.Handler {
		return res.CreateHandler()
	})
	checkResponse(t, w, http.StatusBadRequest, "error fetching\n")
}

func TestShow(t *testing.T) {
	w := testAction("foo", nil, func(res HTTPResource) http.Handler {
		return res.ShowHandler()
	})
	checkResponse(t, w, http.StatusOK, "foo\n")
}

func TestShowError(t *testing.T) {
	w := testAction("", fmt.Errorf("error fetching"), func(res HTTPResource) http.Handler {
		return res.ShowHandler()
	})
	checkResponse(t, w, http.StatusNotFound, "error fetching\n")
}

func TestUpdate(t *testing.T) {
	w := testAction("foo", nil, func(res HTTPResource) http.Handler {
		return res.UpdateHandler()
	})
	checkResponse(t, w, http.StatusOK, "foo\n")
}

func TestUpdateError(t *testing.T) {
	w := testAction("", fmt.Errorf("error fetching"), func(res HTTPResource) http.Handler {
		return res.UpdateHandler()
	})
	checkResponse(t, w, http.StatusBadRequest, "error fetching\n")
}

func TestDelete(t *testing.T) {
	w := testAction("", nil, func(res HTTPResource) http.Handler {
		return res.DeleteHandler()
	})
	checkResponse(t, w, http.StatusOK, "")
}

func TestDeleteError(t *testing.T) {
	w := testAction("", fmt.Errorf("error fetching"), func(res HTTPResource) http.Handler {
		return res.DeleteHandler()
	})
	checkResponse(t, w, http.StatusBadRequest, "error fetching\n")
}

func testAction(val string, err error, handlerFunc func(HTTPResource) http.Handler) *httptest.ResponseRecorder {
	handler := handlerFunc(HTTPResource{Resource: &mockResource{
		result: val,
		err:    err,
	}})

	req, err := http.NewRequest("GET", "http://example.com/resources/1", bytes.NewBuffer([]byte{}))
	if err != nil {
		log.Fatal(err)
	}

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	return w
}

func checkResponse(t *testing.T, w *httptest.ResponseRecorder, code int, body string) {
	if w.Code != code {
		t.Errorf("Expected status code %d; got %d", code, w.Code)
	}
	if !strings.Contains(w.Body.String(), body) {
		t.Errorf("Expected body to contain '%s'; got '%s'", body, w.Body.String())
	}
}
