package detailedreport

import (
	"context"
	"encoding/csv"
	"fmt"
	"sort"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// IngressResourceInfo holds the information for a Kubernetes Ingress resource.
type IngressResourceInfo struct {
	Name               string
	Namespace          string
	Hosts              string
	Paths              string
	BackendServiceName string
	BackendServicePort string
	TLSEnabled         string
	TLSSecretName      string
	IngressClass       string
	Rules              string
	Age                string
	Annotations        string
}

// Generates a CSV report of Kubernetes Ingress resources.
func GenerateIngressReportCSV(writer *csv.Writer, clientset *kubernetes.Clientset) error {
	// Fetch Ingress resources
	ingresses, err := clientset.NetworkingV1().Ingresses("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("error fetching Ingress resources: %v", err)
	}

	var ingressData []IngressResourceInfo

	for _, ingress := range ingresses.Items {
		age := time.Since(ingress.CreationTimestamp.Time).Round(time.Hour).String()

		// Prepare Hosts and Paths
		var hosts, paths, backendServiceName, backendServicePort, tlsEnabled, tlsSecretName, ingressClass, rules string
		for _, rule := range ingress.Spec.Rules {
			hosts += fmt.Sprintf("%s,", rule.Host)
			for _, path := range rule.HTTP.Paths {
				paths += fmt.Sprintf("%s,", path.Path)
				backendServiceName += fmt.Sprintf("%s,", path.Backend.Service.Name)
				backendServicePort += fmt.Sprintf("%d,", path.Backend.Service.Port.Number)
			}
			rules += fmt.Sprintf("%s: %s,", rule.Host, paths)
		}

		if len(hosts) > 0 {
			hosts = hosts[:len(hosts)-1]
		}
		if len(paths) > 0 {
			paths = paths[:len(paths)-1]
		}
		if len(backendServiceName) > 0 {
			backendServiceName = backendServiceName[:len(backendServiceName)-1]
		}
		if len(backendServicePort) > 0 {
			backendServicePort = backendServicePort[:len(backendServicePort)-1]
		}
		if len(rules) > 0 {
			rules = rules[:len(rules)-1]
		}

		// Check if TLS is enabled and get TLS secret name
		if len(ingress.Spec.TLS) > 0 {
			tlsEnabled = "Yes"
			tlsSecretName = ingress.Spec.TLS[0].SecretName // Assuming one TLS secret per Ingress
		} else {
			tlsEnabled = "No"
		}

		// Get Ingress Class from annotations
		ingressClass = ingress.Annotations["kubernetes.io/ingress.class"]

		annotations := fmt.Sprintf("%v", ingress.Annotations)

		// Create a record for the Ingress resource
		ingressData = append(ingressData, IngressResourceInfo{
			Name:               ingress.Name,
			Namespace:          ingress.Namespace,
			Hosts:              hosts,
			Paths:              paths,
			BackendServiceName: backendServiceName,
			BackendServicePort: backendServicePort,
			TLSEnabled:         tlsEnabled,
			TLSSecretName:      tlsSecretName,
			IngressClass:       ingressClass,
			Rules:              rules,
			Age:                age,
			Annotations:        annotations,
		})
	}

	sort.Slice(ingressData, func(i, j int) bool {
		return ingressData[i].Name < ingressData[j].Name
	})

	// Write CSV headers
	if err := writer.Write([]string{
		"INGRESS NAME",
		"NAMESPACE",
		"HOST(S)",
		"PATH(S)",
		"BACKEND SERVICE NAME",
		"BACKEND SERVICE PORT",
		"TLS ENABLED",
		"TLS SECRET NAME",
		"INGRESS CLASS",
		"RULES",
		"AGE",
		"ANNOTATIONS",
	}); err != nil {
		return fmt.Errorf("error writing headers to CSV: %v", err)
	}

	// Write Ingress resource data rows
	for _, ingress := range ingressData {
		record := []string{
			ingress.Name,
			ingress.Namespace,
			ingress.Hosts,
			ingress.Paths,
			ingress.BackendServiceName,
			ingress.BackendServicePort,
			ingress.TLSEnabled,
			ingress.TLSSecretName,
			ingress.IngressClass,
			ingress.Rules,
			ingress.Age,
			ingress.Annotations,
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
