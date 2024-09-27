package tables

import (
	"context"
	"fmt"
	"strconv"

	"github.com/jung-kurt/gofpdf/v2"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
)

// Generates a summary table of cluster resources.
func GenerateClusterSummaryTable(pdf *gofpdf.Fpdf, clientset *kubernetes.Clientset, metricsClient *metricsv.Clientset) error {
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

	// Fetch pod status
	podList, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("error fetching pods: %v", err)
	}

	// Calculate the total number of nodes and pods
	totalNodes := len(nodeList.Items)
	totalPods := len(podList.Items)

	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(75, 9, fmt.Sprintf("Total Nodes: %d", totalNodes), "1", 0, "C", false, 0, "")
	pdf.CellFormat(75, 9, fmt.Sprintf("Total Pods: %d", totalPods), "1", 1, "C", false, 0, "")
	pdf.Ln(5) // Add some space after the line

	// Initialize total resources
	totalAllocatableCPU := resource.NewQuantity(0, resource.DecimalSI)
	totalAllocatableMemory := resource.NewQuantity(0, resource.BinarySI)
	totalAvailableCPU := resource.NewQuantity(0, resource.DecimalSI)
	totalAvailableMemory := resource.NewQuantity(0, resource.BinarySI)

	// Calculate total allocatable and available resources
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

	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(50, 8, "Resource Type", "1", 0, "C", false, 0, "")
	pdf.CellFormat(50, 8, "CPU (mC)", "1", 0, "C", false, 0, "")
	pdf.CellFormat(50, 8, "Memory (MiB)", "1", 0, "C", false, 0, "")
	pdf.Ln(8)

	// Add Cluster Allocatable row
	pdf.SetFont("Arial", "", 10)
	pdf.CellFormat(50, 8, "Cluster Allocatable", "1", 0, "L", false, 0, "")
	pdf.CellFormat(50, 8, strconv.Itoa(int(allocatableCPUInMillis)), "1", 0, "C", false, 0, "")
	pdf.CellFormat(50, 8, strconv.Itoa(int(allocatableMemoryInMi)), "1", 1, "C", false, 0, "")

	pdf.CellFormat(50, 8, "Cluster Available", "1", 0, "L", false, 0, "")
	pdf.CellFormat(50, 8, strconv.Itoa(int(availableCPUInMillis)), "1", 0, "C", false, 0, "")
	pdf.CellFormat(50, 8, strconv.Itoa(int(availableMemoryInMi)), "1", 1, "C", false, 0, "")

	pdf.CellFormat(50, 8, "Cluster Available (%)", "1", 0, "L", false, 0, "")
	pdf.CellFormat(50, 8, fmt.Sprintf("%.2f%%", availableCPUPercent), "1", 0, "C", false, 0, "")
	pdf.CellFormat(50, 8, fmt.Sprintf("%.2f%%", availableMemoryPercent), "1", 1, "C", false, 0, "")

	return nil
}
