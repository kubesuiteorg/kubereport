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

// StatefulSetInfo is used to hold the information for a Kubernetes StatefulSet.
type StatefulSetInfo struct {
	Name            string
	Namespace       string
	DesiredReplicas int32
	CurrentReplicas int32
	PodsReady       int32
	PodsDesired     int32
	ServiceName     string
	Age             string
	Conditions      string
}

// Generates a CSV report of Kubernetes StatefulSets.
func GenerateStatefulSetReportCSV(writer *csv.Writer, clientset *kubernetes.Clientset) error {
	// Fetch StatefulSets
	statefulSets, err := clientset.AppsV1().StatefulSets(metav1.NamespaceAll).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("error fetching StatefulSets: %v", err)
	}

	var statefulSetData []StatefulSetInfo

	for _, ss := range statefulSets.Items {
		age := time.Since(ss.CreationTimestamp.Time).Round(time.Hour).String()

		var conditions []string
		for _, cond := range ss.Status.Conditions {
			if cond.Status == "True" {
				conditions = append(conditions, fmt.Sprintf("%s", cond.Type))
			}
		}
		conditionsStr := strings.Join(conditions, ", ")

		// Get the associated service name
		serviceName := ss.Name + "-svc" // Assuming service naming convention

		// Create a record for the StatefulSet
		statefulSetData = append(statefulSetData, StatefulSetInfo{
			Name:            ss.Name,
			Namespace:       ss.Namespace,
			DesiredReplicas: *ss.Spec.Replicas,
			CurrentReplicas: ss.Status.Replicas,
			PodsReady:       ss.Status.ReadyReplicas,
			PodsDesired:     ss.Status.Replicas,
			ServiceName:     serviceName,
			Age:             age,
			Conditions:      conditionsStr,
		})
	}

	sort.Slice(statefulSetData, func(i, j int) bool {
		return statefulSetData[i].Name < statefulSetData[j].Name
	})

	// Write CSV headers
	if err := writer.Write([]string{
		"STATEFULSET NAME",
		"NAMESPACE",
		"DESIRED REPLICAS",
		"CURRENT REPLICAS",
		"PODS READY",
		"PODS DESIRED",
		"SERVICE NAME",
		"AGE",
		"CONDITIONS",
	}); err != nil {
		return fmt.Errorf("error writing headers to CSV: %v", err)
	}

	// Write StatefulSet data rows
	for _, ss := range statefulSetData {
		record := []string{
			ss.Name,
			ss.Namespace,
			strconv.Itoa(int(ss.DesiredReplicas)),
			strconv.Itoa(int(ss.CurrentReplicas)),
			strconv.Itoa(int(ss.PodsReady)),
			strconv.Itoa(int(ss.PodsDesired)),
			ss.ServiceName,
			ss.Age,
			ss.Conditions,
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
