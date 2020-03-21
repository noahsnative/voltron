package mutate

import (
	"encoding/json"
	"errors"
	"fmt"

	"k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
)

// Admitter validates an admission request and admits it, possibly mutating
type Admitter interface {
	Admit(v1beta1.AdmissionRequest) (v1beta1.AdmissionResponse, error)
}

type mutator struct {
	decoder runtime.Decoder
}

type patchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

var (
	expectedKind = metav1.GroupVersionKind{Group: "core", Version: "v1", Kind: "Pod"}
)

func (m mutator) Admit(request v1beta1.AdmissionRequest) (v1beta1.AdmissionResponse, error) {
	var response v1beta1.AdmissionResponse

	pod, err := m.extractPod(&request)
	if err != nil {
		return response, err
	}

	var patches []patchOperation
	patches = append(patches, patchInitContainers(pod))

	patchesBytes, err := json.Marshal(patches)
	if err != nil {
		return response, fmt.Errorf("tba: %v", err)
	}

	response.Patch = patchesBytes
	return response, nil
}

func (m mutator) extractPod(request *v1beta1.AdmissionRequest) (*corev1.Pod, error) {
	if request.Kind != expectedKind {
		return nil, errors.New("invalid kind: expected a pod")
	}

	object, _, err := m.decoder.Decode(request.Object.Raw, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("could not parse a pod: %v", err)
	}

	pod, ok := object.(*corev1.Pod)
	if !ok {
		return nil, errors.New("invalid object type: not a pod")
	}

	return pod, nil
}

func patchInitContainers(pod *corev1.Pod) patchOperation {
	path := "/spec/initContainers"
	if len(pod.Spec.InitContainers) > 0 {
		path += "/-"
	}

	nginxContainer := corev1.Container{
		Image: "nginx:latest",
	}

	return patchOperation{
		Op:    "add",
		Path:  path,
		Value: nginxContainer,
	}
}

//NewAdmitter returns a Admitter
func NewAdmitter() Admitter {
	return mutator{
		decoder: scheme.Codecs.UniversalDeserializer(),
	}
}
