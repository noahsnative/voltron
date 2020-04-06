package e2e

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"net/url"

	certificates "k8s.io/api/certificates/v1beta1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/certificate/csr"
)

func provisionTLSCertificate(clientset *kubernetes.Clientset) (key []byte, cert []byte, err error) {
	pemKey, pemCsr, err := generatePemEncodedCertificateRequest()
	if err != nil {
		return nil, nil, err
	}

	csrClient := clientset.CertificatesV1beta1().CertificateSigningRequests()
	keyUsages := []certificates.KeyUsage{certificates.UsageDigitalSignature, certificates.UsageKeyEncipherment, certificates.UsageServerAuth}
	certificateRequest, err := csr.RequestCertificate(csrClient, pemCsr, "webhook-csr", keyUsages, pemKey)
	if err != nil {
		return nil, nil, err
	}

	approvedCondition := certificates.CertificateSigningRequestCondition{
		Type: certificates.CertificateApproved,
	}
	certificateRequest.Status.Conditions = append(certificateRequest.Status.Conditions, approvedCondition)
	_, err = csrClient.UpdateApproval(certificateRequest)
	if err != nil {
		return nil, nil, err
	}

	pemCert, err := csr.WaitForCertificate(context.Background(), csrClient, certificateRequest)
	if err != nil {
		return nil, nil, err
	}

	return pemKey, pemCert, nil
}

func generatePemEncodedCertificateRequest() ([]byte, []byte, error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}

	keyBytes := x509.MarshalPKCS1PrivateKey(key)

	csrBytes, err := x509.CreateCertificateRequest(rand.Reader, createCrsTemplate(), key)
	if err != nil {
		return nil, nil, err
	}

	keyBlock := pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: keyBytes,
	}

	csrBlock := pem.Block{
		Type:  "CERTIFICATE REQUEST",
		Bytes: csrBytes,
	}

	return pem.EncodeToMemory(&keyBlock), pem.EncodeToMemory(&csrBlock), nil
}

func createCrsTemplate() *x509.CertificateRequest {
	service := "webhook"
	namespace := "default"

	subject := pkix.Name{
		CommonName: fmt.Sprintf("%s.%s.svc", service, namespace),
	}

	rawURLs := []string{
		service,
		fmt.Sprintf("%s.%s", service, namespace),
		fmt.Sprintf("%s.%s.svc", service, namespace),
	}

	var URIs []*url.URL
	for _, rawURL := range rawURLs {
		url, err := url.Parse(rawURL)
		if err != nil {
			continue
		}

		URIs = append(URIs, url)
	}

	template := x509.CertificateRequest{
		SignatureAlgorithm: x509.SHA256WithRSA,
		Subject:            subject,
		URIs:               URIs,
	}

	return &template
}
