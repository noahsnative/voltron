package mutate

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"k8s.io/api/admission/v1beta1"
)

type Handler struct {
	admitter Admitter
}

func New() *Handler {
	return &Handler{
		admitter: noOpAdmitter{},
	}
}

func (h *Handler) Mutate(w http.ResponseWriter, r *http.Request) {
	if code, err := h.mutate(w, r); err == nil {
		log.Print("Successfully handled a webhook request")
	} else {
		log.Printf("Could not handle a webhook request: %v", err)
		w.WriteHeader(code)
		if _, err = w.Write([]byte(err.Error())); err != nil {
			log.Print("Could not write the error response body")
		}
	}
}

func (h *Handler) mutate(w http.ResponseWriter, r *http.Request) (int, error) {
	if r.Method != http.MethodPost {
		return http.StatusMethodNotAllowed, fmt.Errorf("invalid method %s, only POST requests are allowed", r.Method)
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("could not read the request body: %v", err)
	}

	var admissionReview v1beta1.AdmissionReview
	if err := json.Unmarshal(body, &admissionReview); err != nil {
		return http.StatusBadRequest, fmt.Errorf("could not parse the request body: %v", err)
	}

	if admissionReview.Request == nil {
		return http.StatusBadRequest, fmt.Errorf("mailformed admission review: request is nil")
	}

	admissionResponse, err := h.admitter.Admit(*admissionReview.Request)
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("could not admit the requested resource: %v", err)
	}

	admissionReview.Response = &admissionResponse
	bytes, err := json.Marshal(admissionReview)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("could not marshall the response body: %v", err)
	}

	if _, err = w.Write(bytes); err != nil {
		return http.StatusInternalServerError, fmt.Errorf("could not write the response body: %v", err)
	}

	return http.StatusOK, nil
}
