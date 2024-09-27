package detailedreport

import (
	"context"
	"encoding/csv"
	"fmt"
	"strconv"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Generates a CSV file for namespace resource usage.
func GenerateNamespaceTable(writer *csv.Writer, clientset *kubernetes.Clientset) error {
	headers := []string{"NAMESPACE", "PODS", "RUNNING PODS", "PENDING PODS", "FAILED PODS", "SERVICES", "DEPLOYMENTS", "REPLICASETS", "STATEFULSETS", "DAEMONSETS", "CONFIGMAPS", "SECRETS", "ANNOTATIONS", "CPU REQ (MCPU)", "CPU LIM (MCPU)", "MEMORY REQ (MIB)", "MEMORY LIM (MIB)"}
	if err := writer.Write(headers); err != nil {
		return fmt.Errorf("failed to write header to CSV file: %v", err)
	}

	ctx := context.TODO()
	namespaces, err := clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list namespaces: %v", err)
	}

	for _, ns := range namespaces.Items {
		podCount := 0
		runningPods := 0
		pendingPods := 0
		failedPods := 0

		// List pods in the namespace
		pods, err := clientset.CoreV1().Pods(ns.Name).List(ctx, metav1.ListOptions{})
		if err != nil {
			return fmt.Errorf("failed to list pods in namespace %s: %v", ns.Name, err)
		}

		podCount = len(pods.Items)
		for _, pod := range pods.Items {
			switch pod.Status.Phase {
			case "Running":
				runningPods++
			case "Pending":
				pendingPods++
			case "Failed":
				failedPods++
			}
		}

		// Count resources
		services, _ := clientset.CoreV1().Services(ns.Name).List(ctx, metav1.ListOptions{})
		deployments, _ := clientset.AppsV1().Deployments(ns.Name).List(ctx, metav1.ListOptions{})
		replicaSets, _ := clientset.AppsV1().ReplicaSets(ns.Name).List(ctx, metav1.ListOptions{})
		statefulSets, _ := clientset.AppsV1().StatefulSets(ns.Name).List(ctx, metav1.ListOptions{})
		daemonSets, _ := clientset.AppsV1().DaemonSets(ns.Name).List(ctx, metav1.ListOptions{})
		configMaps, _ := clientset.CoreV1().ConfigMaps(ns.Name).List(ctx, metav1.ListOptions{})
		secrets, _ := clientset.CoreV1().Secrets(ns.Name).List(ctx, metav1.ListOptions{})

		annotations := fmt.Sprintf("%v", ns.Annotations)

		// Calculate resource requests and limits
		cpuReq := 0
		cpuLim := 0
		memReq := 0
		memLim := 0

		for _, pod := range pods.Items {
			for _, container := range pod.Spec.Containers {
				if req, ok := container.Resources.Requests["cpu"]; ok {
					cpuReq += int(req.MilliValue())
				}
				if lim, ok := container.Resources.Limits["cpu"]; ok {
					cpuLim += int(lim.MilliValue())
				}
				if req, ok := container.Resources.Requests["memory"]; ok {
					memReq += int(req.Value() / (1024 * 1024)) // Convert to MiB
				}
				if lim, ok := container.Resources.Limits["memory"]; ok {
					memLim += int(lim.Value() / (1024 * 1024)) // Convert to MiB
				}
			}
		}

		// Prepare the row for the CSV
		row := []string{
			ns.Name,
			strconv.Itoa(podCount),
			strconv.Itoa(runningPods),
			strconv.Itoa(pendingPods),
			strconv.Itoa(failedPods),
			strconv.Itoa(len(services.Items)),
			strconv.Itoa(len(deployments.Items)),
			strconv.Itoa(len(replicaSets.Items)),
			strconv.Itoa(len(statefulSets.Items)),
			strconv.Itoa(len(daemonSets.Items)),
			strconv.Itoa(len(configMaps.Items)),
			strconv.Itoa(len(secrets.Items)),
			annotations,
			strconv.Itoa(cpuReq),
			strconv.Itoa(cpuLim),
			strconv.Itoa(memReq),
			strconv.Itoa(memLim),
		}

		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write row to CSV file: %v", err)
		}
	}

	return nil
}
