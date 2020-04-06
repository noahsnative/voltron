package e2e

import (
	"os"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/kind/pkg/cluster"
)

func TestWebhookServer(t *testing.T) {
	//createCluster()
	clientset, err := createClientset()
	assert.NoError(t, err)

	err = createTLSSecret(clientset)
	assert.NoError(t, err)
}

func createCluster() {
	provider := cluster.NewProvider()
	provider.Create(
		"e2e",
		cluster.CreateWithNodeImage("kindest/node:v1.17.0"),
		cluster.CreateWithWaitForReady(5*time.Minute),
		cluster.CreateWithDisplayUsage(true),
		cluster.CreateWithDisplaySalutation(true))
}

func createClientset() (*kubernetes.Clientset, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	kubeconfigPath := path.Join(homeDir, ".kube", "config")

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(config)
}

func createTLSSecret(clientset *kubernetes.Clientset) error {
	key, cert, err := provisionTLSCertificate(clientset)
	if err != nil {
		return err
	}

	secretName := "webhook-cert-secret"

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: secretName,
		},
		Data: map[string][]byte{
			"cert": cert,
			"key":  key,
		},
	}

	var opts metav1.GetOptions
	if existing, err := clientset.CoreV1().Secrets("default").Get(secretName, opts); err == nil {
		existing.Data = secret.Data
		_, err = clientset.CoreV1().Secrets("default").Update(existing)
	} else {
		_, err = clientset.CoreV1().Secrets("default").Create(secret)
	}

	return err
}
