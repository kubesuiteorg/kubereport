package report

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	detailed "github.com/kubesuiteorg/kubereport/pkg/report/detailed-report"
	general "github.com/kubesuiteorg/kubereport/pkg/report/general-report"

	"github.com/jung-kurt/gofpdf/v2"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
)

var logger *log.Logger

func init() {
	if isRunningInKubernetes() {
		logger = log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile)
	}
}

// Check if the application is running in a Kubernetes pod
func isRunningInKubernetes() bool {
	// Read the environment variable that Kubernetes sets
	_, err := os.Stat("/var/run/secrets/kubernetes.io/serviceaccount")
	return !os.IsNotExist(err)
}

func getClientConfig(kubeconfigPath string) (*rest.Config, string, error) {
	if kubeconfigPath == "" {
		kubeconfigPath = clientcmd.RecommendedHomeFile
	}

	kubeconfig, err := clientcmd.LoadFromFile(kubeconfigPath)
	if err == nil {
		currentContext, ok := kubeconfig.Contexts[kubeconfig.CurrentContext]
		if !ok {
			return nil, "", fmt.Errorf("could not find current context in kubeconfig")
		}

		clusterName := currentContext.Cluster
		if clusterName == "" {
			return nil, "", fmt.Errorf("cluster name not found in current context")
		}

		config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
		if err != nil {
			return nil, "", fmt.Errorf("failed to create Kubernetes client config: %v", err)
		}

		return config, clusterName, nil
	}

	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, "", fmt.Errorf("failed to load in-cluster configuration: %v", err)
	}

	clusterName := config.Host
	return config, clusterName, nil
}

type reportSection struct {
	Title        string
	PDFGenerator func(pdf *gofpdf.Fpdf, clientset *kubernetes.Clientset) error
	CSVGenerator func(writer *csv.Writer, clientset *kubernetes.Clientset) error
}

// GeneratePDF creates a PDF report and saves it to a dynamically named file based on the cluster name and timestamp.
func GeneratePDF(kubeconfigPath string) (string, string, error) {
	if logger != nil {
		logger.Println("Starting PDF report generation...")
	}

	config, clusterName, err := getClientConfig(kubeconfigPath)
	if err != nil {
		if logger != nil {
			logger.Printf("Error getting client config: %v\n", err)
		}
		return "", "", err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		if logger != nil {
			logger.Printf("Failed to create Kubernetes clientset: %v\n", err)
		}
		return "", "", fmt.Errorf("failed to create Kubernetes clientset: %v", err)
	}

	metricsClientset, err := metricsv.NewForConfig(config)
	if err != nil {
		if logger != nil {
			logger.Printf("Failed to create metrics clientset: %v\n", err)
		}
		return "", "", fmt.Errorf("failed to create metrics clientset: %v", err)
	}

	sections := []reportSection{
		{"Cluster Resource Details", func(pdf *gofpdf.Fpdf, cs *kubernetes.Clientset) error {
			return general.GenerateClusterSummaryTable(pdf, cs, metricsClientset)
		}, nil},
		{"Node Resource Details", general.GenerateNodeSummaryTable, nil},
		{"Namespace Resource Details", general.GenerateNamespaceTable, nil},
		{"Namespace Summary ", general.GenerateNamespaceSummaryTable, nil},
		{"Pod Distribution Details", general.GeneratePodDistributionReport, nil},
		{"Pod Resource Details", general.GeneratePodResourceUsageTable, nil},
		{"Pod Status", general.GeneratePodDetailsTable, nil},
	}

	currentTime := time.Now()
	formattedTime := currentTime.Format("02-01-2006-15-04")
	outputPath := fmt.Sprintf("kubernetes_cluster_report_%s.pdf", formattedTime)

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Set font and add the title text
	pdf.SetFont("Arial", "B", 18)
	pdf.Ln(10)
	pdf.Cell(40, 10, "KUBEREPORT")
	pdf.Ln(15)

	pdf.SetFont("Arial", "B", 18)
	pdf.Cell(40, 10, "Kubernetes Cluster Qualification Report")
	pdf.Ln(10)
	pdf.SetFont("Arial", "", 12)

	for _, section := range sections {
		pdf.Ln(5)
		pdf.SetFont("Arial", "B", 15)
		pdf.Cell(0, 10, section.Title)
		pdf.Ln(10)

		if section.PDFGenerator != nil {
			if err := section.PDFGenerator(pdf, clientset); err != nil {
				if logger != nil {
					logger.Printf("Failed to generate %s: %v\n", section.Title, err)
				}
				return "", "", fmt.Errorf("failed to generate %s: %v", section.Title, err)
			}
		}

		separator := strings.Repeat("-", 160)
		pdf.Ln(10)
		pdf.Cell(0, 0, separator)
		pdf.Ln(4)

	}

	if err := pdf.OutputFileAndClose(outputPath); err != nil {
		if logger != nil {
			logger.Printf("Failed to save PDF file: %v\n", err)
		}
		return "", "", fmt.Errorf("failed to save PDF file: %v", err)
	}

	if logger != nil {
		logger.Println("PDF report generated successfully.")
	}
	return clusterName, outputPath, nil
}

