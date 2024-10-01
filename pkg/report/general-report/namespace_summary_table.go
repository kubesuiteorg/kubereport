package tables

import (
	"context"
	"fmt"

	"github.com/jung-kurt/gofpdf/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Generates a summary table of namespaces, deployments, pods, and services.
func GenerateNamespaceSummaryTable(pdf *gofpdf.Fpdf, clientset *kubernetes.Clientset) error {
	// Fetch namespaces
	namespaceList, err := clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("error fetching namespaces: %v", err)
	}

	// Define column widths for each header
	colWidths := []float64{
		90.0,
		30.0,
		30.0,
		30.0,
	}

	headers := []string{
		"Namespace",
		"Deployments",
		"Pods",
		"Services",
	}

	// Function to render table headers
	renderHeaders := func() {
		pdf.SetFont("Arial", "B", 8)
		for i, header := range headers {
			pdf.CellFormat(colWidths[i], 8, header, "1", 0, "C", false, 0, "")
		}
		pdf.Ln(-1)
	}

	// Add headers to the first page
	renderHeaders()

	// Iterate over namespaces to get resource information
	for _, ns := range namespaceList.Items {
		// Count Deployments
		deployments, err := clientset.AppsV1().Deployments(ns.Name).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return fmt.Errorf("error fetching deployments for namespace %s: %v", ns.Name, err)
		}
		deploymentCount := len(deployments.Items)

		// Count Pods
		pods, err := clientset.CoreV1().Pods(ns.Name).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return fmt.Errorf("error fetching pods for namespace %s: %v", ns.Name, err)
		}
		podCount := len(pods.Items)

		// Count Services
		services, err := clientset.CoreV1().Services(ns.Name).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return fmt.Errorf("error fetching services for namespace %s: %v", ns.Name, err)
		}
		serviceCount := len(services.Items)

		// Check if we need to add a new page
		rowHeight := 8.0
		_, pageHeight := pdf.GetPageSize()
		if pdf.GetY()+rowHeight > pageHeight-20 {
			pdf.AddPage()
			renderHeaders() // Reprint the headers on the new page
		}

		// Print data for each namespace
		pdf.SetFont("Arial", "", 8)

		// Print the Namespace column using CellFormat
		pdf.CellFormat(colWidths[0], rowHeight, ns.Name, "1", 0, "L", false, 0, "")
		pdf.CellFormat(colWidths[1], rowHeight, fmt.Sprintf("%d", deploymentCount), "1", 0, "C", false, 0, "")
		pdf.CellFormat(colWidths[2], rowHeight, fmt.Sprintf("%d", podCount), "1", 0, "C", false, 0, "")
		pdf.CellFormat(colWidths[3], rowHeight, fmt.Sprintf("%d", serviceCount), "1", 1, "C", false, 0, "")
	}

	return nil
}
