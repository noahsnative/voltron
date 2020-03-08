package mutate

import "k8s.io/api/admission/v1beta1"

// Admitter validates an admission request and admits it, possibly mutating
type Admitter interface {
	Admit(v1beta1.AdmissionRequest) (v1beta1.AdmissionResponse, error)
}

type noOpAdmitter struct{}

func (a noOpAdmitter) Admit(v1beta1.AdmissionRequest) (v1beta1.AdmissionResponse, error) {
	return v1beta1.AdmissionResponse{Allowed: true}, nil
}

// NewAdmitter creates an instance of Admitter with provided options
func NewAdmitter() Admitter {
	return noOpAdmitter{}
}
