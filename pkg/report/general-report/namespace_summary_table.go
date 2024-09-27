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

	colWidth := 40.0
	headers := []string{
		"Namespace",
		"Deployments",
		"Pods",
		"Services",
	}

	pdf.SetFont("Arial", "B", 8)
	for _, header := range headers {
		pdf.CellFormat(colWidth, 8, header, "1", 0, "C", false, 0, "")
	}
	pdf.Ln(-1)

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

		pdf.SetFont("Arial", "", 8)
		pdf.CellFormat(colWidth, 8, ns.Name, "1", 0, "L", false, 0, "")
		pdf.CellFormat(colWidth, 8, fmt.Sprintf("%d", deploymentCount), "1", 0, "C", false, 0, "")
		pdf.CellFormat(colWidth, 8, fmt.Sprintf("%d", podCount), "1", 0, "C", false, 0, "")
		pdf.CellFormat(colWidth, 8, fmt.Sprintf("%d", serviceCount), "1", 1, "C", false, 0, "")
	}

	return nil
}
