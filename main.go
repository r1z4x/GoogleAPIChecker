package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var (
	apiToken  string
	projectID string
	threads   int
	output    string
	export    string
	exportDir string
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "googleapichecker",
		Short: "Google API Checker - Check all Google API products status and costs",
		Long: `Google API Checker is a CLI tool that checks the status of all Google API products
using multithreading and calculates potential costs based on pricing tables.`,
		Run: runChecker,
	}

	rootCmd.Flags().StringVarP(&apiToken, "token", "t", "", "Google API token (required)")
	rootCmd.Flags().StringVarP(&projectID, "project", "p", "", "Google Cloud Project ID (required for real API calls)")
	rootCmd.Flags().IntVarP(&threads, "threads", "n", 10, "Number of concurrent threads")
	rootCmd.Flags().StringVarP(&output, "output", "o", "results.json", "Output file path")
	rootCmd.Flags().StringVarP(&export, "export", "e", "", "Export format: csv, pdf, both")
	rootCmd.Flags().StringVarP(&exportDir, "export-dir", "d", ".", "Export directory")
	rootCmd.MarkFlagRequired("token")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func runChecker(cmd *cobra.Command, args []string) {
	fmt.Println("🚀 Starting Google API Checker...")
	fmt.Printf("📊 Using %d concurrent threads\n", threads)
	fmt.Printf("💾 Results will be saved to: %s\n", output)
	if export != "" {
		fmt.Printf("📤 Export format: %s\n", export)
		fmt.Printf("📁 Export directory: %s\n", exportDir)
	}
	fmt.Println()

	checker := NewGoogleAPIChecker(apiToken, projectID, threads)
	results, err := checker.CheckAllAPIs()
	if err != nil {
		log.Fatalf("Error checking APIs: %v", err)
	}

	// Save results
	if err := checker.SaveResults(results, output); err != nil {
		log.Fatalf("Error saving results: %v", err)
	}

	// Generate and print report
	report := GenerateReport(results)
	PrintReport(report)

	// Save report
	reportFile := strings.Replace(output, ".json", "_report.json", 1)
	if err := SaveReport(report, reportFile); err != nil {
		log.Fatalf("Error saving report: %v", err)
	}

	// Generate HTML report
	htmlFile := strings.Replace(output, ".json", "_report.html", 1)
	if err := generateHTMLReport(results, htmlFile); err != nil {
		log.Printf("Warning: HTML report generation failed: %v", err)
	}

	// Export if requested
	if export != "" {
		fmt.Println("📤 Exporting results...")
		exportOptions := ExportOptions{
			Format:    export,
			OutputDir: exportDir,
		}

		if err := ExportResults(report, results, exportOptions); err != nil {
			log.Printf("Warning: Export failed: %v", err)
		}

		// Also export summary
		if err := ExportSummary(report, exportOptions); err != nil {
			log.Printf("Warning: Summary export failed: %v", err)
		}
	}

	fmt.Println("✅ API checking completed successfully!")
	fmt.Printf("📄 Results saved to: %s\n", output)
	fmt.Printf("📊 Report saved to: %s\n", reportFile)
}
