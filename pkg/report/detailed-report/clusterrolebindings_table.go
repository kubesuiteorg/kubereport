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

// ClusterRoleBindingInfo holds the information for a Kubernetes ClusterRoleBinding.
type ClusterRoleBindingInfo struct {
	Name            string
	ClusterRoleName string
	Subjects        string
	RoleRefAPIGroup string
	RoleRefKind     string
	RoleRefName     string
	Age             string
	Annotations     string
}

// Generates a CSV report of Kubernetes ClusterRoleBindings.
func GenerateClusterRoleBindingReportCSV(writer *csv.Writer, clientset *kubernetes.Clientset) error {
	// Fetch ClusterRoleBindings from all namespaces
	clusterRoleBindings, err := clientset.RbacV1().ClusterRoleBindings().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("error fetching ClusterRoleBindings: %v", err)
	}

	var clusterRoleBindingData []ClusterRoleBindingInfo

	for _, binding := range clusterRoleBindings.Items {
		age := time.Since(binding.CreationTimestamp.Time).Round(time.Hour).String()

		var subjects []string
		for _, subject := range binding.Subjects {
			subjects = append(subjects, fmt.Sprintf("%s/%s", subject.Kind, subject.Name))
		}
		subjectsStr := "[" + strings.Join(subjects, ", ") + "]"

		var annotations string
		for key, value := range binding.Annotations {
			annotations += fmt.Sprintf("%s=%s, ", key, value)
		}
		if len(annotations) > 0 {
			annotations = annotations[:len(annotations)-2]
		}

		clusterRoleBindingInfo := ClusterRoleBindingInfo{
			Name:            binding.Name,
			ClusterRoleName: binding.RoleRef.Name,
			Subjects:        subjectsStr,
			RoleRefAPIGroup: binding.RoleRef.APIGroup,
			RoleRefKind:     binding.RoleRef.Kind,
			RoleRefName:     binding.RoleRef.Name,
			Age:             age,
			Annotations:     annotations,
		}

		clusterRoleBindingData = append(clusterRoleBindingData, clusterRoleBindingInfo)
	}

	// Write CSV headers
	if err := writer.Write([]string{
		"CLUSTERROLEBINDING NAME",
		"CLUSTERROLE NAME",
		"SUBJECTS",
		"ROLEREF API GROUP",
		"ROLEREF KIND",
		"ROLEREF NAME",
		"AGE",
		"ANNOTATIONS",
	}); err != nil {
		return fmt.Errorf("error writing headers to CSV: %v", err)
	}

	// Write ClusterRoleBinding data rows
	for _, binding := range clusterRoleBindingData {
		record := []string{
			binding.Name,
			binding.ClusterRoleName,
			binding.Subjects,
			binding.RoleRefAPIGroup,
			binding.RoleRefKind,
			binding.RoleRefName,
			binding.Age,
			binding.Annotations,
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
