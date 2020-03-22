package mutate

import (
	"encoding/json"
	"errors"
	"fmt"
	"path"

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

	injectorBinName        = "injector"
	injectorDir            = "/bin/voltron"
	injectorContainerName  = "nginx"
	injectorContainerImage = "nginx:latest"
	injectorVolumeName     = "voltron-env"
)

func (m mutator) Admit(request v1beta1.AdmissionRequest) (v1beta1.AdmissionResponse, error) {
	var response v1beta1.AdmissionResponse

	pod, err := m.extractPod(&request)
	if err != nil {
		return response, err
	}

	err = mutate(pod)
	if err != nil {
		return response, err
	}

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

func mutate(pod *corev1.Pod) error {
	if len(pod.Spec.Containers) == 0 {
		return errors.New("can not admit a pod without containers")
	}

	sharedVolume := corev1.Volume{Name: injectorVolumeName}
	sharedVolume.EmptyDir = &corev1.EmptyDirVolumeSource{Medium: corev1.StorageMediumMemory}
	pod.Spec.Volumes = append(pod.Spec.Volumes, sharedVolume)

	sharedVolumeMount := corev1.VolumeMount{
		Name:      sharedVolume.Name,
		MountPath: injectorDir,
	}
	injectorContainer := corev1.Container{
		Name:         injectorContainerName,
		Image:        injectorContainerImage,
		VolumeMounts: []corev1.VolumeMount{sharedVolumeMount}}
	pod.Spec.InitContainers = append(pod.Spec.InitContainers, injectorContainer)

	mainContainer := &pod.Spec.Containers[0]
	mainContainer.VolumeMounts = append(mainContainer.VolumeMounts, sharedVolumeMount)
	mainContainer.Command = []string{path.Join(injectorDir, injectorBinName)}

	return nil
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

// NewAdmitter returns a Admitter
func NewAdmitter() (Admitter, error) {
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		return nil, err
	}

	codecs := serializer.NewCodecFactory(scheme)

	return mutator{
		decoder: codecs.UniversalDeserializer(),
	}, nil
}
