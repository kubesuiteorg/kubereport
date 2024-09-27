package tables

import (
	"context"
	"fmt"

	"github.com/jung-kurt/gofpdf/v2"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Generates a report of pod details.
func GeneratePodDetailsTable(pdf *gofpdf.Fpdf, clientset *kubernetes.Clientset) error {
	// Fetch pods
	podList, err := clientset.CoreV1().Pods(v1.NamespaceAll).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("error fetching pods: %v", err)
	}

	_, pageHeight := pdf.GetPageSize()
	if pdf.GetY() > pageHeight-20 {
		pdf.AddPage()
	}

	if pdf.GetY() > pageHeight-40 {
		pdf.AddPage()
	}

	colWidths := map[string]float64{
		"Pod Name":  85.0,
		"Namespace": 85.0,
		"Status":    25.0,
	}

	printHeaders := func() {
		pdf.SetFont("Arial", "B", 8)
		headers := []string{
			"Pod Name",
			"Namespace",
			"Status",
		}
		for _, header := range headers {
			pdf.CellFormat(colWidths[header], 8, header, "1", 0, "C", false, 0, "")
		}
		pdf.Ln(-1)
	}

	printHeaders()

	addRow := func(podName string, namespace string, status string) {
		_, pageHeight := pdf.GetPageSize()
		if pdf.GetY() > pageHeight-40 {
			pdf.AddPage()
			printHeaders()
		}

		pdf.SetFont("Arial", "", 10)
		lineHeight := 8.0
		maxLines := max(
			len(pdf.SplitLines([]byte(podName), colWidths["Pod Name"])),
			len(pdf.SplitLines([]byte(namespace), colWidths["Namespace"])),
			len(pdf.SplitLines([]byte(status), colWidths["Status"])),
		)
		rowHeight := float64(maxLines) * lineHeight

		xPos := pdf.GetX()
		yPos := pdf.GetY()

		pdf.CellFormat(colWidths["Pod Name"], rowHeight, podName, "1", 0, "L", false, 0, "")
		pdf.SetXY(xPos+colWidths["Pod Name"], yPos)
		pdf.CellFormat(colWidths["Namespace"], rowHeight, namespace, "1", 0, "L", false, 0, "")
		pdf.SetXY(xPos+colWidths["Pod Name"]+colWidths["Namespace"], yPos)
		pdf.CellFormat(colWidths["Status"], rowHeight, status, "1", 0, "L", false, 0, "")
		pdf.Ln(rowHeight)
	}

	// Iterate over pods to get their details
	for _, pod := range podList.Items {
		podName := pod.Name
		namespace := pod.Namespace
		status := string(pod.Status.Phase)

		addRow(podName, namespace, status)
	}

	return nil
}

// Helper function to find the maximum number of lines needed for a cell
func max(lines ...int) int {
	maxLines := 0
	for _, line := range lines {
		if line > maxLines {
			maxLines = line
		}
	}
	return maxLines
}
