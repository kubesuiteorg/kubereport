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

// StorageClassInfo holds the information for a Kubernetes StorageClass.
type StorageClassInfo struct {
	Name                 string
	Provisioner          string
	ReclaimPolicy        string
	BindingMode          string
	AllowVolumeExpansion string
	Default              bool
	Parameters           string
	Age                  string
	Annotations          string
}

// Generates a CSV report of Kubernetes StorageClasses.
func GenerateStorageClassReportCSV(writer *csv.Writer, clientset *kubernetes.Clientset) error {
	// Fetch StorageClasses
	storageClasses, err := clientset.StorageV1().StorageClasses().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("error fetching StorageClasses: %v", err)
	}

	var storageClassData []StorageClassInfo

	for _, sc := range storageClasses.Items {
		age := time.Since(sc.CreationTimestamp.Time).Round(time.Hour).String()

		reclaimPolicy := ""
		if sc.ReclaimPolicy != nil {
			reclaimPolicy = string(*sc.ReclaimPolicy)
		}

		bindingMode := ""
		if sc.VolumeBindingMode != nil {
			bindingMode = string(*sc.VolumeBindingMode)
		}

		allowVolumeExpansion := "No"
		if sc.AllowVolumeExpansion != nil && *sc.AllowVolumeExpansion {
			allowVolumeExpansion = "Yes"
		}

		isDefault := false
		if sc.Annotations["storageclass.kubernetes.io/is-default-class"] == "true" {
			isDefault = true
		}

		var params string
		for key, value := range sc.Parameters {
			params += fmt.Sprintf("%s=%s, ", key, value)
		}
		if len(params) > 0 {
			params = params[:len(params)-2]
		}

		var annotations string
		for key, value := range sc.Annotations {
			annotations += fmt.Sprintf("%s=%s, ", key, value)
		}
		if len(annotations) > 0 {
			annotations = annotations[:len(annotations)-2]
		}

		storageClassInfo := StorageClassInfo{
			Name:                 sc.Name,
			Provisioner:          sc.Provisioner,
			ReclaimPolicy:        reclaimPolicy,
			BindingMode:          bindingMode,
			AllowVolumeExpansion: allowVolumeExpansion,
			Default:              isDefault,
			Parameters:           params,
			Age:                  age,
			Annotations:          annotations,
		}

		storageClassData = append(storageClassData, storageClassInfo)
	}

	sort.Slice(storageClassData, func(i, j int) bool {
		return storageClassData[i].Name < storageClassData[j].Name
	})

	// Write CSV headers
	if err := writer.Write([]string{
		"STORAGECLASS NAME",
		"PROVISIONER",
		"RECLAIM POLICY",
		"BINDING MODE",
		"ALLOW VOLUME EXPANSION",
		"DEFAULT",
		"PARAMETERS",
		"AGE",
		"ANNOTATIONS",
	}); err != nil {
		return fmt.Errorf("error writing headers to CSV: %v", err)
	}

	// Write StorageClass data rows
	for _, sc := range storageClassData {
		record := []string{
			sc.Name,
			sc.Provisioner,
			sc.ReclaimPolicy,
			sc.BindingMode,
			sc.AllowVolumeExpansion,
			fmt.Sprintf("%t", sc.Default),
			sc.Parameters,
			sc.Age,
			sc.Annotations,
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
