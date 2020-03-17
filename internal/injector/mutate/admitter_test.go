package mutate

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestInvalidKindInAdmissionRequest(t *testing.T) {
	request := v1beta1.AdmissionRequest{
		Kind: metav1.GroupVersionKind{Group: "autoscaling", Version: "v1", Kind: "Scale"},
	}
	sut := NewAdmitter()
	_, err := sut.Admit(request)
	assert.Error(t, err)
}

func TestValidKindInAdmissionRequest(t *testing.T) {
	request := v1beta1.AdmissionRequest{
		Object: runtime.RawExtension{Raw: []byte{}},
		Kind:   metav1.GroupVersionKind{Group: "core", Version: "v1", Kind: "Pod"},
	}
	sut := NewAdmitter()
	_, err := sut.Admit(request)
	assert.NoError(t, err)
}

func TestInvalidObjectInAdmissionRequest(t *testing.T) {
	tests := []struct {
		summary string
		object  []byte
	}{
		{summary: "Nil object", object: nil},
	}

	for _, test := range tests {
		t.Run(test.summary, func(t *testing.T) {
			request := v1beta1.AdmissionRequest{
				Object: runtime.RawExtension{Raw: test.object},
				Kind:   metav1.GroupVersionKind{Group: "core", Version: "v1", Kind: "Pod"},
			}
			sut := NewAdmitter()
			_, err := sut.Admit(request)
			assert.Error(t, err)
		})
	}
}

func TestValidAdmissionRequestAddsInitContainer(t *testing.T) {
	nginxContainer := corev1.Container{
		Image: "nginx:latest",
	}

	tests := []struct {
		summary         string
		pod             corev1.Pod
		expectedPatches []patchOperation
	}{
		{
			summary: "No existing init containers",
			pod:     corev1.Pod{},
			expectedPatches: []patchOperation{
				patchOperation{
					Op:    "add",
					Path:  "/spec/initContainers",
					Value: nginxContainer,
				},
			}},
	}

	for _, test := range tests {
		t.Run(test.summary, func(t *testing.T) {
			podBytes, _ := json.Marshal(test.pod)
			request := v1beta1.AdmissionRequest{
				Object: runtime.RawExtension{Raw: podBytes},
				Kind:   metav1.GroupVersionKind{Group: "core", Version: "v1", Kind: "Pod"},
			}
			sut := NewAdmitter()
			response, _ := sut.Admit(request)

			actual := response.Patch
			expected, _ := json.Marshal(test.expectedPatches)
			assert.JSONEq(t, string(actual), string(expected))
		})
	}
}
