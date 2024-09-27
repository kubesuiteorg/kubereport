package detailedreport

import (
	"context"
	"encoding/csv"
	"fmt"
	"sort"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// PersistentVolumeClaimInfo holds the information for a Kubernetes Persistent Volume Claim.
type PersistentVolumeClaimInfo struct {
	Name         string
	Namespace    string
	Status       string
	Volume       string
	Capacity     string
	AccessModes  string
	StorageClass string
	Age          string
	VolumeMode   string
	Annotations  string
	Selector     string
}

// Generates a CSV report of Kubernetes Persistent Volume Claims.
func GeneratePersistentVolumeClaimReportCSV(writer *csv.Writer, clientset *kubernetes.Clientset) error {
	// Fetch Persistent Volume Claims
	pvcs, err := clientset.CoreV1().PersistentVolumeClaims("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("error fetching Persistent Volume Claims: %v", err)
	}

	var pvcData []PersistentVolumeClaimInfo

	// Iterate over Persistent Volume Claims to get their information
	for _, pvc := range pvcs.Items {
		age := time.Since(pvc.CreationTimestamp.Time).Round(time.Hour).String()

		accessModes := fmt.Sprintf("%v", pvc.Spec.AccessModes)

		volumeMode := ""
		if pvc.Spec.VolumeMode != nil {
			volumeMode = string(*pvc.Spec.VolumeMode)
		}

		annotations := fmt.Sprintf("%v", pvc.Annotations)

		selector := ""
		if pvc.Spec.Selector != nil {
			selector = metav1.FormatLabelSelector(pvc.Spec.Selector)
		}

		storageClass := ""
		if pvc.Spec.StorageClassName != nil {
			storageClass = *pvc.Spec.StorageClassName
		}

		// Get PVC Capacity (if set)
		capacity := "Unknown"
		if pvc.Status.Capacity != nil {
			if val, ok := pvc.Status.Capacity[v1.ResourceStorage]; ok {
				capacity = val.String()
			}
		}

		// Create a record for the Persistent Volume Claim
		pvcData = append(pvcData, PersistentVolumeClaimInfo{
			Name:         pvc.Name,
			Namespace:    pvc.Namespace,
			Status:       string(pvc.Status.Phase),
			Volume:       pvc.Spec.VolumeName,
			Capacity:     capacity,
			AccessModes:  accessModes,
			StorageClass: storageClass,
			Age:          age,
			VolumeMode:   volumeMode,
			Annotations:  annotations,
			Selector:     selector,
		})
	}

	sort.Slice(pvcData, func(i, j int) bool {
		return pvcData[i].Name < pvcData[j].Name
	})

	// Write CSV headers
	if err := writer.Write([]string{
		"PVC NAME",
		"NAMESPACE",
		"STATUS",
		"VOLUME",
		"CAPACITY",
		"ACCESS MODES",
		"STORAGE CLASS",
		"AGE",
		"VOLUME MODE",
		"ANNOTATIONS",
		"SELECTOR",
	}); err != nil {
		return fmt.Errorf("error writing headers to CSV: %v", err)
	}

	// Write Persistent Volume Claim data rows
	for _, pvc := range pvcData {
		record := []string{
			pvc.Name,
			pvc.Namespace,
			pvc.Status,
			pvc.Volume,
			pvc.Capacity,
			pvc.AccessModes,
			pvc.StorageClass,
			pvc.Age,
			pvc.VolumeMode,
			pvc.Annotations,
			pvc.Selector,
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
