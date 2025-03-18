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

func GenerateDeploymentImageTable(pdf *gofpdf.Fpdf, clientset *kubernetes.Clientset, releaseVersion, teamLabel string) error {
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
		"Deployment":             60.0,
		"Application Image Name": 85.0,
		"Version":                30.0,
	}

	headers := []string{"Deployment"}

	if hasInitContainers {
		colWidths["Init Containers Image & Version"] = 85.0
		headers = append(headers, "Init Containers Image & Version")
	}

	headers = append(headers, "Application Image Name", "Version")

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

	renderHeaders() // Add headers to the first page

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

		rowData := []string{deploymentName}
		if hasInitContainers {
			rowData = append(rowData, initImages)
		}
		rowData = append(rowData, mainImageNames, mainVersions)
		if includeTeamColumn {
			rowData = append(rowData, teamName)
		}

		// Calculate row height
		rowHeight := calculateMaxHeight(pdf, rowData, colWidths, headers, hasInitContainers, includeTeamColumn)
		_, pageHeight := pdf.GetPageSize()
		if pdf.GetY()+rowHeight > pageHeight-20 {
			pdf.AddPage()
			renderHeaders()
		}

		startY := pdf.GetY()
		pdf.SetFont("Arial", "", 7)

		// Manually print each column to maintain proper alignment
		for i, text := range rowData {
			x := pdf.GetX()
			colWidth := colWidths[headers[i]]

			// Check if text needs MultiCell (if it might wrap)
			lines := pdf.SplitLines([]byte(text), colWidth)
			if len(lines) > 1 {
				pdf.MultiCell(colWidth, 4, text, "1", "L", false)
			} else {
				pdf.CellFormat(colWidth, rowHeight, text, "1", 0, "L", false, 0, "")
			}

			// Move X position for the next column
			pdf.SetXY(x+colWidth, startY)
		}

		// Move to the next row without adding an extra blank row
		pdf.SetY(startY + rowHeight)
	}

	return nil
}

func getImageVersions(deployment appsv1.Deployment, hasInitContainers bool) (string, string, string) {
	var initImages []string
	var mainImageNames []string
	var mainVersions []string

	for _, container := range deployment.Spec.Template.Spec.InitContainers {
		initImages = append(initImages, container.Image)
	}

	for _, container := range deployment.Spec.Template.Spec.Containers {
		imageParts := strings.Split(container.Image, ":")
		imageName := imageParts[0]

		imageVersion := "latest"
		if len(imageParts) > 1 {
			imageVersion = imageParts[1]
		}

		mainImageNames = append(mainImageNames, imageName)
		mainVersions = append(mainVersions, imageVersion)
	}

	if hasInitContainers && len(initImages) == 0 {
		initImages = append(initImages, "-")
	}

	return strings.Join(initImages, ", "), strings.Join(mainImageNames, ", "), strings.Join(mainVersions, ", ")
}

func calculateMaxHeight(pdf *gofpdf.Fpdf, texts []string, colWidths map[string]float64, headers []string, hasInitContainers bool, includeTeamColumn bool) float64 {
	maxHeight := 8.0

	for i, text := range texts {
		if headers[i] == "Init Containers Image & Version" && !hasInitContainers {
			continue
		}
		if headers[i] == "Team Name" && !includeTeamColumn {
			continue
		}
		lines := pdf.SplitLines([]byte(text), colWidths[headers[i]])
		height := float64(len(lines)) * 4
		if height > maxHeight {
			maxHeight = height
		}
	}
	return maxHeight
}
