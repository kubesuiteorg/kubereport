package cmd

import (
	"fmt"
	"log"
	"time"

	"github.com/kubesuiteorg/kubereport/pkg/email"
	"github.com/kubesuiteorg/kubereport/pkg/report"
	"github.com/robfig/cron/v3"
	"github.com/spf13/cobra"
)

var (
	recipient   string
	sender      string
	password    string
	subject     string
	body        string
	kubeconfig  string
	schedule    string
	reportType  string
	smtpServer  string
	smtpPort    string
	useTLS      bool
	showVersion bool
)

var version = "v0.1.1"

var rootCmd = &cobra.Command{
	Use:   "kubereport",
	Short: "Generate Kubernetes cluster reports",
	Run: func(cmd *cobra.Command, args []string) {
		if showVersion {
			fmt.Println("Kubereport Version:", version)
			return
		}

		if schedule != "" {
			// Schedule the report generation
			c := cron.New()
			_, err := c.AddFunc(schedule, func() {
				runReportGeneration()
			})
			if err != nil {
				log.Fatalf("Error scheduling report: %v", err)
			}
			c.Start()
			// Keep the application running
			select {}
		} else {
			// Run report generation immediately
			runReportGeneration()
		}
	},
}

func runReportGeneration() {
	var (
		clusterName string
		outputPath  string
		err         error
	)

	switch reportType {
	case "detailed":
		// Generate the CSV report
		clusterName, outputPath, err = report.GenerateCSV(kubeconfig)
	default:
		// Generate the PDF report
		clusterName, outputPath, err = report.GeneratePDF(kubeconfig)
	}

	if err != nil {
		log.Fatalf("Error generating report for the %s cluster: %v", clusterName, err)
	}

	reportDate := time.Now().Format("02-01-2006")
	emailSubject := fmt.Sprintf("%s - %s", subject, reportDate)
	emailBody := body
	// Check if email parameters are provided
	if recipient != "" && sender != "" && password != "" {
		// Send the email with the attached report
		if err := email.SendEmail(emailSubject, emailBody, recipient, sender, smtpServer, smtpPort, sender, password, outputPath, useTLS); err != nil {
			log.Fatalf("Error sending email: %v", err)
		}
	}
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.Flags().BoolVarP(&showVersion, "version", "v", false, "Show KubeReport version.")
	rootCmd.Flags().StringVarP(&recipient, "recipient", "r", "", "Email address of the report recipient.")
	rootCmd.Flags().StringVarP(&sender, "sender", "s", "", "Email address of the report sender.")
	rootCmd.Flags().StringVarP(&password, "password", "p", "", "SMTP password for authentication.")
	rootCmd.Flags().StringVarP(&subject, "subject", "j", "", "Subject line for the email report.")
	rootCmd.Flags().StringVarP(&body, "body", "b", "", "Body content of the email report.")
	rootCmd.Flags().StringVarP(&kubeconfig, "kubeconfig", "k", "", "Path to kubeconfig file.")
	rootCmd.Flags().StringVarP(&schedule, "schedule", "t", "", "Cron schedule for report generation (e.g., '* * * * *').")
	rootCmd.Flags().StringVarP(&reportType, "report", "d", "general", "Report type: 'general' (PDF) or 'detailed' (CSV).")
	rootCmd.Flags().StringVarP(&smtpServer, "smtp-server", "m", "", "SMTP server address (e.g., smtp.gmail.com).")
	rootCmd.Flags().StringVarP(&smtpPort, "smtp-port", "o", "", "SMTP server port (default: 587).")
	rootCmd.Flags().BoolVarP(&useTLS, "use-tls", "u", true, "Enable TLS for SMTP connection (default: true).")
}
