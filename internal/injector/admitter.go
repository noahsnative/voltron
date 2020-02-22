package injector

import "k8s.io/api/admission/v1beta1"

// Admitter validates an admission request and admits it, possibly mutating
type Admitter interface {
	Admit(v1beta1.AdmissionRequest) (v1beta1.AdmissionResponse, error)
}