// Generate a CSV report and saves it to a dynamically named file based on the cluster name and timestamp.
func GenerateCSV(kubeconfigPath string) (string, string, error) {
	if logger != nil {
		logger.Println("Starting CSV report generation...")
	}

	config, clusterName, err := getClientConfig(kubeconfigPath)
	if err != nil {
		if logger != nil {
			logger.Printf("Error getting client config: %v\n", err)
		}
		return "", "", err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		if logger != nil {
			logger.Printf("Failed to create Kubernetes clientset: %v\n", err)
		}
		return "", "", fmt.Errorf("failed to create Kubernetes clientset: %v", err)
	}

	metricsClientset, err := metricsv.NewForConfig(config)
	if err != nil {
		if logger != nil {
			logger.Printf("Failed to create metrics clientset: %v\n", err)
		}
		return "", "", fmt.Errorf("failed to create metrics clientset: %v", err)
	}

	sections := []reportSection{
		{"[ CLUSTER RESOURCE DETAILS ]", nil, func(writer *csv.Writer, cs *kubernetes.Clientset) error {
			return detailed.GenerateClusterSummaryCSV(writer, cs, metricsClientset)
		}},
		{"[ NODE RESOURCE DETAILS ]", nil, detailed.GenerateNodeSummaryTable},
		{"[ NAMESPACE DETAILS ]", nil, detailed.GenerateNamespaceTable},
		{"[ POD DETAILS ]", nil, detailed.GeneratePodResourceUsageCSV},
		{"[ DEPLOYMENT DETAILS ]", nil, detailed.GenerateDeploymentReportCSV},
		{"[ SERVICE DETAILS ]", nil, detailed.GenerateServiceReportCSV},
		{"[ ENDPOINTS DETAILS ]", nil, detailed.GenerateEndpointsReportCSV},
		{"[ REPLICASET DETAILS ]", nil, detailed.GenerateReplicaSetReportCSV},
		{"[ STATEFULSET DETAILS ]", nil, detailed.GenerateStatefulSetReportCSV},
		{"[ DAEMONSETS DETAILS ]", nil, detailed.GenerateDaemonSetReportCSV},
		{"[ CONFIGMAP DETAILS ]", nil, detailed.GenerateConfigMapReportCSV},
		{"[ SECRET DETAILS ]", nil, detailed.GenerateSecretReportCSV},
		{"[ SERVICEACCOUNT DETAILS ]", nil, detailed.GenerateServiceAccountReportCSV},
		{"[ PERSISTENT VOLUMES DETAILS ]", nil, detailed.GeneratePersistentVolumeReportCSV},
		{"[ PERSISTENT VOLUME CLAIM DETAILS ]", nil, detailed.GeneratePersistentVolumeClaimReportCSV},
		{"[ STORAGE CLASS DETAILS ]", nil, detailed.GenerateStorageClassReportCSV},
		{"[ INGRESS RESOURCES DETAILS ]", nil, detailed.GenerateIngressReportCSV},
		{"[ NETWORK POLICY DETAILS ]", nil, detailed.GenerateNetworkPolicyReportCSV},
		{"[ RESOURCE QUOTA DETAILS ]", nil, detailed.GenerateResourceQuotaReportCSV},
		{"[ LIMIT RANGE DETAILS ]", nil, detailed.GenerateLimitRangeReportCSV},
		{"[ HORIZONTAL POD AUTOSCALERS DETAILS ]", nil, detailed.GenerateHPAReportCSV},
		{"[ JOB DETAILS ]", nil, detailed.GenerateJobReportCSV},
		{"[ CRONJOB DETAILS ]", nil, detailed.GenerateCronJobReportCSV},
		{"[ ROLE DETAILS ]", nil, detailed.GenerateRoleReportCSV},
		{"[ ROLEBINDING DETAILS ]", nil, detailed.GenerateRoleBindingReportCSV},
		{"[ CLUSTERROLE DETAILS ]", nil, detailed.GenerateClusterRoleReportCSV},
		{"[ CLUSTERROLEBINDING DETAILS ]", nil, detailed.GenerateClusterRoleBindingReportCSV},
	}

	currentTime := time.Now()
	formattedTime := currentTime.Format("02-01-2006-15-04")
	outputPath := fmt.Sprintf("kubernetes_cluster_report_%s.csv", formattedTime)

	file, err := os.Create(outputPath)
	if err != nil {
		if logger != nil {
			logger.Printf("Failed to create CSV file: %v\n", err)
		}
		return "", "", fmt.Errorf("failed to create CSV file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Add image reference at the top of the CSV
	if err := writer.Write([]string{"KUBE☸️REPORT"}); err != nil {
		if logger != nil {
			logger.Printf("Failed to write KUBEREPORT to CSV: %v\n", err)
		}
		return "", "", fmt.Errorf("failed to write KUBEREPORT to CSV: %v", err)
	}

	for _, section := range sections {
		if err := writer.Write([]string{section.Title}); err != nil {
			if logger != nil {
				logger.Printf("Failed to write %s title row to CSV: %v\n", section.Title, err)
			}
			return "", "", fmt.Errorf("failed to write %s title row to CSV: %v", section.Title, err)
		}

		if section.CSVGenerator != nil {
			if err := section.CSVGenerator(writer, clientset); err != nil {
				if logger != nil {
					logger.Printf("Failed to generate %s: %v\n", section.Title, err)
				}
				return "", "", fmt.Errorf("failed to generate %s: %v", section.Title, err)
			}
		}

		if err := writer.Write([]string{}); err != nil {
			if logger != nil {
				logger.Printf("Failed to write empty row to CSV: %v\n", err)
			}
			return "", "", fmt.Errorf("failed to write empty row to CSV: %v", err)
		}
	}

	if logger != nil {
		logger.Println("CSV report generated successfully.")
	}
	return clusterName, outputPath, nil
}
