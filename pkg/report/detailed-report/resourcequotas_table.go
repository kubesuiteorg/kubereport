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

// ResourceQuotaInfo holds the information for a Kubernetes Resource Quota.
type ResourceQuotaInfo struct {
	Name          string
	Namespace     string
	HardLimits    string
	UsedResources string
	Age           string
	Annotations   string
	Status        string
	UsedPods      int
	RequestLimits string
	LimitType     string
}

// Generates a CSV report of Kubernetes Resource Quotas.
func GenerateResourceQuotaReportCSV(writer *csv.Writer, clientset *kubernetes.Clientset) error {
	// Fetch Resource Quotas
	rqs, err := clientset.CoreV1().ResourceQuotas("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("error fetching Resource Quotas: %v", err)
	}

	var rqData []ResourceQuotaInfo

	// Iterate over Resource Quotas to get their information
	for _, rq := range rqs.Items {
		age := time.Since(rq.CreationTimestamp.Time).Round(time.Hour).String()

		hardLimits := fmt.Sprintf("%v", rq.Spec.Hard)

		usedResources := fmt.Sprintf("%v", rq.Status.Used)

		status := "Active"
		for resourceName, hardLimit := range rq.Spec.Hard {
			if used, ok := rq.Status.Used[resourceName]; ok && used.Cmp(hardLimit) > 0 {
				status = "Exceeded"
				break
			}
		}

		// Prepare Used Pods
		usedPods := 0 // Placeholder for counting used pods, logic may vary based on your needs

		// Request Limits - assuming it's the same as Hard Limits for this example
		requestLimits := hardLimits // Customize this logic as needed

		limitType := "Resource Limit"

		annotations := fmt.Sprintf("%v", rq.Annotations)

		// Create a record for the Resource Quota
		rqData = append(rqData, ResourceQuotaInfo{
			Name:          rq.Name,
			Namespace:     rq.Namespace,
			HardLimits:    hardLimits,
			UsedResources: usedResources,
			Age:           age,
			Annotations:   annotations,
			Status:        status,
			UsedPods:      usedPods,
			RequestLimits: requestLimits,
			LimitType:     limitType,
		})
	}

	sort.Slice(rqData, func(i, j int) bool {
		return rqData[i].Name < rqData[j].Name
	})

	// Write CSV headers
	if err := writer.Write([]string{
		"RESOURCE NAME",
		"NAMESPACE",
		"HARD LIMITS",
		"USED RESOURCES",
		"AGE",
		"ANNOTATIONS",
		"STATUS",
		"USED PODS",
		"REQUEST LIMITS",
		"LIMIT TYPE",
	}); err != nil {
		return fmt.Errorf("error writing headers to CSV: %v", err)
	}

	// Write Resource Quota data rows
	for _, rq := range rqData {
		record := []string{
			rq.Name,
			rq.Namespace,
			rq.HardLimits,
			rq.UsedResources,
			rq.Age,
			rq.Annotations,
			rq.Status,
			fmt.Sprintf("%d", rq.UsedPods),
			rq.RequestLimits,
			rq.LimitType,
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
