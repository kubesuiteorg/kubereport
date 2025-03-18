package tables

import (
	"context"
	"fmt"

	"github.com/jung-kurt/gofpdf/v2"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// GenerateDeploymentReadyTable creates a table with deployment details: Name, Namespace, Ready.
func GenerateDeploymentReadyTable(pdf *gofpdf.Fpdf, clientset *kubernetes.Clientset) error {
	// Fetch deployments
	deployList, err := clientset.AppsV1().Deployments(v1.NamespaceAll).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("error fetching deployments: %v", err)
	}

	// Ensure a new page if needed
	ensureNewPage(pdf)

	// Define column widths
	colWidths := map[string]float64{
		"Deployment Name": 88.0,
		"Namespace":       88.0,
		"Ready":           20.0,
	}

	// Print table headers
	printReadyHeaders(pdf, colWidths)

	// Iterate over deployments to extract details
	for _, deploy := range deployList.Items {
		deployName := deploy.Name
		namespace := deploy.Namespace
		ready := getDeploymentReadyStatus(deploy)

		// Add row to table
		addReadyRow(pdf, colWidths, deployName, namespace, ready)
	}

	return nil
}

// Get deployment ready status in "X/Y" format (Available Replicas / Desired Replicas)
func getDeploymentReadyStatus(deploy appsv1.Deployment) string {
	availableReplicas := deploy.Status.AvailableReplicas
	desiredReplicas := *deploy.Spec.Replicas
	return fmt.Sprintf("%d/%d", availableReplicas, desiredReplicas)
}

// Prints the headers for the Ready table
func printReadyHeaders(pdf *gofpdf.Fpdf, colWidths map[string]float64) {
	pdf.SetFont("Arial", "B", 8)
	headers := []string{"Deployment Name", "Namespace", "Ready"}

	for _, header := range headers {
		x, y := pdf.GetXY()
		pdf.MultiCell(colWidths[header], 8, header, "1", "C", false)
		pdf.SetXY(x+colWidths[header], y)
	}
	pdf.Ln(-1)
}

// Adds a row with deployment details
func addReadyRow(pdf *gofpdf.Fpdf, colWidths map[string]float64, deployName, namespace, ready string) {
	ensureNewPage(pdf)

	pdf.SetFont("Arial", "", 8)
	x, y := pdf.GetXY()

	pdf.MultiCell(colWidths["Deployment Name"], 8, deployName, "1", "L", false)
	pdf.SetXY(x+colWidths["Deployment Name"], y)
	pdf.MultiCell(colWidths["Namespace"], 8, namespace, "1", "L", false)
	pdf.SetXY(x+colWidths["Deployment Name"]+colWidths["Namespace"], y)
	pdf.CellFormat(colWidths["Ready"], 8, ready, "1", 0, "C", false, 0, "")

	pdf.Ln(8)
}
