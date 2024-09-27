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

// PersistentVolumeInfo holds the information for a Kubernetes Persistent Volume.
type PersistentVolumeInfo struct {
	Name                  string
	Capacity              string
	AccessModes           string
	ReclaimPolicy         string
	Status                string
	PersistentVolumeClaim string
	StorageClass          string
	Age                   string
	Phase                 string
	Annotations           string
	Claimant              string // The application or pod using the PV
	VolumeMode            string // Block or Filesystem
	MountOptions          string // Specific mount options used for the PV
}

// getPodUsingPVC finds the pod(s) using a specific Persistent Volume Claim (PVC) in a namespace.
func getPodUsingPVC(clientset *kubernetes.Clientset, namespace, pvcName string) (string, error) {
	// List all pods in the namespace
	pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return "", fmt.Errorf("error fetching pods in namespace %s: %v", namespace, err)
	}

	// Iterate over all pods to find the one using the PVC
	for _, pod := range pods.Items {
		for _, volume := range pod.Spec.Volumes {
			if volume.PersistentVolumeClaim != nil && volume.PersistentVolumeClaim.ClaimName == pvcName {
				return pod.Name, nil
			}
		}
	}

	return "Unknown", nil
}

// Generates a CSV report of Kubernetes Persistent Volumes.
func GeneratePersistentVolumeReportCSV(writer *csv.Writer, clientset *kubernetes.Clientset) error {
	// Fetch Persistent Volumes
	persistentVolumes, err := clientset.CoreV1().PersistentVolumes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("error fetching Persistent Volumes: %v", err)
	}

	var pvData []PersistentVolumeInfo

	// Iterate over Persistent Volumes to get their information
	for _, pv := range persistentVolumes.Items {
		age := time.Since(pv.CreationTimestamp.Time).Round(time.Hour).String()

		accessModes := fmt.Sprintf("%v", pv.Spec.AccessModes)

		reclaimPolicy := string(pv.Spec.PersistentVolumeReclaimPolicy)

		pvClaim := ""
		claimant := "Unknown"
		if pv.Spec.ClaimRef != nil {
			pvClaim = pv.Spec.ClaimRef.Name

			// Find the pod or application using the PVC
			claimant, err = getPodUsingPVC(clientset, pv.Spec.ClaimRef.Namespace, pvClaim)
			if err != nil {
				fmt.Printf("error finding pod using PVC %s in namespace %s: %v\n", pvClaim, pv.Spec.ClaimRef.Namespace, err)
			}
		}

		volumeMode := ""
		if pv.Spec.VolumeMode != nil {
			volumeMode = string(*pv.Spec.VolumeMode)
		}

		mountOptions := fmt.Sprintf("%v", pv.Spec.MountOptions)

		// Access Capacity as Decimal and get its string representation
		capacity := pv.Spec.Capacity[v1.ResourceStorage]
		capacityStr := capacity.String()

		// Create a record for the Persistent Volume
		pvData = append(pvData, PersistentVolumeInfo{
			Name:                  pv.Name,
			Capacity:              capacityStr,
			AccessModes:           accessModes,
			ReclaimPolicy:         reclaimPolicy,
			Status:                string(pv.Status.Phase),
			PersistentVolumeClaim: pvClaim,
			StorageClass:          pv.Spec.StorageClassName,
			Age:                   age,
			Phase:                 string(pv.Status.Phase),
			Annotations:           fmt.Sprintf("%v", pv.Annotations),
			Claimant:              claimant,     // Set the pod/application name using the PVC
			VolumeMode:            volumeMode,   // Block or Filesystem
			MountOptions:          mountOptions, // Specific mount options
		})
	}

	sort.Slice(pvData, func(i, j int) bool {
		return pvData[i].Name < pvData[j].Name
	})

	// Write CSV headers
	if err := writer.Write([]string{
		"PV NAME",
		"CAPACITY",
		"ACCESS MODES",
		"RECLAIM POLICY",
		"STATUS",
		"PERSISTENT VOLUME CLAIM",
		"STORAGE CLASS",
		"AGE",
		"PHASE",
		"ANNOTATIONS",
		"CLAIMANT",
		"VOLUME MODE",
		"MOUNT OPTIONS",
	}); err != nil {
		return fmt.Errorf("error writing headers to CSV: %v", err)
	}

	// Write Persistent Volume data rows
	for _, pv := range pvData {
		record := []string{
			pv.Name,
			pv.Capacity,
			pv.AccessModes,
			pv.ReclaimPolicy,
			pv.Status,
			pv.PersistentVolumeClaim,
			pv.StorageClass,
			pv.Age,
			pv.Phase,
			pv.Annotations,
			pv.Claimant,
			pv.VolumeMode,
			pv.MountOptions,
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
