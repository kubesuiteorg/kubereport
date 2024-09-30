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

func GenerateNamespaceTable(pdf *gofpdf.Fpdf, clientset *kubernetes.Clientset) error {
	printTableHeaders := func() {
		pdf.SetFont("Arial", "B", 8)
		headers := []string{
			"Namespace",
			"CPU Lim(mCPU)",
			"CPU Req(mCPU)",
			"Memory Lim(MiB)",
			"Memory Req(MiB)",
		}
		colWidths := []float64{90.0, 25.0, 25.0, 25.0, 25.0} // Set different widths for each column

		for i, header := range headers {
			pdf.CellFormat(colWidths[i], 8, header, "1", 0, "C", false, 0, "")
		}
		pdf.Ln(8)
	}

	printTableHeaders()

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
		margin := 20.0

		neededHeight := rowHeight
		startY := pdf.GetY()

		if startY+neededHeight > pageHeight-margin {
			pdf.AddPage()
			printTableHeaders() // Reprint headers on the new page
		}

		pdf.SetFont("Arial", "", 8)
		pdf.CellFormat(90.0, rowHeight, ns.Name, "1", 0, "L", false, 0, "")
		pdf.CellFormat(25.0, rowHeight, strconv.Itoa(int(nsCPULimits.MilliValue())), "1", 0, "C", false, 0, "")
		pdf.CellFormat(25.0, rowHeight, strconv.Itoa(int(nsCPURequests.MilliValue())), "1", 0, "C", false, 0, "")
		pdf.CellFormat(25.0, rowHeight, fmt.Sprintf("%.2f", float64(nsMemoryLimits.Value())/1024/1024), "1", 0, "C", false, 0, "")
		pdf.CellFormat(25.0, rowHeight, fmt.Sprintf("%.2f", float64(nsMemoryRequests.Value())/1024/1024), "1", 1, "C", false, 0, "")
	}

	pdf.SetFont("Arial", "B", 8)
	pdf.CellFormat(90.0, 8, "Total", "1", 0, "L", false, 0, "")
	pdf.CellFormat(25.0, 8, strconv.Itoa(int(totalCPULimits.MilliValue())), "1", 0, "C", false, 0, "")
	pdf.CellFormat(25.0, 8, strconv.Itoa(int(totalCPURequests.MilliValue())), "1", 0, "C", false, 0, "")
	pdf.CellFormat(25.0, 8, fmt.Sprintf("%.2f", float64(totalMemoryLimits.Value())/1024/1024), "1", 0, "C", false, 0, "")
	pdf.CellFormat(25.0, 8, fmt.Sprintf("%.2f", float64(totalMemoryRequests.Value())/1024/1024), "1", 1, "C", false, 0, "")

	return nil
}
