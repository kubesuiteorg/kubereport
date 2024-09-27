package tables

import (
	"context"
	"fmt"
	"sort"
	"strconv"

	"github.com/jung-kurt/gofpdf/v2"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// PodResourceUsage holds the resource usage data for a pod.
type PodResourceUsage struct {
	Name                 string
	RequestedCPUInMillis int64
	LimitCPUInMillis     int64
	RequestedMemoryInMi  int64
	LimitMemoryInMi      int64
}

// Generates a report of pod resource usage.
func GeneratePodResourceUsageTable(pdf *gofpdf.Fpdf, clientset *kubernetes.Clientset) error {
	// Fetch pods
	podList, err := clientset.CoreV1().Pods(v1.NamespaceAll).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("error fetching pods: %v", err)
	}

	var podData []PodResourceUsage

	// Iterate over pods to get their resource information
	for _, pod := range podList.Items {
		podName := pod.Name

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

		podData = append(podData, PodResourceUsage{
			Name:                 podName,
			RequestedCPUInMillis: requestedCPUInMillis,
			LimitCPUInMillis:     limitCPUInMillis,
			RequestedMemoryInMi:  requestedMemoryInMi,
			LimitMemoryInMi:      limitMemoryInMi,
		})
	}

	// Sort pods by CPU usage (or any other metric) in descending order
	sort.Slice(podData, func(i, j int) bool {
		return podData[i].RequestedCPUInMillis > podData[j].RequestedCPUInMillis
	})

	_, pageHeight := pdf.GetPageSize()
	if pdf.GetY() > pageHeight-20 {
		pdf.AddPage()
	}

	if pdf.GetY() > pageHeight-40 {
		pdf.AddPage()
	}

	colWidth := 38.0

	// Function to print table headers
	printHeaders := func() {
		pdf.SetFont("Arial", "B", 8)
		headers := []string{
			"Pod Name",
			"CPU mC req",
			"CPU mC limit",
			"Mem MiB req",
			"Mem MiB limit",
		}
		for _, header := range headers {
			x, y := pdf.GetXY()
			pdf.MultiCell(colWidth, 8, header, "1", "C", false)
			pdf.SetXY(x+colWidth, y)
		}
		pdf.Ln(-1)
	}

	printHeaders()

	// Define a function to handle row addition
	addRow := func(podName string, requestedCPUInMillis int64, limitCPUInMillis int64, requestedMemoryInMi int64, limitMemoryInMi int64) {
		_, pageHeight := pdf.GetPageSize()
		if pdf.GetY() > pageHeight-40 {
			pdf.AddPage()
			printHeaders()
		}

		// Add pod data to table in a single row
		pdf.SetFont("Arial", "", 10)
		x, y := pdf.GetXY()
		pdf.MultiCell(colWidth, 8, podName, "1", "L", false)
		height := pdf.GetY() - y

		pdf.SetXY(x+colWidth, y)
		pdf.CellFormat(colWidth, height, strconv.Itoa(int(requestedCPUInMillis)), "1", 0, "C", false, 0, "")
		pdf.CellFormat(colWidth, height, strconv.Itoa(int(limitCPUInMillis)), "1", 0, "C", false, 0, "")
		pdf.CellFormat(colWidth, height, strconv.Itoa(int(requestedMemoryInMi)), "1", 0, "C", false, 0, "")
		pdf.CellFormat(colWidth, height, strconv.Itoa(int(limitMemoryInMi)), "1", 1, "C", false, 0, "")
	}

	for _, pod := range podData {
		addRow(pod.Name, pod.RequestedCPUInMillis, pod.LimitCPUInMillis, pod.RequestedMemoryInMi, pod.LimitMemoryInMi)
	}

	return nil
}
