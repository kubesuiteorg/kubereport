package detailedreport

import (
	"context"
	"encoding/csv"
	"fmt"
	"sort"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// HPAInfo holds the information for a Kubernetes Horizontal Pod Autoscaler.
type HPAInfo struct {
	Name                  string
	Namespace             string
	ScaleTargetRef        string
	MinReplicas           int32
	MaxReplicas           int32
	TargetCPUUtilization  *int32
	CurrentReplicas       int32
	Age                   string
	Conditions            string
	Metrics               string
	CurrentCPUUtilization string
	LastScaleTime         string
	Behavior              string
}

// Generates a CSV report of Kubernetes Horizontal Pod Autoscalers.
func GenerateHPAReportCSV(writer *csv.Writer, clientset *kubernetes.Clientset) error {
	// Fetch Horizontal Pod Autoscalers
	hpas, err := clientset.AutoscalingV1().HorizontalPodAutoscalers("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("error fetching Horizontal Pod Autoscalers: %v", err)
	}

	var hpaData []HPAInfo

	for _, hpa := range hpas.Items {
		age := time.Since(hpa.CreationTimestamp.Time).Round(time.Hour).String()

		scaleTargetRef := fmt.Sprintf("%s/%s", hpa.Spec.ScaleTargetRef.Kind, hpa.Spec.ScaleTargetRef.Name)

		// Metrics and Current CPU Utilization
		metrics := "N/A" // Currently not supported in autoscaling/v1
		currentCPUUtilization := "N/A"

		lastScaleTime := "N/A"
		if hpa.Status.LastScaleTime != nil {
			lastScaleTime = hpa.Status.LastScaleTime.String()
		}

		behavior := "N/A" // Can be implemented based on the HPA's behavior configuration

		// Create a record for the Horizontal Pod Autoscaler
		hpaInfo := HPAInfo{
			Name:                  hpa.Name,
			Namespace:             hpa.Namespace,
			ScaleTargetRef:        scaleTargetRef,
			MinReplicas:           *hpa.Spec.MinReplicas, // Ensure it's not nil before dereferencing
			MaxReplicas:           hpa.Spec.MaxReplicas,
			TargetCPUUtilization:  hpa.Spec.TargetCPUUtilizationPercentage,
			CurrentReplicas:       hpa.Status.CurrentReplicas,
			Age:                   age,
			Conditions:            "AbleToScale",
			Metrics:               metrics,
			CurrentCPUUtilization: currentCPUUtilization,
			LastScaleTime:         lastScaleTime,
			Behavior:              behavior,
		}

		hpaData = append(hpaData, hpaInfo)
	}

	sort.Slice(hpaData, func(i, j int) bool {
		return hpaData[i].Name < hpaData[j].Name
	})

	// Write CSV headers
	if err := writer.Write([]string{
		"HPA NAME",
		"NAMESPACE",
		"SCALE TARGET REF",
		"MIN REPLICAS",
		"MAX REPLICAS",
		"TARGET CPU UTILIZATION",
		"CURRENT REPLICAS",
		"AGE",
		"CONDITIONS",
		"METRICS",
		"CURRENT CPU UTILIZATION",
		"LAST SCALE TIME",
		"BEHAVIOR",
	}); err != nil {
		return fmt.Errorf("error writing headers to CSV: %v", err)
	}

	// Write HPA data rows
	for _, hpa := range hpaData {
		record := []string{
			hpa.Name,
			hpa.Namespace,
			hpa.ScaleTargetRef,
			fmt.Sprintf("%d", hpa.MinReplicas),
			fmt.Sprintf("%d", hpa.MaxReplicas),
			fmt.Sprintf("%d", *hpa.TargetCPUUtilization), // Ensure it's not nil before dereferencing
			fmt.Sprintf("%d", hpa.CurrentReplicas),
			hpa.Age,
			hpa.Conditions,
			hpa.Metrics,
			hpa.CurrentCPUUtilization,
			hpa.LastScaleTime,
			hpa.Behavior,
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
