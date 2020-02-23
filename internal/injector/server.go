package injector

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"k8s.io/api/admission/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/util/json"
)

// Server represents a mutating webhook HTTP server
type Server struct {
	server   *http.Server
	decoder  runtime.Decoder
	admitter Admitter
}

// ServerOptions represent server configuration
type ServerOptions func(s *Server)

// WithPort sets a TCP port a Server will be listen on
func WithPort(port int) ServerOptions {
	return func(s *Server) {
		s.server.Addr = fmt.Sprintf(":%v", port)
	}
}

// NewServer creates a Server instance with provided options
func NewServer(admitter Admitter, opts ...ServerOptions) *Server {
	server := Server{
		server:   &http.Server{},
		decoder:  serializer.NewCodecFactory(runtime.NewScheme()).UniversalDeserializer(),
		admitter: admitter,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/mutate", server.handleMutate)

	server.server.Handler = mux

	for _, opt := range opts {
		opt(&server)
	}

	return &server
}

// Run starts listening for incoming HTTP requests
func (s *Server) Run() error {
	fmt.Printf("Listening on %s\n", s.server.Addr)
	return s.server.ListenAndServe()
}

// ServerHTTP handles an HTTP request
func (s *Server) ServerHTTP(w http.ResponseWriter, r *http.Request) {
	s.server.Handler.ServeHTTP(w, r)
}

func (s *Server) handleMutate(w http.ResponseWriter, r *http.Request) {
	if code, err := s.mutate(w, r); err == nil {
		log.Print("Successfully handled a webhook request")
	} else {
		log.Printf("Could not handle a webhook request: %v", err)
		w.WriteHeader(code)
		w.Write([]byte(err.Error()))
	}
}

func (s *Server) mutate(w http.ResponseWriter, r *http.Request) (int, error) {
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

	admissionResponse, err := s.admitter.Admit(*admissionReview.Request)
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("could not admit the requested resource: %v", err)
	}

	admissionReview.Response = &admissionResponse
	bytes, err := json.Marshal(admissionReview)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("Could not marshall the response body: %v", err)
	}

	if _, err = w.Write(bytes); err != nil {
		return http.StatusInternalServerError, fmt.Errorf("Could not write the response body: %v", err)
	}

	return http.StatusOK, nil
}
