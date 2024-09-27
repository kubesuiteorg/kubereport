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

// JobInfo holds the information for a Kubernetes Job.
type JobInfo struct {
	Name          string
	Namespace     string
	Completions   int32
	Parallelism   int32
	ActivePods    int32
	SucceededPods int32
	FailedPods    int32
	Age           string
	Conditions    string
	JobDuration   string
	JobTemplate   string
}

// Generates a CSV report of Kubernetes Jobs.
func GenerateJobReportCSV(writer *csv.Writer, clientset *kubernetes.Clientset) error {
	// Fetch Jobs
	jobs, err := clientset.BatchV1().Jobs("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("error fetching Jobs: %v", err)
	}

	var jobData []JobInfo

	for _, job := range jobs.Items {
		age := time.Since(job.CreationTimestamp.Time).Round(time.Hour).String()

		var jobDuration string
		if job.Status.StartTime != nil && job.Status.CompletionTime != nil {
			duration := job.Status.CompletionTime.Time.Sub(job.Status.StartTime.Time)
			jobDuration = duration.String()
		} else {
			jobDuration = "N/A"
		}

		jobTemplate := fmt.Sprintf("%s/%s", job.Spec.Template.Name, job.Spec.Template.Namespace)

		var conditions string
		for _, condition := range job.Status.Conditions {
			conditions += fmt.Sprintf("%s (%s), ", string(condition.Type), condition.Status)
		}
		if len(conditions) > 0 {
			conditions = conditions[:len(conditions)-2]
		}

		jobInfo := JobInfo{
			Name:          job.Name,
			Namespace:     job.Namespace,
			Completions:   *job.Spec.Completions,
			Parallelism:   *job.Spec.Parallelism,
			ActivePods:    job.Status.Active,
			SucceededPods: job.Status.Succeeded,
			FailedPods:    job.Status.Failed,
			Age:           age,
			Conditions:    conditions,
			JobDuration:   jobDuration,
			JobTemplate:   jobTemplate,
		}

		jobData = append(jobData, jobInfo)
	}

	sort.Slice(jobData, func(i, j int) bool {
		return jobData[i].Name < jobData[j].Name
	})

	if err := writer.Write([]string{
		"JOB NAME",
		"NAMESPACE",
		"COMPLETIONS",
		"PARALLELISM",
		"ACTIVE PODS",
		"SUCCEEDED PODS",
		"FAILED PODS",
		"AGE",
		"CONDITIONS",
		"JOB DURATION",
		"JOB TEMPLATE",
	}); err != nil {
		return fmt.Errorf("error writing headers to CSV: %v", err)
	}
	for _, job := range jobData {
		record := []string{
			job.Name,
			job.Namespace,
			fmt.Sprintf("%d", job.Completions),
			fmt.Sprintf("%d", job.Parallelism),
			fmt.Sprintf("%d", job.ActivePods),
			fmt.Sprintf("%d", job.SucceededPods),
			fmt.Sprintf("%d", job.FailedPods),
			job.Age,
			job.Conditions,
			job.JobDuration,
			job.JobTemplate,
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("error writing record to CSV: %v", err)
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return fmt.Errorf("error flushing CSV writer: %v", err)
	}

	return nil
}
