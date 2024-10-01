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

	// Check and add a new page if needed
	ensureNewPage(pdf)

	colWidths := map[string]float64{
		"Pod Name":  88.0,
		"Namespace": 88.0,
		"Status":    20.0,
	}

	// Print table headers
	printHeaders(pdf, colWidths)

	// Iterate over pods to get their details
	for _, pod := range podList.Items {
		podName := pod.Name
		namespace := pod.Namespace
		status := string(pod.Status.Phase)

		addRow(pdf, colWidths, podName, namespace, status)
	}

	return nil
}

// Ensures a new page is added if there is not enough space.
func ensureNewPage(pdf *gofpdf.Fpdf) {
	_, pageHeight := pdf.GetPageSize()
	if pdf.GetY() > pageHeight-60 { // Adjusted threshold for adding a new page
		pdf.AddPage()
	}
}

// Prints the table headers.
func printHeaders(pdf *gofpdf.Fpdf, colWidths map[string]float64) {
	pdf.SetFont("Arial", "B", 8)
	headers := []string{"Pod Name", "Namespace", "Status"}
	for _, header := range headers {
		x, y := pdf.GetXY()
		pdf.MultiCell(colWidths[header], 8, header, "1", "C", false)
		pdf.SetXY(x+colWidths[header], y)
	}
	pdf.Ln(-1)
}

// Adds a row for a pod in the table.
func addRow(pdf *gofpdf.Fpdf, colWidths map[string]float64, podName, namespace, status string) {
	ensureNewPage(pdf)

	pdf.SetFont("Arial", "", 8)
	x, y := pdf.GetXY()
	pdf.MultiCell(colWidths["Pod Name"], 8, podName, "1", "L", false)
	height := pdf.GetY() - y

	pdf.SetXY(x+colWidths["Pod Name"], y)
	pdf.CellFormat(colWidths["Namespace"], height, namespace, "1", 0, "L", false, 0, "")
	pdf.CellFormat(colWidths["Status"], height, status, "1", 0, "L", false, 0, "")
	pdf.Ln(height)
}
