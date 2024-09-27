package tables

import (
	"context"
	"fmt"
	"strconv"

	"github.com/jung-kurt/gofpdf/v2"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Generates a summary table of node resources.
func GenerateNodeSummaryTable(pdf *gofpdf.Fpdf, clientset *kubernetes.Clientset) error {
	// Fetch nodes
	nodeList, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("error fetching nodes: %v", err)
	}

	colWidth := 28.0
	headers := []string{
		"Node Name [Status]",
		"CPU Allo (mCPU)",
		"Memory Allo (MiB)",
		"CPU Req (mCPU)",
		"CPU Lim (mCPU)",
		"Memory Req (MiB)",
		"Memory Lim (MiB)",
	}

	pdf.SetFont("Arial", "B", 8)
	for _, header := range headers {
		pdf.CellFormat(colWidth, 8, header, "1", 0, "C", false, 0, "")
	}
	pdf.Ln(-1)

	// Iterate over nodes to get their resource information
	for _, node := range nodeList.Items {
		// Get the node status
		nodeStatus := "Unknown"
		for _, condition := range node.Status.Conditions {
			if condition.Type == v1.NodeReady && condition.Status == v1.ConditionTrue {
				nodeStatus = "Ready"
				break
			} else if condition.Type == v1.NodeReady && condition.Status != v1.ConditionTrue {
				nodeStatus = "NotReady"
				break
			}
		}

		// Combine node name and status
		nodeNameWithStatus := fmt.Sprintf("%s [%s]", node.Name, nodeStatus)

		allocatableCPU := node.Status.Allocatable[v1.ResourceCPU]
		allocatableMemory := node.Status.Allocatable[v1.ResourceMemory]

		// Convert to milli values and MiB
		allocatableCPUInMillis := allocatableCPU.MilliValue()
		allocatableMemoryInMi := allocatableMemory.ScaledValue(resource.Mega)

		// Initialize counters for requested and limit values
		totalRequestedCPU := resource.NewQuantity(0, resource.DecimalSI)
		totalLimitCPU := resource.NewQuantity(0, resource.DecimalSI)
		totalRequestedMemory := resource.NewQuantity(0, resource.BinarySI)
		totalLimitMemory := resource.NewQuantity(0, resource.BinarySI)

		// Fetch pods on the node
		podList, err := clientset.CoreV1().Pods(v1.NamespaceAll).List(context.TODO(), metav1.ListOptions{
			FieldSelector: fmt.Sprintf("spec.nodeName=%s", node.Name),
		})
		if err != nil {
			return fmt.Errorf("error fetching pods for node %s: %v", node.Name, err)
		}

		// Calculate requested and limit values for each pod
		for _, pod := range podList.Items {
			for _, container := range pod.Spec.Containers {
				// Get the requested and limit values
				requestedCPU := container.Resources.Requests[v1.ResourceCPU]
				limitCPU := container.Resources.Limits[v1.ResourceCPU]
				requestedMemory := container.Resources.Requests[v1.ResourceMemory]
				limitMemory := container.Resources.Limits[v1.ResourceMemory]

				totalRequestedCPU.Add(requestedCPU)
				totalLimitCPU.Add(limitCPU)
				totalRequestedMemory.Add(requestedMemory)
				totalLimitMemory.Add(limitMemory)
			}
		}

		// Convert totals to milli values and MiB
		requestedCPUInMillis := totalRequestedCPU.MilliValue()
		limitCPUInMillis := totalLimitCPU.MilliValue()
		requestedMemoryInMi := totalRequestedMemory.ScaledValue(resource.Mega)
		limitMemoryInMi := totalLimitMemory.ScaledValue(resource.Mega)

		pdf.SetFont("Arial", "", 8)

		x, y := pdf.GetXY()
		pdf.MultiCell(colWidth, 8, nodeNameWithStatus, "1", "L", false)
		height := pdf.GetY() - y

		pdf.SetXY(x+colWidth, y)
		pdf.CellFormat(colWidth, height, strconv.Itoa(int(allocatableCPUInMillis)), "1", 0, "C", false, 0, "")
		pdf.CellFormat(colWidth, height, strconv.Itoa(int(allocatableMemoryInMi)), "1", 0, "C", false, 0, "")
		pdf.CellFormat(colWidth, height, strconv.Itoa(int(requestedCPUInMillis)), "1", 0, "C", false, 0, "")
		pdf.CellFormat(colWidth, height, strconv.Itoa(int(limitCPUInMillis)), "1", 0, "C", false, 0, "")
		pdf.CellFormat(colWidth, height, strconv.Itoa(int(requestedMemoryInMi)), "1", 0, "C", false, 0, "")
		pdf.CellFormat(colWidth, height, strconv.Itoa(int(limitMemoryInMi)), "1", 1, "C", false, 0, "")
	}

	return nil
}
