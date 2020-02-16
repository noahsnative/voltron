package injector

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithPort(t *testing.T) {
	t.Run("Should set the server address", func(t *testing.T) {
		sut := New(WithPort(8080))
		assert.Equal(t, ":8080", sut.server.Addr)
	})
}

func TestHandleMutate(t *testing.T) {
	t.Run("Should only allow POST requests", func(t *testing.T) {
		cases := []struct {
			Method             string
			ExpectedStatusCode int
		}{
			{http.MethodGet, http.StatusMethodNotAllowed},
			{http.MethodHead, http.StatusMethodNotAllowed},
			{http.MethodPut, http.StatusMethodNotAllowed},
			{http.MethodDelete, http.StatusMethodNotAllowed},
			{http.MethodOptions, http.StatusMethodNotAllowed},
			{http.MethodPost, http.StatusOK},
		}

		for _, c := range cases {
			sut := New()

			request := httptest.NewRequest(c.Method, "/mutate", nil)
			recorder := httptest.NewRecorder()

			sut.ServerHTTP(recorder, request)

			assert.Equal(t, c.ExpectedStatusCode, recorder.Result().StatusCode)
		}
	})
}
