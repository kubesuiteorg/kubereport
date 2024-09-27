package detailedreport

import (
	"context"
	"encoding/csv"
	"fmt"
	"sort"
	"strconv"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// SecretInfo holds the information for a Kubernetes Secret.
type SecretInfo struct {
	Name      string
	Namespace string
	Type      string
	DataItems int
	Age       string
	Labels    string
}

// Generates a CSV report of Kubernetes Secrets.
func GenerateSecretReportCSV(writer *csv.Writer, clientset *kubernetes.Clientset) error {
	// Fetch Secrets
	secrets, err := clientset.CoreV1().Secrets(metav1.NamespaceAll).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("error fetching Secrets: %v", err)
	}

	var secretData []SecretInfo

	for _, secret := range secrets.Items {
		age := time.Since(secret.CreationTimestamp.Time).Round(time.Hour).String()

		dataItems := len(secret.Data)

		labels := ""
		if len(secret.Labels) > 0 {
			labels = fmt.Sprintf("%v", secret.Labels)
		}

		// Create a record for the Secret
		secretData = append(secretData, SecretInfo{
			Name:      secret.Name,
			Namespace: secret.Namespace,
			Type:      string(secret.Type),
			DataItems: dataItems,
			Age:       age,
			Labels:    labels,
		})
	}

	sort.Slice(secretData, func(i, j int) bool {
		return secretData[i].Name < secretData[j].Name
	})

	// Write CSV headers
	if err := writer.Write([]string{
		"SECRET NAME",
		"NAMESPACE",
		"TYPE",
		"DATA ITEMS",
		"AGE",
		"LABELS",
	}); err != nil {
		return fmt.Errorf("error writing headers to CSV: %v", err)
	}

	// Write Secret data rows
	for _, secret := range secretData {
		record := []string{
			secret.Name,
			secret.Namespace,
			secret.Type,
			strconv.Itoa(secret.DataItems),
			secret.Age,
			secret.Labels,
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
