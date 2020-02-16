package injector

import "k8s.io/api/admission/v1beta1"

// Mutator changes a requested resource and delivers response with a patch
type Mutator interface {
	mutate(v1beta1.AdmissionReview) (v1beta1.AdmissionResponse, error)
}
