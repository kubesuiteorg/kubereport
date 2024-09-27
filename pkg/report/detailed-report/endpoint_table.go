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

// EndpointInfo holds the information for a Kubernetes Endpoint.
type EndpointInfo struct {
	Name        string
	Namespace   string
	Subsets     int
	IPAddresses string
	Ports       string
	Age         string
}

// Generates a CSV report of Kubernetes Endpoints.
func GenerateEndpointsReportCSV(writer *csv.Writer, clientset *kubernetes.Clientset) error {
	// Fetch Endpoints from all namespaces
	endpointsList, err := clientset.CoreV1().Endpoints("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("error fetching Endpoints: %v", err)
	}

	var endpointData []EndpointInfo

	for _, endpoint := range endpointsList.Items {
		age := time.Since(endpoint.CreationTimestamp.Time).Round(time.Hour).String()

		subsetsCount := len(endpoint.Subsets)

		var ipAddresses []string
		var ports []string

		for _, subset := range endpoint.Subsets {
			for _, address := range subset.Addresses {
				ipAddresses = append(ipAddresses, address.IP)
			}
			for _, port := range subset.Ports {
				ports = append(ports, fmt.Sprintf("%d/%s", port.Port, port.Protocol))
			}
		}

		ipAddressesStr := "[" + strings.Join(ipAddresses, ", ") + "]"
		portsStr := strings.Join(ports, ", ")

		endpointInfo := EndpointInfo{
			Name:        endpoint.Name,
			Namespace:   endpoint.Namespace,
			Subsets:     subsetsCount,
			IPAddresses: ipAddressesStr,
			Ports:       portsStr,
			Age:         age,
		}

		endpointData = append(endpointData, endpointInfo)
	}

	// Write CSV headers
	if err := writer.Write([]string{
		"ENDPOINT NAME",
		"NAMESPACE",
		"SUBSETS",
		"IP ADDRESSES",
		"PORTS",
		"AGE",
	}); err != nil {
		return fmt.Errorf("error writing headers to CSV: %v", err)
	}

	// Write Endpoint data rows
	for _, ep := range endpointData {
		record := []string{
			ep.Name,
			ep.Namespace,
			fmt.Sprintf("%d", ep.Subsets),
			ep.IPAddresses,
			ep.Ports,
			ep.Age,
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
