package detailedreport

import (
	"context"
	"encoding/csv"
	"fmt"
	"strconv"
	"time"

	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Generates a CSV file for node resource usage.
func GenerateNodeSummaryTable(writer *csv.Writer, clientset *kubernetes.Clientset) error {
	headers := []string{"NODE NAME", "STATUS", "SCHEDULABLE", "ROLES", "CPU CAPACITY", "CPU REQUESTS", "CPU LIMITS", "MEMORY CAPACITY", "MEMORY REQUESTS", "MEMORY LIMITS", "DISK CAPACITY", "DISK USAGE", "NODE AGE", "POD COUNT", "CONDITIONS", "TAINTS"}
	if err := writer.Write(headers); err != nil {
		return fmt.Errorf("failed to write header to CSV file: %v", err)
	}

	ctx := context.TODO()
	nodes, err := clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list nodes: %v", err)
	}

	for _, node := range nodes.Items {
		nodeName := node.Name

		// Determine node health status
		nodeStatus := "Unknown"
		for _, condition := range node.Status.Conditions {
			if condition.Type == "Ready" {
				if condition.Status == "True" {
					nodeStatus = "Healthy"
				} else {
					nodeStatus = "Unhealthy"
				}
				break
			}
		}

		schedulable := "Yes"
		if node.Spec.Unschedulable {
			schedulable = "No"
		}

		roles := "worker"
		if _, ok := node.Labels["node-role.kubernetes.io/master"]; ok {
			roles = "master"
		}

		cpuCapacity := node.Status.Capacity.Cpu().MilliValue()
		memoryCapacity := node.Status.Capacity.Memory().Value()
		diskCapacity := node.Status.Capacity.StorageEphemeral().Value()

		cpuRequests := resource.NewMilliQuantity(0, resource.DecimalSI)
		cpuLimits := resource.NewMilliQuantity(0, resource.DecimalSI)
		memoryRequests := resource.NewQuantity(0, resource.BinarySI)
		memoryLimits := resource.NewQuantity(0, resource.BinarySI)
		diskUsage := resource.NewQuantity(0, resource.BinarySI)

		pods, err := clientset.CoreV1().Pods(metav1.NamespaceAll).List(ctx, metav1.ListOptions{
			FieldSelector: fmt.Sprintf("spec.nodeName=%s", nodeName),
		})
		if err != nil {
			return fmt.Errorf("failed to list pods on node %s: %v", nodeName, err)
		}

		for _, pod := range pods.Items {
			for _, container := range pod.Spec.Containers {
				if cpuRequest, ok := container.Resources.Requests["cpu"]; ok {
					cpuRequests.Add(cpuRequest)
				}
				if cpuLimit, ok := container.Resources.Limits["cpu"]; ok {
					cpuLimits.Add(cpuLimit)
				}
				if memRequest, ok := container.Resources.Requests["memory"]; ok {
					memoryRequests.Add(memRequest)
				}
				if memLimit, ok := container.Resources.Limits["memory"]; ok {
					memoryLimits.Add(memLimit)
				}
			}
		}

		nodeAge := time.Since(node.CreationTimestamp.Time).Round(time.Hour).String()
		podCount := len(pods.Items)

		conditions := ""
		for _, condition := range node.Status.Conditions {
			if conditions != "" {
				conditions += ", "
			}
			conditions += fmt.Sprintf("%s=%s", condition.Type, condition.Status)
		}

		taints := ""
		for _, taint := range node.Spec.Taints {
			if taints != "" {
				taints += ", "
			}
			taints += fmt.Sprintf("%s=%s:%s", taint.Key, taint.Value, taint.Effect)
		}

		row := []string{
			nodeName,
			nodeStatus,
			schedulable,
			roles,
			strconv.Itoa(int(cpuCapacity)) + "m",
			strconv.Itoa(int(cpuRequests.MilliValue())) + "m",
			strconv.Itoa(int(cpuLimits.MilliValue())) + "m",
			fmt.Sprintf("%.2fGi", float64(memoryCapacity)/1024/1024/1024),
			fmt.Sprintf("%.2fGi", float64(memoryRequests.Value())/1024/1024/1024),
			fmt.Sprintf("%.2fGi", float64(memoryLimits.Value())/1024/1024/1024),
			fmt.Sprintf("%.2fGi", float64(diskCapacity)/1024/1024/1024),
			fmt.Sprintf("%.2fGi", float64(diskUsage.Value())/1024/1024/1024),
			nodeAge,
			strconv.Itoa(podCount),
			conditions,
			taints,
		}

		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write row to CSV file: %v", err)
		}
	}

	return nil
}
