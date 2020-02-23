package injector

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/noahsnative/voltron/internal/injector/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"k8s.io/api/admission/v1beta1"
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
		sut := NewServer(&mocks.Admitter{}, WithPort(8080))
		assert.Equal(t, ":8080", sut.server.Addr)
	})
}

func TestFailIfNonPostRequest(t *testing.T) {
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
		admitterStub := &mocks.Admitter{}
		admitterStub.On("Admit", mock.Anything).Return(v1beta1.AdmissionResponse{}, nil)
		sut := NewServer(admitterStub)

		request := httptest.NewRequest(c.Method, "/mutate", strings.NewReader(validAdmissionReview))
		recorder := httptest.NewRecorder()

		sut.ServerHTTP(recorder, request)

		assert.Equal(t, c.ExpectedStatusCode, recorder.Result().StatusCode)
	}
}

func TestFailIfInvalidRequestBody(t *testing.T) {
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
			sut := NewServer(&mocks.Admitter{})

			request := httptest.NewRequest(http.MethodPost, "/mutate", strings.NewReader(c.RequestBody))
			recorder := httptest.NewRecorder()

			sut.ServerHTTP(recorder, request)

			assert.Equal(t, http.StatusBadRequest, recorder.Result().StatusCode)
		})
	}
}

func TestCallAdmitterIfValidRequestBody(t *testing.T) {
	admitterMock := &mocks.Admitter{}
	admitterMock.On("Admit", mock.Anything).Return(v1beta1.AdmissionResponse{}, nil)
	sut := NewServer(admitterMock)

	request := httptest.NewRequest(http.MethodPost, "/mutate", strings.NewReader(validAdmissionReview))
	recorder := httptest.NewRecorder()

	sut.ServerHTTP(recorder, request)

	admitterMock.AssertNumberOfCalls(t, "Admit", 1)
}

func TestFailIfAdmitterFails(t *testing.T) {
	admitterMock := &mocks.Admitter{}
	admitterMock.
		On("Admit", mock.Anything).
		Return(v1beta1.AdmissionResponse{}, errors.New("admission failed"))
	sut := NewServer(admitterMock)

	request := httptest.NewRequest(http.MethodPost, "/mutate", strings.NewReader(validAdmissionReview))
	recorder := httptest.NewRecorder()

	sut.ServerHTTP(recorder, request)

	assert.Equal(t, http.StatusBadRequest, recorder.Result().StatusCode)
}

func TestSucceedIfAdmitterSucceeds(t *testing.T) {
	admitterMock := &mocks.Admitter{}
	var admissionResponse v1beta1.AdmissionResponse
	admitterMock.
		On("Admit", mock.Anything).
		Return(admissionResponse, nil)
	sut := NewServer(admitterMock)

	request := httptest.NewRequest(http.MethodPost, "/mutate", strings.NewReader(validAdmissionReview))
	recorder := httptest.NewRecorder()

	sut.ServerHTTP(recorder, request)

	assert.Equal(t, http.StatusOK, recorder.Result().StatusCode)

	var actual v1beta1.AdmissionReview
	assert.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &actual))
	assert.True(t, assert.ObjectsAreEqual(admissionResponse, *actual.Response))
}
