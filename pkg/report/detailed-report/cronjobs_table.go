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

// CronJobInfo holds the information for a Kubernetes CronJob.
type CronJobInfo struct {
	Name              string
	Namespace         string
	Schedule          string
	ActiveJobs        int32
	LastSchedule      string
	Age               string
	JobDuration       string
	JobTemplate       string
	HistoryLimit      int32
	ConcurrencyPolicy string
}

// Generates a CSV report of Kubernetes CronJobs.
func GenerateCronJobReportCSV(writer *csv.Writer, clientset *kubernetes.Clientset) error {
	// Fetch CronJobs
	cronJobs, err := clientset.BatchV1().CronJobs("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("error fetching CronJobs: %v", err)
	}

	var cronJobData []CronJobInfo

	for _, cronJob := range cronJobs.Items {
		age := time.Since(cronJob.CreationTimestamp.Time).Round(time.Hour).String()

		var lastSchedule string
		if cronJob.Status.LastScheduleTime != nil {
			lastSchedule = time.Since(cronJob.Status.LastScheduleTime.Time).Round(time.Hour).String()
		} else {
			lastSchedule = "N/A"
		}

		var jobDuration string
		if cronJob.Status.LastSuccessfulTime != nil {
			duration := time.Since(cronJob.Status.LastSuccessfulTime.Time).Round(time.Hour)
			jobDuration = duration.String()
		} else {
			jobDuration = "N/A"
		}

		jobTemplate := fmt.Sprintf("%s/%s", cronJob.Spec.JobTemplate.Name, cronJob.Spec.JobTemplate.Namespace)
		historyLimit := int32(0)
		if cronJob.Spec.SuccessfulJobsHistoryLimit != nil {
			historyLimit = *cronJob.Spec.SuccessfulJobsHistoryLimit
		}

		concurrencyPolicy := string(cronJob.Spec.ConcurrencyPolicy)

		cronJobInfo := CronJobInfo{
			Name:              cronJob.Name,
			Namespace:         cronJob.Namespace,
			Schedule:          cronJob.Spec.Schedule,
			ActiveJobs:        int32(len(cronJob.Status.Active)),
			LastSchedule:      lastSchedule,
			Age:               age,
			JobDuration:       jobDuration,
			JobTemplate:       jobTemplate,
			HistoryLimit:      historyLimit,
			ConcurrencyPolicy: concurrencyPolicy,
		}

		cronJobData = append(cronJobData, cronJobInfo)
	}

	sort.Slice(cronJobData, func(i, j int) bool {
		return cronJobData[i].Name < cronJobData[j].Name
	})

	if err := writer.Write([]string{
		"CRONJOB NAME",
		"NAMESPACE",
		"SCHEDULE",
		"ACTIVE JOBS",
		"LAST SCHEDULE",
		"AGE",
		"JOB DURATION",
		"JOB TEMPLATE",
		"HISTORY LIMIT",
		"CONCURRENCY POLICY",
	}); err != nil {
		return fmt.Errorf("error writing headers to CSV: %v", err)
	}
	for _, cronJob := range cronJobData {
		record := []string{
			cronJob.Name,
			cronJob.Namespace,
			cronJob.Schedule,
			fmt.Sprintf("%d", cronJob.ActiveJobs),
			cronJob.LastSchedule,
			cronJob.Age,
			cronJob.JobDuration,
			cronJob.JobTemplate,
			fmt.Sprintf("%d", cronJob.HistoryLimit),
			cronJob.ConcurrencyPolicy,
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
