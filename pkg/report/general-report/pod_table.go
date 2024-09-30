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

type PodResourceUsage struct {
	Name                 string
	RequestedCPUInMillis int64
	LimitCPUInMillis     int64
	RequestedMemoryInMi  int64
	LimitMemoryInMi      int64
}

func GeneratePodResourceUsageTable(pdf *gofpdf.Fpdf, clientset *kubernetes.Clientset) error {
	podList, err := clientset.CoreV1().Pods(v1.NamespaceAll).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("error fetching pods: %v", err)
	}

	var podData []PodResourceUsage

	for _, pod := range podList.Items {
		podName := pod.Name

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

	// Define column widths
	colWidths := []float64{90.0, 25.0, 25.0, 25.0, 25.0}
	headers := []string{
		"Pod Name",
		"CPU Lim(mCPU)",
		"CPU Req(mCPU)",
		"Memory Lim(MiB)",
		"Memory Req(MiB)",
	}

	// Function to print the headers
	printHeaders := func() {
		pdf.SetFont("Arial", "B", 8)
		for i, header := range headers {
			pdf.CellFormat(colWidths[i], 8, header, "1", 0, "C", false, 0, "")
		}
		pdf.Ln(8)
	}

	printHeaders()

	addRow := func(podName string, limitCPUInMillis int64, requestedCPUInMillis int64, limitMemoryInMi int64, requestedMemoryInMi int64) {
		_, pageHeight := pdf.GetPageSize()
		if pdf.GetY() > pageHeight-40 {
			pdf.AddPage()
			printHeaders()
		}

		pdf.SetFont("Arial", "", 8)

		// Print the pod name
		pdf.CellFormat(colWidths[0], 8, podName, "1", 0, "L", false, 0, "")

		// Print CPU and Memory values
		pdf.CellFormat(colWidths[1], 8, strconv.Itoa(int(limitCPUInMillis)), "1", 0, "C", false, 0, "")
		pdf.CellFormat(colWidths[2], 8, strconv.Itoa(int(requestedCPUInMillis)), "1", 0, "C", false, 0, "")
		pdf.CellFormat(colWidths[3], 8, strconv.Itoa(int(limitMemoryInMi)), "1", 0, "C", false, 0, "")
		pdf.CellFormat(colWidths[4], 8, strconv.Itoa(int(requestedMemoryInMi)), "1", 1, "C", false, 0, "")
	}

	for _, pod := range podData {
		addRow(pod.Name, pod.LimitCPUInMillis, pod.RequestedCPUInMillis, pod.LimitMemoryInMi, pod.RequestedMemoryInMi)
	}

	return nil
}
