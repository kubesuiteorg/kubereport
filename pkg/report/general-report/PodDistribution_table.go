package tables

import (
	"context"
	"fmt"

	"github.com/jung-kurt/gofpdf/v2"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Generates a report of pod distribution by namespace and node.
func GeneratePodDistributionReport(pdf *gofpdf.Fpdf, clientset *kubernetes.Clientset) error {
	// Fetch pods
	podList, err := clientset.CoreV1().Pods(v1.NamespaceAll).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("error fetching pods: %v", err)
	}

	namespaceCounts := make(map[string]int)
	nodeCounts := make(map[string]int)

	// Count pods by namespace and node
	for _, pod := range podList.Items {
		namespaceCounts[pod.Namespace]++
		nodeCounts[pod.Spec.NodeName]++
	}

	pdf.Ln(5)
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(0, 10, "Pod Distribution By Namespace")
	pdf.Ln(10)

	colWidth := 95.0

	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(colWidth, 10, "Name", "1", 0, "L", false, 0, "")
	pdf.CellFormat(30, 10, "Value", "1", 1, "L", false, 0, "")

	pdf.SetFont("Arial", "", 10)
	for ns, count := range namespaceCounts {
		pdf.CellFormat(colWidth, 10, ns, "1", 0, "L", false, 0, "")
		pdf.CellFormat(30, 10, fmt.Sprintf("%d pods", count), "1", 1, "L", false, 0, "")
	}

	pdf.Ln(5)
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(0, 10, "Pod Distribution By Node")
	pdf.Ln(10)

	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(colWidth, 10, "Node", "1", 0, "L", false, 0, "")
	pdf.CellFormat(30, 10, "Value", "1", 1, "L", false, 0, "")

	pdf.SetFont("Arial", "", 10)
	for node, count := range nodeCounts {
		pdf.CellFormat(colWidth, 10, node, "1", 0, "L", false, 0, "")
		pdf.CellFormat(30, 10, fmt.Sprintf("%d pods", count), "1", 1, "L", false, 0, "")
	}

	return nil
}
