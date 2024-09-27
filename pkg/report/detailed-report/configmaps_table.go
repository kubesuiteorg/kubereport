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

// ConfigMapInfo holds the information for a Kubernetes ConfigMap.
type ConfigMapInfo struct {
	Name      string
	Namespace string
	DataItems int
	Age       string
	Labels    string
}

// Generates a CSV report of Kubernetes ConfigMaps.
func GenerateConfigMapReportCSV(writer *csv.Writer, clientset *kubernetes.Clientset) error {
	// Fetch ConfigMaps
	configMaps, err := clientset.CoreV1().ConfigMaps(metav1.NamespaceAll).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("error fetching ConfigMaps: %v", err)
	}

	var configMapData []ConfigMapInfo

	// Iterate over ConfigMaps to get their information
	for _, cm := range configMaps.Items {
		age := time.Since(cm.CreationTimestamp.Time).Round(time.Hour).String()

		dataItems := len(cm.Data)

		labels := ""
		if len(cm.Labels) > 0 {
			labels = fmt.Sprintf("%v", cm.Labels)
		}

		// Create a record for the ConfigMap
		configMapData = append(configMapData, ConfigMapInfo{
			Name:      cm.Name,
			Namespace: cm.Namespace,
			DataItems: dataItems,
			Age:       age,
			Labels:    labels,
		})
	}

	sort.Slice(configMapData, func(i, j int) bool {
		return configMapData[i].Name < configMapData[j].Name
	})

	// Write CSV headers
	if err := writer.Write([]string{
		"CONFIGMAP NAME",
		"NAMESPACE",
		"DATA ITEMS",
		"AGE",
		"LABELS",
	}); err != nil {
		return fmt.Errorf("error writing headers to CSV: %v", err)
	}

	// Write ConfigMap data rows
	for _, cm := range configMapData {
		record := []string{
			cm.Name,
			cm.Namespace,
			strconv.Itoa(cm.DataItems),
			cm.Age,
			cm.Labels,
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
