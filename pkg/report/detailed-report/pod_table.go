package detailedreport

import (
	"context"
	"encoding/csv"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// PodResourceUsage holds the resource usage data for a pod.
type PodResourceUsage struct {
	Name                 string
	Namespace            string
	NodeName             string
	RequestedCPUInMillis int64
	LimitCPUInMillis     int64
	RequestedMemoryInMi  int64
	LimitMemoryInMi      int64
	Status               string
	RestartCount         int32
	Conditions           string
	Age                  string
}

// Generates a CSV report of pod resource usage.
func GeneratePodResourceUsageCSV(writer *csv.Writer, clientset *kubernetes.Clientset) error {
	// Fetch pods
	podList, err := clientset.CoreV1().Pods(v1.NamespaceAll).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("error fetching pods: %v", err)
	}

	var podData []PodResourceUsage

	// Iterate over pods to get their resource information
	for _, pod := range podList.Items {
		podName := pod.Name
		namespace := pod.Namespace
		nodeName := pod.Spec.NodeName
		status := string(pod.Status.Phase)
		restartCount := pod.Status.ContainerStatuses[0].RestartCount

		// Initialize counters for requested and limit values
		totalRequestedCPU := resource.NewQuantity(0, resource.DecimalSI)
		totalRequestedMemory := resource.NewQuantity(0, resource.BinarySI)
		totalLimitCPU := resource.NewQuantity(0, resource.DecimalSI)
		totalLimitMemory := resource.NewQuantity(0, resource.BinarySI)

		for _, container := range pod.Spec.Containers {
			requestedCPU := container.Resources.Requests[v1.ResourceCPU]
			requestedMemory := container.Resources.Requests[v1.ResourceMemory]
			limitCPU := container.Resources.Limits[v1.ResourceCPU]
			limitMemory := container.Resources.Limits[v1.ResourceMemory]

			totalRequestedCPU.Add(requestedCPU)
			totalRequestedMemory.Add(requestedMemory)
			totalLimitCPU.Add(limitCPU)
			totalLimitMemory.Add(limitMemory)
		}

		// Convert to milli values and MiB
		requestedCPUInMillis := totalRequestedCPU.MilliValue()
		requestedMemoryInMi := totalRequestedMemory.ScaledValue(resource.Mega)
		limitCPUInMillis := totalLimitCPU.MilliValue()
		limitMemoryInMi := totalLimitMemory.ScaledValue(resource.Mega)

		// Calculate the age of the pod
		age := time.Since(pod.CreationTimestamp.Time).Round(time.Hour).String()

		var conditions []string
		for _, cond := range pod.Status.Conditions {
			conditions = append(conditions, fmt.Sprintf("%s=%v", cond.Type, cond.Status))
		}
		conditionsStr := strings.Join(conditions, " ")

		// Create a record for the pod
		podData = append(podData, PodResourceUsage{
			Name:                 podName,
			Namespace:            namespace,
			NodeName:             nodeName,
			RequestedCPUInMillis: requestedCPUInMillis,
			LimitCPUInMillis:     limitCPUInMillis,
			RequestedMemoryInMi:  requestedMemoryInMi,
			LimitMemoryInMi:      limitMemoryInMi,
			Status:               status,
			RestartCount:         restartCount,
			Conditions:           conditionsStr,
			Age:                  age,
		})
	}

	sort.Slice(podData, func(i, j int) bool {
		return podData[i].RequestedCPUInMillis > podData[j].RequestedCPUInMillis
	})

	// Write CSV headers
	if err := writer.Write([]string{
		"POD NAME",
		"NAMESPACE",
		"NODE NAME",
		"CPU REQUESTS",
		"CPU LIMITS",
		"MEMORY REQUESTS",
		"MEMORY LIMITS",
		"STATUS",
		"RESTART COUNT",
		"CONDITIONS",
		"AGE",
	}); err != nil {
		return fmt.Errorf("error writing headers to CSV: %v", err)
	}

	// Write pod data rows
	for _, pod := range podData {
		record := []string{
			pod.Name,
			pod.Namespace,
			pod.NodeName,
			strconv.Itoa(int(pod.RequestedCPUInMillis)),
			strconv.Itoa(int(pod.LimitCPUInMillis)),
			strconv.Itoa(int(pod.RequestedMemoryInMi)),
			strconv.Itoa(int(pod.LimitMemoryInMi)),
			pod.Status,
			strconv.Itoa(int(pod.RestartCount)),
			pod.Conditions,
			pod.Age,
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
