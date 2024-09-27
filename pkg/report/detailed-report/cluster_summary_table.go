package detailedreport

import (
	"context"
	"encoding/csv"
	"fmt"
	"strconv"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
)

// Generates a summary table of cluster resources in CSV format.
func GenerateClusterSummaryCSV(writer *csv.Writer, clientset *kubernetes.Clientset, metricsClient *metricsv.Clientset) error {
	// Fetch node metrics
	nodeMetricsList, err := metricsClient.MetricsV1beta1().NodeMetricses().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("error fetching node metrics: %v", err)
	}

	// Fetch node status
	nodeList, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("error fetching nodes: %v", err)
	}

	// Fetch pods
	podList, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("error fetching pods: %v", err)
	}

	totalAllocatableCPU := resource.NewQuantity(0, resource.DecimalSI)
	totalAllocatableMemory := resource.NewQuantity(0, resource.BinarySI)
	totalAvailableCPU := resource.NewQuantity(0, resource.DecimalSI)
	totalAvailableMemory := resource.NewQuantity(0, resource.BinarySI)

	for _, node := range nodeList.Items {
		allocatableCPU := node.Status.Allocatable[v1.ResourceCPU]
		allocatableMemory := node.Status.Allocatable[v1.ResourceMemory]

		// Add allocatable resources
		totalAllocatableCPU.Add(allocatableCPU)
		totalAllocatableMemory.Add(allocatableMemory)

		for _, metric := range nodeMetricsList.Items {
			if metric.Name == node.Name {
				usedCPU := metric.Usage[v1.ResourceCPU]
				usedMemory := metric.Usage[v1.ResourceMemory]

				// Calculate available resources
				availableCPU := allocatableCPU.DeepCopy()
				availableCPU.Sub(usedCPU)
				totalAvailableCPU.Add(availableCPU)

				availableMemory := allocatableMemory.DeepCopy()
				availableMemory.Sub(usedMemory)
				totalAvailableMemory.Add(availableMemory)
			}
		}
	}

	// Convert quantities to meaningful units
	allocatableCPUInMillis := totalAllocatableCPU.MilliValue()
	allocatableMemoryInMi := totalAllocatableMemory.ScaledValue(resource.Mega)

	availableCPUInMillis := totalAvailableCPU.MilliValue()
	availableMemoryInMi := totalAvailableMemory.ScaledValue(resource.Mega)

	availableCPUPercent := (float64(availableCPUInMillis) / float64(allocatableCPUInMillis)) * 100
	availableMemoryPercent := (float64(availableMemoryInMi) / float64(allocatableMemoryInMi)) * 100

	// Write Total Nodes and Total Pods rows
	totalNodesRow := []string{"TOTAL NODES", strconv.Itoa(len(nodeList.Items)), ""}
	if err := writer.Write(totalNodesRow); err != nil {
		return fmt.Errorf("error writing Total Nodes row: %v", err)
	}

	totalPodsRow := []string{"TOTAL PODS", strconv.Itoa(len(podList.Items)), ""}
	if err := writer.Write(totalPodsRow); err != nil {
		return fmt.Errorf("error writing Total Pods row: %v", err)
	}

	emptyRow := []string{"", "", ""}
	if err := writer.Write(emptyRow); err != nil {
		return fmt.Errorf("error writing empty row: %v", err)
	}

	headers := []string{"RESOURCE TYPE", "CPU (mC)", "Memory (MiB)"}
	if err := writer.Write(headers); err != nil {
		return fmt.Errorf("error writing CSV headers: %v", err)
	}

	allocatableRow := []string{
		"Cluster Allocatable",
		strconv.Itoa(int(allocatableCPUInMillis)),
		strconv.Itoa(int(allocatableMemoryInMi)),
	}
	if err := writer.Write(allocatableRow); err != nil {
		return fmt.Errorf("error writing Cluster Allocatable row: %v", err)
	}

	availableRow := []string{
		"Cluster Available",
		strconv.Itoa(int(availableCPUInMillis)),
		strconv.Itoa(int(availableMemoryInMi)),
	}
	if err := writer.Write(availableRow); err != nil {
		return fmt.Errorf("error writing Cluster Available row: %v", err)
	}

	availablePercentRow := []string{
		"Cluster Available (%)",
		fmt.Sprintf("%.2f%%", availableCPUPercent),
		fmt.Sprintf("%.2f%%", availableMemoryPercent),
	}
	if err := writer.Write(availablePercentRow); err != nil {
		return fmt.Errorf("error writing Cluster Available Percent row: %v", err)
	}

	return nil
}
