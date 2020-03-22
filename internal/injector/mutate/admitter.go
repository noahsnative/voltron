package mutate

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/appscode/jsonpatch"
	"k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

// Admitter validates an admission request and admits it, possibly mutating
type Admitter interface {
	Admit(v1beta1.AdmissionRequest) (v1beta1.AdmissionResponse, error)
}

type mutator struct {
	decoder runtime.Decoder
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

	if len(pod.Spec.Containers) == 0 {
		return response, errors.New("can not admit a pod without containers")
	}

	mutate(pod)

	patchBytes, err := createJSONPatch(request.Object.Raw, pod)
	if err != nil {
		return response, err
	}

	response.Patch = patchBytes
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

func mutate(pod *corev1.Pod) {
	secretInjectorContainer := corev1.Container{Name: "nginx", Image: "nginx:latest"}
	pod.Spec.InitContainers = append(pod.Spec.InitContainers, secretInjectorContainer)

	sharedVolume := corev1.Volume{Name: "voltron-env"}
	sharedVolume.EmptyDir = &corev1.EmptyDirVolumeSource{Medium: corev1.StorageMediumMemory}
	pod.Spec.Volumes = append(pod.Spec.Volumes, sharedVolume)
}

func createJSONPatch(originalPodJSON []byte, mutated *corev1.Pod) ([]byte, error) {
	mutatedPodJSON, err := json.Marshal(mutated)
	if err != nil {
		return nil, fmt.Errorf("could not marshall pod: %v", err)
	}

	patch, err := jsonpatch.CreatePatch(originalPodJSON, mutatedPodJSON)
	if err != nil {
		return nil, fmt.Errorf("could not create a patch: %v", err)
	}

	patchBytes, err := json.Marshal(patch)
	if err != nil {
		return nil, fmt.Errorf("could not marshall patches: %v", err)
	}

	return patchBytes, nil
}

//NewAdmitter returns a Admitter
func NewAdmitter() Admitter {
	scheme := runtime.NewScheme()
	corev1.AddToScheme(scheme)
	codecs := serializer.NewCodecFactory(scheme)

	return mutator{
		decoder: codecs.UniversalDeserializer(),
	}
}
