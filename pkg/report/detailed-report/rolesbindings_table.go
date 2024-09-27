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

// RoleBindingInfo holds the information for a Kubernetes RoleBinding.
type RoleBindingInfo struct {
	Name        string
	Namespace   string
	RoleName    string
	Subjects    string
	Kind        string
	APIGroup    string
	Age         string
	Annotations string
}

// Generates a CSV report of Kubernetes RoleBindings.
func GenerateRoleBindingReportCSV(writer *csv.Writer, clientset *kubernetes.Clientset) error {
	// Fetch RoleBindings from all namespaces
	roleBindings, err := clientset.RbacV1().RoleBindings(metav1.NamespaceAll).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("error fetching RoleBindings: %v", err)
	}

	var roleBindingData []RoleBindingInfo

	for _, roleBinding := range roleBindings.Items {
		age := time.Since(roleBinding.CreationTimestamp.Time).Round(time.Hour).String()

		roleName := roleBinding.RoleRef.Name

		// Subjects - Simplified for display
		var subjects []string
		for _, subject := range roleBinding.Subjects {
			subjects = append(subjects, subject.Name)
		}
		subjectsStr := "[" + strings.Join(subjects, ", ") + "]"

		kind := roleBinding.RoleRef.Kind
		apiGroup := roleBinding.RoleRef.APIGroup

		var annotations string
		for key, value := range roleBinding.Annotations {
			annotations += fmt.Sprintf("%s=%s, ", key, value)
		}
		if len(annotations) > 0 {
			annotations = annotations[:len(annotations)-2]
		}

		roleBindingInfo := RoleBindingInfo{
			Name:        roleBinding.Name,
			Namespace:   roleBinding.Namespace,
			RoleName:    roleName,
			Subjects:    subjectsStr,
			Kind:        kind,
			APIGroup:    apiGroup,
			Age:         age,
			Annotations: annotations,
		}

		roleBindingData = append(roleBindingData, roleBindingInfo)
	}

	sort.Slice(roleBindingData, func(i, j int) bool {
		if roleBindingData[i].Name == roleBindingData[j].Name {
			return roleBindingData[i].Namespace < roleBindingData[j].Namespace
		}
		return roleBindingData[i].Name < roleBindingData[j].Name
	})

	// Write CSV headers
	if err := writer.Write([]string{
		"ROLEBINDING NAME",
		"NAMESPACE",
		"ROLE NAME",
		"SUBJECTS",
		"KIND",
		"API GROUP",
		"AGE",
		"ANNOTATIONS",
	}); err != nil {
		return fmt.Errorf("error writing headers to CSV: %v", err)
	}

	// Write RoleBinding data rows
	for _, roleBinding := range roleBindingData {
		record := []string{
			roleBinding.Name,
			roleBinding.Namespace,
			roleBinding.RoleName,
			roleBinding.Subjects,
			roleBinding.Kind,
			roleBinding.APIGroup,
			roleBinding.Age,
			roleBinding.Annotations,
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
