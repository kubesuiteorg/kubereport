package detailedreport

import (
	"context"
	"encoding/csv"
	"fmt"
	"sort"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// RoleInfo holds the information for a Kubernetes Role.
type RoleInfo struct {
	Name        string
	Namespace   string
	Rules       string
	Age         string
	Annotations string
}

// Generates a CSV report of Kubernetes Roles.
func GenerateRoleReportCSV(writer *csv.Writer, clientset *kubernetes.Clientset) error {
	// Fetch Roles from all namespaces
	roles, err := clientset.RbacV1().Roles(metav1.NamespaceAll).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("error fetching Roles: %v", err)
	}

	var roleData []RoleInfo

	for _, role := range roles.Items {
		age := time.Since(role.CreationTimestamp.Time).Round(time.Hour).String()

		// Rules - Simplified for display
		var rules []string
		for _, rule := range role.Rules {
			verbs := strings.Join(rule.Verbs, ", ")
			resources := strings.Join(rule.Resources, ", ")
			rules = append(rules, fmt.Sprintf("[%s] on [%s]", verbs, resources))
		}
		rulesStr := strings.Join(rules, "; ")

		var annotations string
		for key, value := range role.Annotations {
			annotations += fmt.Sprintf("%s=%s, ", key, value)
		}
		if len(annotations) > 0 {
			annotations = annotations[:len(annotations)-2]
		}

		roleInfo := RoleInfo{
			Name:        role.Name,
			Namespace:   role.Namespace,
			Rules:       rulesStr,
			Age:         age,
			Annotations: annotations,
		}

		roleData = append(roleData, roleInfo)
	}

	sort.Slice(roleData, func(i, j int) bool {
		if roleData[i].Name == roleData[j].Name {
			return roleData[i].Namespace < roleData[j].Namespace
		}
		return roleData[i].Name < roleData[j].Name
	})

	// Write CSV headers
	if err := writer.Write([]string{
		"ROLE NAME",
		"NAMESPACE",
		"RULES",
		"AGE",
		"ANNOTATIONS",
	}); err != nil {
		return fmt.Errorf("error writing headers to CSV: %v", err)
	}

	// Write Role data rows
	for _, role := range roleData {
		record := []string{
			role.Name,
			role.Namespace,
			role.Rules,
			role.Age,
			role.Annotations,
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
