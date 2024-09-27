package detailedreport

import (
	"context"
	"encoding/csv"
	"fmt"
	"sort"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// LimitRangeInfo holds the information for a Kubernetes Limit Range.
type LimitRangeInfo struct {
	Name            string
	Namespace       string
	Limits          string
	Requests        string
	Age             string
	Annotations     string
	Status          string
	LimitType       string
	DefaultLimits   string
	DefaultRequests string
}

// Generates a CSV report of Kubernetes Limit Ranges.
func GenerateLimitRangeReportCSV(writer *csv.Writer, clientset *kubernetes.Clientset) error {
	// Fetch Limit Ranges
	lrs, err := clientset.CoreV1().LimitRanges("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("error fetching Limit Ranges: %v", err)
	}

	var lrData []LimitRangeInfo

	// Iterate over Limit Ranges to get their information
	for _, lr := range lrs.Items {
		age := time.Since(lr.CreationTimestamp.Time).Round(time.Hour).String()

		// Prepare Limits and Requests
		limits := fmt.Sprintf("%v", lr.Spec.Limits)
		requests := fmt.Sprintf("%v", lr.Spec.Limits)

		defaultLimits := ""
		defaultRequests := ""
		for _, limit := range lr.Spec.Limits {
			if limit.Default != nil {
				defaultLimits = fmt.Sprintf("CPU: %s, Memory: %s", limit.Default[v1.ResourceCPU], limit.Default[v1.ResourceMemory])
			}
			if limit.DefaultRequest != nil {
				defaultRequests = fmt.Sprintf("CPU: %s, Memory: %s", limit.DefaultRequest[v1.ResourceCPU], limit.DefaultRequest[v1.ResourceMemory])
			}
		}

		status := "Active" // Placeholder; adjust logic as needed

		annotations := fmt.Sprintf("%v", lr.Annotations)

		// Create a record for the Limit Range
		lrData = append(lrData, LimitRangeInfo{
			Name:            lr.Name,
			Namespace:       lr.Namespace,
			Limits:          limits,
			Requests:        requests,
			Age:             age,
			Annotations:     annotations,
			Status:          status,
			LimitType:       "Container",
			DefaultLimits:   defaultLimits,
			DefaultRequests: defaultRequests,
		})
	}

	sort.Slice(lrData, func(i, j int) bool {
		return lrData[i].Name < lrData[j].Name
	})

	// Write CSV headers
	if err := writer.Write([]string{
		"RESOURCE NAME",
		"NAMESPACE",
		"LIMITS",
		"REQUESTS",
		"AGE",
		"ANNOTATIONS",
		"STATUS",
		"LIMIT TYPE",
		"DEFAULT LIMITS",
		"DEFAULT REQUESTS",
	}); err != nil {
		return fmt.Errorf("error writing headers to CSV: %v", err)
	}

	// Write Limit Range data rows
	for _, lr := range lrData {
		record := []string{
			lr.Name,
			lr.Namespace,
			lr.Limits,
			lr.Requests,
			lr.Age,
			lr.Annotations,
			lr.Status,
			lr.LimitType,
			lr.DefaultLimits,
			lr.DefaultRequests,
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
