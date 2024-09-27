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

// ReplicaSetInfo holds the information for a Kubernetes ReplicaSet.
type ReplicaSetInfo struct {
	Name            string
	Namespace       string
	DesiredReplicas int32
	CurrentReplicas int32
	PodsReady       int32
	PodsDesired     int32
	Age             string
	Conditions      string
}

// Generates a CSV report of Kubernetes ReplicaSets.
func GenerateReplicaSetReportCSV(writer *csv.Writer, clientset *kubernetes.Clientset) error {
	// Fetch ReplicaSets
	replicaSets, err := clientset.AppsV1().ReplicaSets(metav1.NamespaceAll).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("error fetching ReplicaSets: %v", err)
	}

	var replicaSetData []ReplicaSetInfo

	// Iterate over ReplicaSets to get their information
	for _, rs := range replicaSets.Items {
		age := time.Since(rs.CreationTimestamp.Time).Round(time.Hour).String()

		var conditions []string
		for _, cond := range rs.Status.Conditions {
			fmt.Printf("Condition: %v\n", cond)
			if cond.Status == "True" {
				conditions = append(conditions, fmt.Sprintf("%s", cond.Type))
			}
		}
		conditionsStr := strings.Join(conditions, ", ")
		if conditionsStr == "" {
			conditionsStr = "No conditions met"
		}

		// Create a record for the ReplicaSet
		replicaSetData = append(replicaSetData, ReplicaSetInfo{
			Name:            rs.Name,
			Namespace:       rs.Namespace,
			DesiredReplicas: *rs.Spec.Replicas,
			CurrentReplicas: rs.Status.Replicas,
			PodsReady:       rs.Status.ReadyReplicas,
			PodsDesired:     rs.Status.Replicas,
			Age:             age,
			Conditions:      conditionsStr,
		})
	}

	sort.Slice(replicaSetData, func(i, j int) bool {
		return replicaSetData[i].Name < replicaSetData[j].Name
	})

	// Write CSV headers
	if err := writer.Write([]string{
		"REPLICASET NAME",
		"NAMESPACE",
		"DESIRED REPLICAS",
		"CURRENT REPLICAS",
		"PODS READY",
		"PODS DESIRED",
		"AGE",
		"CONDITIONS",
	}); err != nil {
		return fmt.Errorf("error writing headers to CSV: %v", err)
	}

	// Write ReplicaSet data rows
	for _, rs := range replicaSetData {
		record := []string{
			rs.Name,
			rs.Namespace,
			strconv.Itoa(int(rs.DesiredReplicas)),
			strconv.Itoa(int(rs.CurrentReplicas)),
			strconv.Itoa(int(rs.PodsReady)),
			strconv.Itoa(int(rs.PodsDesired)),
			rs.Age,
			rs.Conditions,
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
