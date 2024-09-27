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

// ServiceInfo holds the information for a Kubernetes service.
type ServiceInfo struct {
	Name            string
	Namespace       string
	ServiceType     string
	ClusterIP       string
	ExternalIP      string
	Ports           string
	TargetPort      string
	Selector        string
	SessionAffinity string
	Age             string
	Conditions      string
}

// Generates a CSV report of Kubernetes services.
func GenerateServiceReportCSV(writer *csv.Writer, clientset *kubernetes.Clientset) error {
	// Fetch services
	services, err := clientset.CoreV1().Services(metav1.NamespaceAll).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("error fetching services: %v", err)
	}

	var serviceData []ServiceInfo

	for _, svc := range services.Items {
		age := time.Since(svc.CreationTimestamp.Time).Round(time.Hour).String()

		// Prepare ports and target ports
		var ports []string
		var targetPorts []string
		for _, port := range svc.Spec.Ports {
			ports = append(ports, fmt.Sprintf("%d/%s", port.Port, port.Protocol))
			targetPorts = append(targetPorts, strconv.Itoa(int(port.TargetPort.IntVal)))
		}

		// Get External IPs from LoadBalancer Ingress
		var externalIPs []string
		for _, ingress := range svc.Status.LoadBalancer.Ingress {
			if ingress.IP != "" {
				externalIPs = append(externalIPs, ingress.IP)
			}
		}
		externalIPStr := strings.Join(externalIPs, ", ")

		serviceData = append(serviceData, ServiceInfo{
			Name:            svc.Name,
			Namespace:       svc.Namespace,
			ServiceType:     string(svc.Spec.Type),
			ClusterIP:       svc.Spec.ClusterIP,
			ExternalIP:      externalIPStr,
			Ports:           strings.Join(ports, ", "),
			TargetPort:      strings.Join(targetPorts, ", "),
			Selector:        formatSelector(svc.Spec.Selector),
			SessionAffinity: string(svc.Spec.SessionAffinity),
			Age:             age,
			Conditions:      "Active", // Assuming all services are active for the sake of simplicity
		})
	}

	sort.Slice(serviceData, func(i, j int) bool {
		return serviceData[i].Name < serviceData[j].Name
	})

	// Write CSV headers
	if err := writer.Write([]string{
		"SERVICE NAME",
		"NAMESPACE",
		"TYPE",
		"CLUSTER IP",
		"EXTERNAL IP",
		"PORT(S)",
		"TARGET PORT",
		"SELECTOR",
		"SESSION AFFINITY",
		"AGE",
		"CONDITIONS",
	}); err != nil {
		return fmt.Errorf("error writing headers to CSV: %v", err)
	}

	// Write service data rows
	for _, svc := range serviceData {
		record := []string{
			svc.Name,
			svc.Namespace,
			svc.ServiceType,
			svc.ClusterIP,
			svc.ExternalIP,
			svc.Ports,
			svc.TargetPort,
			svc.Selector,
			svc.SessionAffinity,
			svc.Age,
			svc.Conditions,
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

func formatSelector(selector map[string]string) string {
	var parts []string
	for key, value := range selector {
		parts = append(parts, fmt.Sprintf("%s=%s", key, value))
	}
	return strings.Join(parts, ", ")
}
