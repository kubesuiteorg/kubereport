package tables

import (
	"context"
	"fmt"
	"strings"

	"github.com/jung-kurt/gofpdf/v2"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Generates a summary table of deployments and their container images.
func GenerateDeploymentImageTable(pdf *gofpdf.Fpdf, clientset *kubernetes.Clientset) error {
	// Fetch deployments from all namespaces
	deploymentList, err := clientset.AppsV1().Deployments("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("error fetching deployments: %v", err)
	}

	// Define column widths using a map for better readability
	colWidths := map[string]float64{
		"Deployment": 60.0,
		"Image Name & Version (Init Containers First)": 135.0,
	}

	headers := []string{"Deployment", "Image Name & Version (Init Containers First)"}

	// Function to render table headers
	renderHeaders := func() {
		pdf.SetFont("Arial", "B", 7)
		for _, header := range headers {
			pdf.CellFormat(colWidths[header], 8, header, "1", 0, "C", false, 0, "")
		}
		pdf.Ln(-1)
	}

	// Add headers to the first page
	renderHeaders()

	// Iterate over deployments to get resource information
	for _, deployment := range deploymentList.Items {
		deploymentName := deployment.Name
		imageVersions := getImageVersions(deployment)

		// Check if we need to add a new page
		rowHeight := 8.0
		_, pageHeight := pdf.GetPageSize()
		if pdf.GetY()+rowHeight > pageHeight-20 {
			pdf.AddPage()
			renderHeaders() // Reprint the headers on the new page
		}

		// Print data for each deployment
		pdf.SetFont("Arial", "", 6)

		pdf.CellFormat(colWidths["Deployment"], rowHeight, deploymentName, "1", 0, "L", false, 0, "")
		pdf.MultiCell(colWidths["Image Name & Version (Init Containers First)"], rowHeight, imageVersions, "1", "L", false) // MultiCell for wrapping long image lists
	}

	return nil
}

// Extracts image versions from a deployment, including init containers first.
func getImageVersions(deployment appsv1.Deployment) string {
	var images []string

	// Init Containers
	for _, container := range deployment.Spec.Template.Spec.InitContainers {
		images = append(images, container.Image)
	}

	// Main Containers
	for _, container := range deployment.Spec.Template.Spec.Containers {
		images = append(images, container.Image)
	}

	return strings.Join(images, ", ")
}
