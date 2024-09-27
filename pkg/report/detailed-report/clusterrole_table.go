package detailedreport

import (
	"context"
	"encoding/csv"
	"fmt"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// ClusterRoleInfo holds the information for a Kubernetes ClusterRole.
type ClusterRoleInfo struct {
	Name        string
	Rules       string
	APIGroups   string
	Resources   string
	Verbs       string
	Age         string
	Annotations string
}

// Generates a CSV report of Kubernetes ClusterRoles.
func GenerateClusterRoleReportCSV(writer *csv.Writer, clientset *kubernetes.Clientset) error {
	// Fetch ClusterRoles from all namespaces
	clusterRoles, err := clientset.RbacV1().ClusterRoles().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("error fetching ClusterRoles: %v", err)
	}

	var clusterRoleData []ClusterRoleInfo

	for _, clusterRole := range clusterRoles.Items {
		age := time.Since(clusterRole.CreationTimestamp.Time).Round(time.Hour).String()

		// Rules - Simplified for display
		var rules []string
		for _, rule := range clusterRole.Rules {
			apiGroups := "[" + strings.Join(rule.APIGroups, ", ") + "]"
			resources := "[" + strings.Join(rule.Resources, ", ") + "]"
			verbs := "[" + strings.Join(rule.Verbs, ", ") + "]"
			rules = append(rules, fmt.Sprintf("APIGroups: %s, Resources: %s, Verbs: %s", apiGroups, resources, verbs))
		}
		rulesStr := "[" + strings.Join(rules, "; ") + "]"

		var annotations string
		for key, value := range clusterRole.Annotations {
			annotations += fmt.Sprintf("%s=%s, ", key, value)
		}
		if len(annotations) > 0 {
			annotations = annotations[:len(annotations)-2]
		}

		clusterRoleInfo := ClusterRoleInfo{
			Name:        clusterRole.Name,
			Rules:       rulesStr,
			APIGroups:   rulesStr,
			Resources:   rulesStr,
			Verbs:       rulesStr,
			Age:         age,
			Annotations: annotations,
		}

		clusterRoleData = append(clusterRoleData, clusterRoleInfo)
	}

	// Write CSV headers
	if err := writer.Write([]string{
		"CLUSTERROLE NAME",
		"RULES",
		"API GROUPS",
		"RESOURCES",
		"VERBS",
		"AGE",
		"ANNOTATIONS",
	}); err != nil {
		return fmt.Errorf("error writing headers to CSV: %v", err)
	}

	// Write ClusterRole data rows
	for _, clusterRole := range clusterRoleData {
		record := []string{
			clusterRole.Name,
			clusterRole.Rules,
			clusterRole.APIGroups,
			clusterRole.Resources,
			clusterRole.Verbs,
			clusterRole.Age,
			clusterRole.Annotations,
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
