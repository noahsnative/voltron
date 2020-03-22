package mutate

import (
	"encoding/json"
	"strings"
	"testing"
	"unicode"

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
	pod := corev1.Pod{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{{Image: "foo:latest"}},
		},
	}
	podBytes, _ := json.Marshal(ensureKind(pod))

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

func TestReturnsJSONPatch(t *testing.T) {
	tests := []struct {
		summary         string
		pod             corev1.Pod
		expectedPatches []string
	}{
		{
			summary: "Patches init containers if pod had no init containers",
			pod: corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{Image: "foo:latest"}},
				},
			},
			expectedPatches: []string{
				`{
					"op":"add",
					"path":"/spec/initContainers",
					"value":[{"image":"nginx:latest","name":"nginx","resources":{},"volumeMounts":[{"mountPath":"/bin/voltron","name":"voltron-env"}]}]
				}`,
			},
		},
		{
			summary: "Patches init containers if pod had a signle init container",
			pod: corev1.Pod{
				Spec: corev1.PodSpec{
					Containers:     []corev1.Container{{Image: "foo:latest"}},
					InitContainers: []corev1.Container{{Image: "foo:latest"}},
				},
			},
			expectedPatches: []string{
				`{
					"op":"add",
					"path":"/spec/initContainers/1",
					"value":{"image":"nginx:latest","name":"nginx","resources":{},"volumeMounts":[{"mountPath":"/bin/voltron","name":"voltron-env"}]}
				}`,
			},
		},
		{
			summary: "Patches init containers if pod had multiple init containers",
			pod: corev1.Pod{
				Spec: corev1.PodSpec{
					Containers:     []corev1.Container{{Image: "foo:latest"}},
					InitContainers: []corev1.Container{{Image: "foo:latest"}, {Image: "bar:latest"}},
				},
			},
			expectedPatches: []string{
				`{
					"op":"add",
					"path":"/spec/initContainers/2",
					"value":{"image":"nginx:latest","name":"nginx","resources":{},"volumeMounts":[{"mountPath":"/bin/voltron","name":"voltron-env"}]}
				}`,
			},
		},
		{
			summary: "Patches volumes if pod had no volumes",
			pod: corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{Image: "foo:latest"}},
				},
			},
			expectedPatches: []string{
				`{"op":"add","path":"/spec/volumes","value":[{"emptyDir":{"medium":"Memory"},"name":"voltron-env"}]}`,
			},
		},
		{
			summary: "Patches volumes if pod had a signle volume",
			pod: corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{Image: "foo:latest"}},
					Volumes:    []corev1.Volume{{Name: "foo"}},
				},
			},
			expectedPatches: []string{
				`{"op":"add","path":"/spec/volumes/1","value":{"emptyDir":{"medium":"Memory"},"name":"voltron-env"}}`,
			},
		},
		{
			summary: "Patches volumes if pod had multiple volumes",
			pod: corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{Image: "foo:latest"}},
					Volumes:    []corev1.Volume{{Name: "foo"}, {Name: "bar"}},
				},
			},
			expectedPatches: []string{
				`{"op":"add","path":"/spec/volumes/2","value":{"emptyDir":{"medium":"Memory"},"name":"voltron-env"}}`,
			},
		},
		{
			summary: "Patches first container's volume mounts if it had no volume mounts",
			pod: corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{Image: "foo:latest"}, {Image: "bar:latest"}},
					Volumes:    []corev1.Volume{{Name: "foo"}, {Name: "bar"}},
				},
			},
			expectedPatches: []string{
				`{
					"op":"add",
					"path":"/spec/containers/0/volumeMounts",
					"value":[{"mountPath":"/bin/voltron","name":"voltron-env"}]
				}`,
			},
		},
		{
			summary: "Patches first container's volume mounts if it had volume mounts",
			pod: corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						corev1.Container{Image: "foo:latest", VolumeMounts: []corev1.VolumeMount{{Name: "foo"}}},
						corev1.Container{Image: "bar:latest", VolumeMounts: []corev1.VolumeMount{{Name: "bar"}}},
					},
					Volumes: []corev1.Volume{{Name: "foo"}, {Name: "bar"}},
				},
			},
			expectedPatches: []string{
				`{
					"op":"add",
					"path":"/spec/containers/0/volumeMounts/1",
					"value":{"mountPath":"/bin/voltron","name":"voltron-env"}
				}`,
			},
		},
		{
			summary: "Patches first container's command if it had no command",
			pod: corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						corev1.Container{Image: "foo:latest"},
						corev1.Container{Image: "bar:latest"},
					},
					Volumes: []corev1.Volume{{Name: "foo"}, {Name: "bar"}},
				},
			},
			expectedPatches: []string{
				`{
					"op":"add",
					"path":"/spec/containers/0/command",
					"value":["/bin/voltron/injector"]
				}`,
			},
		},
		{
			summary: "Patches first container's command if it had a command",
			pod: corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						corev1.Container{Image: "foo:latest", Command: []string{"sh"}},
						corev1.Container{Image: "bar:latest"},
					},
					Volumes: []corev1.Volume{{Name: "foo"}, {Name: "bar"}},
				},
			},
			expectedPatches: []string{
				`{
					"op":"replace",
					"path":"/spec/containers/0/command/0",
					"value":"/bin/voltron/injector"
				}`,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.summary, func(t *testing.T) {
			podBytes, _ := json.Marshal(ensureKind(test.pod))
			request := v1beta1.AdmissionRequest{
				Object: runtime.RawExtension{Raw: podBytes},
				Kind:   metav1.GroupVersionKind{Group: "core", Version: "v1", Kind: "Pod"},
			}

			sut := NewAdmitter()
			response, err := sut.Admit(request)
			assert.NoError(t, err)

			actualPatches := string(response.Patch)
			for _, patch := range test.expectedPatches {
				assert.Contains(t, actualPatches, compress(patch))
			}
		})
	}
}

func ensureKind(pod corev1.Pod) corev1.Pod {
	pod.Kind = "Pod"
	pod.APIVersion = "v1"
	return pod
}

func compress(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}

		return r
	}, s)
}
