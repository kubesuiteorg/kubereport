package detailedreport

import (
	"context"
	"encoding/csv"
	"fmt"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// ServiceAccountInfo holds the information for a Kubernetes ServiceAccount.
type ServiceAccountInfo struct {
	Name             string
	Namespace        string
	Secrets          string
	Annotations      string
	Age              string
	ImagePullSecrets string
}

// Generates a CSV report of Kubernetes ServiceAccounts.
func GenerateServiceAccountReportCSV(writer *csv.Writer, clientset *kubernetes.Clientset) error {
	// Fetch ServiceAccounts from all namespaces
	serviceAccounts, err := clientset.CoreV1().ServiceAccounts("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("error fetching ServiceAccounts: %v", err)
	}

	var serviceAccountData []ServiceAccountInfo

	for _, sa := range serviceAccounts.Items {
		age := time.Since(sa.CreationTimestamp.Time).Round(time.Hour).String()

		var secrets []string
		for _, secret := range sa.Secrets {
			secrets = append(secrets, secret.Name)
		}
		secretsStr := "[" + strings.Join(secrets, ", ") + "]"

		var annotations string
		for key, value := range sa.Annotations {
			annotations += fmt.Sprintf("%s=%s, ", key, value)
		}
		if len(annotations) > 0 {
			annotations = annotations[:len(annotations)-2]
		}

		// Image Pull Secrets
		var imagePullSecrets []string
		for _, secret := range sa.ImagePullSecrets {
			imagePullSecrets = append(imagePullSecrets, secret.Name)
		}
		imagePullSecretsStr := "[" + strings.Join(imagePullSecrets, ", ") + "]"

		serviceAccountInfo := ServiceAccountInfo{
			Name:             sa.Name,
			Namespace:        sa.Namespace,
			Secrets:          secretsStr,
			Annotations:      annotations,
			Age:              age,
			ImagePullSecrets: imagePullSecretsStr,
		}

		serviceAccountData = append(serviceAccountData, serviceAccountInfo)
	}

	// Write CSV headers
	if err := writer.Write([]string{
		"SERVICEACCOUNT NAME",
		"NAMESPACE",
		"SECRETS",
		"ANNOTATIONS",
		"AGE",
		"IMAGE PULL SECRETS",
	}); err != nil {
		return fmt.Errorf("error writing headers to CSV: %v", err)
	}

	// Write ServiceAccount data rows
	for _, sa := range serviceAccountData {
		record := []string{
			sa.Name,
			sa.Namespace,
			sa.Secrets,
			sa.Annotations,
			sa.Age,
			sa.ImagePullSecrets,
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("error writing record to CSV: %v", err)
		}
	}

	// Flush the writer
	writer.Flush()
	if err := writer.Error(); err != nil {
		return fmt.Errorf("error flushing CSV writer: %v", err)
	}

	return nil
}
