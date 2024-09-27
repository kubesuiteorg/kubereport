package tables

import (
	"context"
	"fmt"
	"strconv"

	"github.com/jung-kurt/gofpdf/v2"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Generates a table for namespace resource usage.
func GenerateNamespaceTable(pdf *gofpdf.Fpdf, clientset *kubernetes.Clientset) error {

	pdf.SetFont("Arial", "B", 10)
	headers := []string{
		"Namespace",
		"CPU Req (mCPU)",
		"CPU Lim (mCPU)",
		"Memory Req (MiB)",
		"Memory Lim (MiB)",
	}
	colWidth := 38.0
	for _, header := range headers {
		pdf.CellFormat(colWidth, 8, header, "1", 0, "C", false, 0, "")
	}
	pdf.Ln(8)

	totalCPURequests := resource.NewMilliQuantity(0, resource.DecimalSI)
	totalCPULimits := resource.NewMilliQuantity(0, resource.DecimalSI)
	totalMemoryRequests := resource.NewQuantity(0, resource.BinarySI)
	totalMemoryLimits := resource.NewQuantity(0, resource.BinarySI)

	ctx := context.TODO()
	namespaces, err := clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list namespaces: %v", err)
	}

	for _, ns := range namespaces.Items {
		nsCPURequests := resource.NewMilliQuantity(0, resource.DecimalSI)
		nsCPULimits := resource.NewMilliQuantity(0, resource.DecimalSI)
		nsMemoryRequests := resource.NewQuantity(0, resource.BinarySI)
		nsMemoryLimits := resource.NewQuantity(0, resource.BinarySI)

		pods, err := clientset.CoreV1().Pods(ns.Name).List(ctx, metav1.ListOptions{})
		if err != nil {
			return fmt.Errorf("failed to list pods in namespace %s: %v", ns.Name, err)
		}

		for _, pod := range pods.Items {
			for _, container := range pod.Spec.Containers {
				if cpuRequest, ok := container.Resources.Requests["cpu"]; ok {
					nsCPURequests.Add(cpuRequest)
				}
				if cpuLimit, ok := container.Resources.Limits["cpu"]; ok {
					nsCPULimits.Add(cpuLimit)
				}
				if memRequest, ok := container.Resources.Requests["memory"]; ok {
					nsMemoryRequests.Add(memRequest)
				}
				if memLimit, ok := container.Resources.Limits["memory"]; ok {
					nsMemoryLimits.Add(memLimit)
				}
			}
		}

		totalCPURequests.Add(*nsCPURequests)
		totalCPULimits.Add(*nsCPULimits)
		totalMemoryRequests.Add(*nsMemoryRequests)
		totalMemoryLimits.Add(*nsMemoryLimits)

		rowHeight := 8.0
		_, pageHeight := pdf.GetPageSize()
		if pdf.GetY()+rowHeight > pageHeight-20 {
			pdf.AddPage()
			pdf.SetFont("Arial", "B", 10)
			for _, header := range headers {
				pdf.CellFormat(colWidth, 8, header, "1", 0, "C", false, 0, "")
			}
			pdf.Ln(8)
		}

		pdf.SetFont("Arial", "", 10)
		x, y := pdf.GetXY()
		pdf.MultiCell(colWidth, 8, ns.Name, "1", "L", false)
		height := pdf.GetY() - y

		pdf.SetXY(x+colWidth, y)
		pdf.CellFormat(colWidth, height, strconv.Itoa(int(nsCPURequests.MilliValue())), "1", 0, "C", false, 0, "")
		pdf.CellFormat(colWidth, height, strconv.Itoa(int(nsCPULimits.MilliValue())), "1", 0, "C", false, 0, "")
		pdf.CellFormat(colWidth, height, fmt.Sprintf("%.2f", float64(nsMemoryRequests.Value())/1024/1024), "1", 0, "C", false, 0, "")
		pdf.CellFormat(colWidth, height, fmt.Sprintf("%.2f", float64(nsMemoryLimits.Value())/1024/1024), "1", 1, "C", false, 0, "")
	}

	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(colWidth, 8, "Total", "1", 0, "L", false, 0, "")
	pdf.CellFormat(colWidth, 8, strconv.Itoa(int(totalCPURequests.MilliValue())), "1", 0, "C", false, 0, "")
	pdf.CellFormat(colWidth, 8, strconv.Itoa(int(totalCPULimits.MilliValue())), "1", 0, "C", false, 0, "")
	pdf.CellFormat(colWidth, 8, fmt.Sprintf("%.2f", float64(totalMemoryRequests.Value())/1024/1024), "1", 0, "C", false, 0, "")
	pdf.CellFormat(colWidth, 8, fmt.Sprintf("%.2f", float64(totalMemoryLimits.Value())/1024/1024), "1", 1, "C", false, 0, "")

	return nil
}
