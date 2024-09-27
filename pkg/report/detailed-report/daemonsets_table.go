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

// DaemonSetInfo holds the information for a Kubernetes DaemonSet.
type DaemonSetInfo struct {
	Name         string
	Namespace    string
	DesiredPods  int32
	CurrentPods  int32
	PodsReady    int32
	PodsDesired  int32
	NodeSelector string
	Age          string
	Conditions   string
}

// Generates a CSV report of Kubernetes DaemonSets.
func GenerateDaemonSetReportCSV(writer *csv.Writer, clientset *kubernetes.Clientset) error {
	// Fetch DaemonSets
	daemonSets, err := clientset.AppsV1().DaemonSets(metav1.NamespaceAll).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("error fetching DaemonSets: %v", err)
	}

	var daemonSetData []DaemonSetInfo

	for _, ds := range daemonSets.Items {
		age := time.Since(ds.CreationTimestamp.Time).Round(time.Hour).String()

		var conditions []string
		for _, cond := range ds.Status.Conditions {
			conditions = append(conditions, fmt.Sprintf("%s", cond.Type))
		}
		conditionsStr := strings.Join(conditions, ", ")

		// Get node selector from template
		nodeSelector := ""
		if ds.Spec.Template.Spec.NodeSelector != nil {
			nodeSelector = fmt.Sprintf("%v", ds.Spec.Template.Spec.NodeSelector)
		}

		daemonSetData = append(daemonSetData, DaemonSetInfo{
			Name:         ds.Name,
			Namespace:    ds.Namespace,
			DesiredPods:  ds.Status.DesiredNumberScheduled,
			CurrentPods:  ds.Status.CurrentNumberScheduled,
			PodsReady:    ds.Status.NumberReady,
			PodsDesired:  ds.Status.DesiredNumberScheduled,
			NodeSelector: nodeSelector,
			Age:          age,
			Conditions:   conditionsStr,
		})
	}

	sort.Slice(daemonSetData, func(i, j int) bool {
		return daemonSetData[i].Name < daemonSetData[j].Name
	})

	// Write CSV headers
	if err := writer.Write([]string{
		"DAEMONSET NAME",
		"NAMESPACE",
		"DESIRED PODS",
		"CURRENT PODS",
		"PODS READY",
		"PODS DESIRED",
		"NODE SELECTOR",
		"AGE",
		"CONDITIONS",
	}); err != nil {
		return fmt.Errorf("error writing headers to CSV: %v", err)
	}

	// Write DaemonSet data rows
	for _, ds := range daemonSetData {
		record := []string{
			ds.Name,
			ds.Namespace,
			strconv.Itoa(int(ds.DesiredPods)),
			strconv.Itoa(int(ds.CurrentPods)),
			strconv.Itoa(int(ds.PodsReady)),
			strconv.Itoa(int(ds.PodsDesired)),
			ds.NodeSelector,
			ds.Age,
			ds.Conditions,
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
