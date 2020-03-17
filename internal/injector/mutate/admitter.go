package mutate

import (
	"encoding/json"
	"errors"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/api/admission/v1beta1"
)

// Admitter validates an admission request and admits it, possibly mutating
type Admitter interface {
	Admit(v1beta1.AdmissionRequest) (v1beta1.AdmissionResponse, error)
}

type mutator struct{}

type patchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

var (
	expectedKind = metav1.GroupVersionKind{Group: "core", Version: "v1", Kind: "Pod"}
)

func (a mutator) Admit(request v1beta1.AdmissionRequest) (v1beta1.AdmissionResponse, error) {
	var response v1beta1.AdmissionResponse

	if request.Kind != expectedKind {
		return response, errors.New("invalid kind: expected Pod")
	}

	if request.Object.Raw == nil {
		return response, errors.New("invalid object")
	}

	nginxContainer := corev1.Container{
		Image: "nginx:latest",
	}

	patches := []patchOperation{
		patchOperation{
			Op:    "add",
			Path:  "/spec/initContainers",
			Value: nginxContainer,
		},
	}

	b, err := json.Marshal(patches)
	if err != nil {
		return response, fmt.Errorf("tba: %v", err)
	}

	response.Patch = b
	return response, nil
}

//NewAdmitter returns a Admitter
func NewAdmitter() Admitter {
	return mutator{}
}
