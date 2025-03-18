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

// GenerateDeploymentImageTable creates a PDF table summarizing deployments and container images.
func GenerateDeploymentImageTable(pdf *gofpdf.Fpdf, clientset *kubernetes.Clientset, releaseVersion, teamLabel string) error {
	// Fetch all deployments
	deploymentList, err := clientset.AppsV1().Deployments("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("error fetching deployments: %v", err)
	}

	// Check if any deployment has an init container
	hasInitContainers := false
	for _, deployment := range deploymentList.Items {
		if len(deployment.Spec.Template.Spec.InitContainers) > 0 {
			hasInitContainers = true
			break
		}
	}

	// Define column widths
	colWidths := map[string]float64{
		"Deployment":             70.0,
		"Application Image Name": 90.0,
		"Version":                25.0,
	}

	headers := []string{"Deployment"}

	// Add "Init Containers Image & Version" after "Deployment" only if needed
	if hasInitContainers {
		colWidths["Init Containers Image & Version"] = 86.0
		headers = append(headers, "Init Containers Image & Version")
	}

	// Add remaining columns
	headers = append(headers, "Application Image Name", "Version")

	// Add "Team Name" column if team label is provided
	includeTeamColumn := teamLabel != ""
	if includeTeamColumn {
		colWidths["Team Name"] = 40.0
		headers = append(headers, "Team Name")
	}

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

	// Iterate over deployments
	for _, deployment := range deploymentList.Items {
		deploymentName := deployment.Name
		initImages, mainImageNames, mainVersions := getImageVersions(deployment, hasInitContainers)

		// Fetch team name from metadata if applicable
		teamName := "N/A"
		if includeTeamColumn {
			if value, exists := deployment.Labels[teamLabel]; exists {
				teamName = value
			}
		}

		// Check if we need to add a new page
		rowHeight := 8.0
		_, pageHeight := pdf.GetPageSize()
		if pdf.GetY()+rowHeight > pageHeight-20 {
			pdf.AddPage()
			renderHeaders() // Reprint headers
		}

		// Print Deployment Name
		pdf.SetFont("Arial", "", 7)
		pdf.CellFormat(colWidths["Deployment"], rowHeight, deploymentName, "1", 0, "L", false, 0, "")

		// Print Init Containers Image & Version (if applicable)
		if hasInitContainers {
			startX := pdf.GetX()
			startY := pdf.GetY()
			pdf.MultiCell(colWidths["Init Containers Image & Version"], rowHeight, initImages, "1", "L", false)
			pdf.SetXY(startX+colWidths["Init Containers Image & Version"], startY)
		}

		// Print Application Image Name Image Name
		startX := pdf.GetX()
		startY := pdf.GetY()
		pdf.MultiCell(colWidths["Application Image Name"], rowHeight, mainImageNames, "1", "L", false)
		pdf.SetXY(startX+colWidths["Application Image Name"], startY)

		// Print Application Image Name Version
		startX = pdf.GetX()
		startY = pdf.GetY()
		pdf.MultiCell(colWidths["Version"], rowHeight, mainVersions, "1", "L", false)
		pdf.SetXY(startX+colWidths["Version"], startY)

		// Print Team Name (if applicable)
		if includeTeamColumn {
			pdf.CellFormat(colWidths["Team Name"], rowHeight, teamName, "1", 0, "C", false, 0, "")
		}

		pdf.Ln(-1) // Move to the next row
	}

	return nil
}

// Extracts init container images and separates main container image names and versions.
func getImageVersions(deployment appsv1.Deployment, hasInitContainers bool) (string, string, string) {
	var initImages []string
	var mainImageNames []string
	var mainVersions []string

	// Init Containers
	for _, container := range deployment.Spec.Template.Spec.InitContainers {
		initImages = append(initImages, container.Image)
	}

	// Application Image Name - Extract Image Name and Version separately
	for _, container := range deployment.Spec.Template.Spec.Containers {
		imageParts := strings.Split(container.Image, ":")
		imageName := imageParts[0] // Extracts the repository name

		// If the version is missing, assume "latest"
		imageVersion := "latest"
		if len(imageParts) > 1 {
			imageVersion = imageParts[1]
		}

		mainImageNames = append(mainImageNames, imageName)
		mainVersions = append(mainVersions, imageVersion)
	}

	// If there are no init containers, return "-" to ensure column visibility
	if hasInitContainers && len(initImages) == 0 {
		initImages = append(initImages, "-")
	}

	return strings.Join(initImages, ", "), strings.Join(mainImageNames, ", "), strings.Join(mainVersions, ", ")
}
