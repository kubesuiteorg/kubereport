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

// NetworkPolicyInfo holds the information for a Kubernetes Network Policy.
type NetworkPolicyInfo struct {
	Name              string
	Namespace         string
	PodSelector       string
	NamespaceSelector string
	PolicyTypes       string
	IngressRules      string
	EgressRules       string
	IngressAction     string
	EgressAction      string
	MatchLabels       string
	Age               string
	Annotations       string
}

// Generates a CSV report of Kubernetes Network Policies.
func GenerateNetworkPolicyReportCSV(writer *csv.Writer, clientset *kubernetes.Clientset) error {
	// Fetch Network Policies
	nps, err := clientset.NetworkingV1().NetworkPolicies("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("error fetching Network Policies: %v", err)
	}

	var npData []NetworkPolicyInfo

	// Iterate over Network Policies to get their information
	for _, np := range nps.Items {
		age := time.Since(np.CreationTimestamp.Time).Round(time.Hour).String()

		// Prepare Pod Selector
		podSelector := fmt.Sprintf("%v", np.Spec.PodSelector)

		// Prepare Namespace Selector
		namespaceSelector := ""
		if np.Spec.PodSelector.MatchLabels != nil {
			namespaceSelector = fmt.Sprintf("%v", np.Spec.PodSelector.MatchLabels)
		}

		// Prepare Policy Types
		var policyTypes string
		for _, policyType := range np.Spec.PolicyTypes {
			policyTypes += string(policyType) + ", "
		}
		if len(policyTypes) > 0 {
			policyTypes = policyTypes[:len(policyTypes)-2]
		}

		// Prepare Ingress Rules
		var ingressRules, ingressAction string
		if len(np.Spec.Ingress) > 0 {
			for _, rule := range np.Spec.Ingress {
				ingressRules += fmt.Sprintf("Allow from %v; ", rule.From)
				ingressAction = "Allow"
			}
		} else {
			ingressRules = "Deny from all"
			ingressAction = "Deny"
		}

		// Prepare Egress Rules
		var egressRules, egressAction string
		if len(np.Spec.Egress) > 0 {
			for _, rule := range np.Spec.Egress {
				egressRules += fmt.Sprintf("Allow to %v; ", rule.To)
				egressAction = "Allow"
			}
		} else {
			egressRules = "Deny to all"
			egressAction = "Deny"
		}

		matchLabels := fmt.Sprintf("%v", np.Spec.PodSelector.MatchLabels)

		annotations := fmt.Sprintf("%v", np.Annotations)

		// Create a record for the Network Policy
		npData = append(npData, NetworkPolicyInfo{
			Name:              np.Name,
			Namespace:         np.Namespace,
			PodSelector:       podSelector,
			NamespaceSelector: namespaceSelector,
			PolicyTypes:       policyTypes,
			IngressRules:      ingressRules,
			EgressRules:       egressRules,
			IngressAction:     ingressAction,
			EgressAction:      egressAction,
			MatchLabels:       matchLabels,
			Age:               age,
			Annotations:       annotations,
		})
	}

	sort.Slice(npData, func(i, j int) bool {
		return npData[i].Name < npData[j].Name
	})

	// Write CSV headers
	if err := writer.Write([]string{
		"NETWORK POLICY NAME",
		"NAMESPACE",
		"POD SELECTOR",
		"NAMESPACE SELECTOR",
		"POLICY TYPES",
		"INGRESS RULES",
		"EGRESS RULES",
		"INGRESS ACTION",
		"EGRESS ACTION",
		"MATCH LABELS",
		"AGE",
		"ANNOTATIONS",
	}); err != nil {
		return fmt.Errorf("error writing headers to CSV: %v", err)
	}

	// Write Network Policy data rows
	for _, np := range npData {
		record := []string{
			np.Name,
			np.Namespace,
			np.PodSelector,
			np.NamespaceSelector,
			np.PolicyTypes,
			np.IngressRules,
			np.EgressRules,
			np.IngressAction,
			np.EgressAction,
			np.MatchLabels,
			np.Age,
			np.Annotations,
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
