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

func TestFailsInvalidAdmissionRequestKind(t *testing.T) {
	request := v1beta1.AdmissionRequest{
		Kind: metav1.GroupVersionKind{Group: "autoscaling", Version: "v1", Kind: "Scale"},
	}
	sut := NewAdmitter()
	_, err := sut.Admit(request)
	assert.Error(t, err)
}

func TestSucceedsIfValidAdmissionRequestKind(t *testing.T) {
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

func TestFailsIfInvalidAdmissionRequestObject(t *testing.T) {
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

func TestFailsIfPodHasNoContainers(t *testing.T) {
	var pod corev1.Pod
	pod.Kind = "Pod"
	pod.APIVersion = "v1"
	podBytes, _ := json.Marshal(pod)

	request := v1beta1.AdmissionRequest{
		Object: runtime.RawExtension{Raw: podBytes},
		Kind:   metav1.GroupVersionKind{Group: "core", Version: "v1", Kind: "Pod"},
	}

	sut := NewAdmitter()
	_, err := sut.Admit(request)
	assert.Error(t, err)
}

func TestPatchesInitContainers(t *testing.T) {
	tests := []struct {
		summary         string
		pod             corev1.Pod
		expectedPatches []string
	}{
		{
			summary: "No existing init containers",
			pod:     corev1.Pod{},
			expectedPatches: []string{
				`{"op":"add","path":"/spec/initContainers","value":[{"image":"nginx:latest","name":"nginx","resources":{}}]}`,
			},
		},
		{
			summary: "One existing init container",
			pod: corev1.Pod{
				Spec: corev1.PodSpec{
					InitContainers: []corev1.Container{
						corev1.Container{Image: "foo:latest"},
					},
				},
			},
			expectedPatches: []string{
				`{"op":"add","path":"/spec/initContainers/1","value":{"image":"nginx:latest","name":"nginx","resources":{}}}`,
			},
		},
		{
			summary: "Multiple existing init containers",
			pod: corev1.Pod{
				Spec: corev1.PodSpec{
					InitContainers: []corev1.Container{
						corev1.Container{Image: "foo:latest"},
						corev1.Container{Image: "bar:latest"},
					},
				},
			},
			expectedPatches: []string{
				`{"op":"add","path":"/spec/initContainers/2","value":{"image":"nginx:latest","name":"nginx","resources":{}}}`,
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

			actualPatches := string(response.Patch)
			for _, patch := range test.expectedPatches {
				assert.Contains(t, actualPatches, patch)
			}
		})
	}
}

func ensureValid(pod corev1.Pod) corev1.Pod {
	pod.Kind = "Pod"
	pod.APIVersion = "v1"
	pod.Spec.Containers = append(pod.Spec.Containers, corev1.Container{Image: "foo:latest"})
	return pod
}
