package detailedreport

import (
	"context"
	"encoding/csv"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// DeploymentInfo holds the information for a Kubernetes deployment.
type DeploymentInfo struct {
	Name              string
	Namespace         string
	Replicas          int32
	AvailableReplicas int32
	PodsReady         int32
	PodsDesired       int32
	StrategyType      string
	Age               string
	Conditions        string
	Revision          int64
}

// Generates a CSV report of Kubernetes deployments.
func GenerateDeploymentReportCSV(writer *csv.Writer, clientset *kubernetes.Clientset) error {
	// Fetch deployments
	deployments, err := clientset.AppsV1().Deployments(metav1.NamespaceAll).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("error fetching deployments: %v", err)
	}

	var deploymentData []DeploymentInfo

	for _, deploy := range deployments.Items {
		age := time.Since(deploy.CreationTimestamp.Time).Round(time.Hour).String()

		var conditions []string
		for _, cond := range deploy.Status.Conditions {
			conditions = append(conditions, fmt.Sprintf("%s", cond.Type))
		}
		conditionsStr := strings.Join(conditions, ", ")

		deploymentData = append(deploymentData, DeploymentInfo{
			Name:              deploy.Name,
			Namespace:         deploy.Namespace,
			Replicas:          *deploy.Spec.Replicas,
			AvailableReplicas: deploy.Status.AvailableReplicas,
			PodsReady:         deploy.Status.ReadyReplicas,
			PodsDesired:       deploy.Status.Replicas,
			StrategyType:      string(deploy.Spec.Strategy.Type),
			Age:               age,
			Conditions:        conditionsStr,
			Revision:          deploy.Status.ObservedGeneration,
		})
	}

	sort.Slice(deploymentData, func(i, j int) bool {
		return deploymentData[i].Name < deploymentData[j].Name
	})

	// Write CSV headers
	if err := writer.Write([]string{
		"DEPLOYMENT NAME",
		"NAMESPACE",
		"REPLICAS",
		"AVAILABLE REPLICAS",
		"PODS READY",
		"PODS DESIRED",
		"STRATEGY TYPE",
		"REVISION",
		"AGE",
		"CONDITIONS",
	}); err != nil {
		return fmt.Errorf("error writing headers to CSV: %v", err)
	}

	// Write deployment data rows
	for _, deploy := range deploymentData {
		record := []string{
			deploy.Name,
			deploy.Namespace,
			strconv.Itoa(int(deploy.Replicas)),
			strconv.Itoa(int(deploy.AvailableReplicas)),
			strconv.Itoa(int(deploy.PodsReady)),
			strconv.Itoa(int(deploy.PodsDesired)),
			deploy.StrategyType,
			strconv.FormatInt(deploy.Revision, 10),
			deploy.Age,
			deploy.Conditions,
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("error writing record to CSV: %v", err)
		}
	}

	// Flush the writer
	writer.Flush()
	if err := writer.Error(); err != nil {
		return fmt.Errorf("error flushing CSV writer: %v", err)
	}

	return nil
}
