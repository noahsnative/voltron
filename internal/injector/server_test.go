package injector

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/noahsnative/voltron/internal/injector/mocks"
	"github.com/stretchr/testify/assert"
)

var (
	validAdmissionReview = `{
		"apiVersion": "admission.k8s.io/v1",
		"kind": "AdmissionReview",
		"request": {
		  "uid": "705ab4f5-6393-11e8-b7cc-42010a800002",
		  "kind": {"group":"autoscaling","version":"v1","kind":"Scale"},
		  "resource": {"group":"apps","version":"v1","resource":"deployments"},
		  "subResource": "scale",
		  "requestKind": {"group":"autoscaling","version":"v1","kind":"Scale"},
		  "requestResource": {"group":"apps","version":"v1","resource":"deployments"},
		  "requestSubResource": "scale",
		  "name": "my-deployment",
		  "namespace": "my-namespace",
		  "operation": "UPDATE",
		  "userInfo": {
			"username": "admin",
			"uid": "014fbff9a07c",
			"groups": ["system:authenticated","my-admin-group"],
			"extra": {
			  "some-key":["some-value1", "some-value2"]
			}
		  },
	  
		  "object": {"apiVersion":"autoscaling/v1","kind":"Scale"},
		  "oldObject": {"apiVersion":"autoscaling/v1","kind":"Scale"},
		  "options": {"apiVersion":"meta.k8s.io/v1","kind":"UpdateOptions"},
		  "dryRun": false
		}
	}`
)

func TestWithPort(t *testing.T) {
	t.Run("Should set the server address", func(t *testing.T) {
		sut := New(&mocks.Admitter{}, WithPort(8080))
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
			sut := New(&mocks.Admitter{})

			request := httptest.NewRequest(c.Method, "/mutate", strings.NewReader(validAdmissionReview))
			recorder := httptest.NewRecorder()

			sut.ServerHTTP(recorder, request)

			assert.Equal(t, c.ExpectedStatusCode, recorder.Result().StatusCode)
		}
	})

	t.Run("Should return BadRequest if invalid request body", func(t *testing.T) {
		cases := []struct {
			Summary     string
			RequestBody string
		}{
			{Summary: "Empty", RequestBody: ""},
			{Summary: "Plain text", RequestBody: "not a JSON"},
			{Summary: "Not an admission review", RequestBody: `{"foo":"bar"}`},
		}

		for _, c := range cases {
			t.Run(c.Summary, func(t *testing.T) {
				sut := New(&mocks.Admitter{})

				request := httptest.NewRequest(http.MethodPost, "/mutate", strings.NewReader(c.RequestBody))
				recorder := httptest.NewRecorder()

				sut.ServerHTTP(recorder, request)

				assert.Equal(t, http.StatusBadRequest, recorder.Result().StatusCode)
			})
		}
	})
}
