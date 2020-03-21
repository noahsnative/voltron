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
	var pod corev1.Pod
	podBytes, _ := json.Marshal(ensureValid(pod))

	request := v1beta1.AdmissionRequest{
		Object: runtime.RawExtension{Raw: podBytes},
		Kind:   metav1.GroupVersionKind{Group: "core", Version: "v1", Kind: "Pod"},
	}

	sut := NewAdmitter()
	_, err := sut.Admit(request)
	assert.NoError(t, err)
}

func TestInvalidObjectInAdmissionRequest(t *testing.T) {
	var pv corev1.PersistentVolume
	pv.Kind = "PersistentVolume"
	pv.APIVersion = "v1"
	pvBytes, _ := json.Marshal(pv)

	tests := []struct {
		summary string
		object  []byte
	}{
		{summary: "Nil object", object: nil},
		{summary: "Not a pod", object: pvBytes},
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
			},
		},
		{
			summary: "One existing init container",
			pod: corev1.Pod{
				Spec: corev1.PodSpec{
					InitContainers: []corev1.Container{nginxContainer},
				},
			},
			expectedPatches: []patchOperation{
				patchOperation{
					Op:    "add",
					Path:  "/spec/initContainers/-",
					Value: nginxContainer,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.summary, func(t *testing.T) {
			podBytes, _ := json.Marshal(ensureValid(test.pod))
			request := v1beta1.AdmissionRequest{
				Object: runtime.RawExtension{Raw: podBytes},
				Kind:   metav1.GroupVersionKind{Group: "core", Version: "v1", Kind: "Pod"},
			}

			sut := NewAdmitter()
			response, err := sut.Admit(request)
			assert.NoError(t, err)

			actual := response.Patch
			expected, _ := json.Marshal(test.expectedPatches)
			assert.JSONEq(t, string(actual), string(expected))
		})
	}
}

func ensureValid(pod corev1.Pod) corev1.Pod {
	pod.Kind = "Pod"
	pod.APIVersion = "v1"
	return pod
}
