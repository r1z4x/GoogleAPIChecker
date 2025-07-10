package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/jung-kurt/gofpdf"
)

// ExportOptions contains export configuration
type ExportOptions struct {
	Format     string // "csv", "pdf", "both"
	OutputDir  string
	IncludeRaw bool
}

// ExportResults exports the results in various formats
func ExportResults(report *Report, results []APIResult, options ExportOptions) error {
	switch options.Format {
	case "csv":
		return exportToCSV(report, results, options)
	case "pdf":
		return exportToPDF(report, results, options)
	case "both":
		if err := exportToCSV(report, results, options); err != nil {
			return fmt.Errorf("CSV export failed: %v", err)
		}
		if err := exportToPDF(report, results, options); err != nil {
			return fmt.Errorf("PDF export failed: %v", err)
		}
		return nil
	default:
		return fmt.Errorf("unsupported export format: %s", options.Format)
	}
}

// exportToCSV exports results to CSV format
func exportToCSV(report *Report, results []APIResult, options ExportOptions) error {
	filename := filepath.Join(options.OutputDir, fmt.Sprintf("google_api_checker_%s.csv", time.Now().Format("20060102_150405")))

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create CSV file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{
		"API Name",
		"Display Name",
		"Status",
		"Enabled",
		"Has Pricing",
		"Unlimited Cost",
		"Estimated Cost (USD)",
		"Currency",
		"Pricing Details",
		"Checked At",
		"Error",
	}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %v", err)
	}

	// Write data rows
	for _, result := range results {
		row := []string{
			result.Name,
			result.DisplayName,
			result.Status,
			strconv.FormatBool(result.Enabled),
			strconv.FormatBool(result.CostInfo.HasPricing),
			strconv.FormatBool(result.CostInfo.UnlimitedCost),
			fmt.Sprintf("%.2f", result.CostInfo.EstimatedCost),
			result.CostInfo.Currency,
			result.CostInfo.PricingDetails,
			result.CheckedAt.Format("2006-01-02 15:04:05"),
			result.Error,
		}
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row: %v", err)
		}
	}

	fmt.Printf("✅ CSV exported to: %s\n", filename)
	return nil
}

// exportToPDF exports results to PDF format
func exportToPDF(report *Report, results []APIResult, options ExportOptions) error {
	filename := filepath.Join(options.OutputDir, fmt.Sprintf("google_api_checker_%s.pdf", time.Now().Format("20060102_150405")))

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Title
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(190, 10, "Google API Checker Report")
	pdf.Ln(15)

	// Summary section
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(190, 8, "Summary")
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 10)
	pdf.Cell(95, 6, fmt.Sprintf("Total APIs checked: %d", report.Summary.TotalAPIs))
	pdf.Cell(95, 6, fmt.Sprintf("Enabled APIs: %d", report.Summary.EnabledCount))
	pdf.Ln(6)
	pdf.Cell(95, 6, fmt.Sprintf("Disabled APIs: %d", report.Summary.DisabledCount))
	pdf.Cell(95, 6, fmt.Sprintf("Errors: %d", report.Summary.ErrorCount))
	pdf.Ln(6)
	pdf.Cell(95, 6, fmt.Sprintf("Total estimated cost: $%.2f %s", report.Summary.TotalCost, report.Summary.Currency))
	pdf.Ln(15)

	// Unlimited cost APIs section
	if len(report.CostAnalysis.UnlimitedCostAPIs) > 0 {
		pdf.SetFont("Arial", "B", 12)
		pdf.Cell(190, 8, fmt.Sprintf("⚠️ Unlimited Cost APIs (%d)", len(report.CostAnalysis.UnlimitedCostAPIs)))
		pdf.Ln(10)

		pdf.SetFont("Arial", "", 10)
		for _, api := range report.CostAnalysis.UnlimitedCostAPIs {
			pdf.Cell(190, 6, fmt.Sprintf("• %s", api.DisplayName))
			pdf.Ln(6)
			pdf.Cell(190, 6, fmt.Sprintf("  %s", api.CostInfo.PricingDetails))
			pdf.Ln(8)
		}
		pdf.Ln(10)
	}

	// High cost APIs section
	if len(report.CostAnalysis.HighCostAPIs) > 0 {
		pdf.SetFont("Arial", "B", 12)
		pdf.Cell(190, 8, fmt.Sprintf("💰 High Cost APIs (%d)", len(report.CostAnalysis.HighCostAPIs)))
		pdf.Ln(10)

		pdf.SetFont("Arial", "", 10)
		for _, api := range report.CostAnalysis.HighCostAPIs {
			pdf.Cell(190, 6, fmt.Sprintf("• %s: $%.2f/month", api.DisplayName, api.CostInfo.EstimatedCost))
			pdf.Ln(6)
		}
		pdf.Ln(10)
	}

	// Recommendations section
	if len(report.Recommendations) > 0 {
		pdf.SetFont("Arial", "B", 12)
		pdf.Cell(190, 8, "💡 Recommendations")
		pdf.Ln(10)

		pdf.SetFont("Arial", "", 10)
		for _, rec := range report.Recommendations {
			pdf.Cell(190, 6, fmt.Sprintf("• %s", rec))
			pdf.Ln(6)
		}
		pdf.Ln(10)
	}

	// Detailed results table
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(190, 8, "Detailed API Results")
	pdf.Ln(10)

	// Table header
	pdf.SetFont("Arial", "B", 8)
	headers := []string{"API Name", "Status", "Enabled", "Cost", "Unlimited"}
	widths := []float64{60, 25, 20, 25, 25}

	for i, header := range headers {
		pdf.CellFormat(widths[i], 6, header, "1", 0, "", false, 0, "")
	}
	pdf.Ln(-1)

	// Table data
	pdf.SetFont("Arial", "", 8)
	for _, result := range results {
		if pdf.GetY() > 250 { // Check if we need a new page
			pdf.AddPage()
			// Repeat header
			pdf.SetFont("Arial", "B", 8)
			for i, header := range headers {
				pdf.CellFormat(widths[i], 6, header, "1", 0, "", false, 0, "")
			}
			pdf.Ln(-1)
			pdf.SetFont("Arial", "", 8)
		}

		// Truncate long names
		apiName := result.DisplayName
		if len(apiName) > 25 {
			apiName = apiName[:22] + "..."
		}

		enabled := "No"
		if result.Enabled {
			enabled = "Yes"
		}

		unlimited := "No"
		if result.CostInfo.UnlimitedCost {
			unlimited = "Yes"
		}

		cost := fmt.Sprintf("$%.2f", result.CostInfo.EstimatedCost)

		row := []string{apiName, result.Status, enabled, cost, unlimited}
		for i, cell := range row {
			pdf.CellFormat(widths[i], 6, cell, "1", 0, "", false, 0, "")
		}
		pdf.Ln(-1)
	}

	// Footer
	pdf.SetY(-20)
	pdf.SetFont("Arial", "I", 8)
	pdf.Cell(190, 6, fmt.Sprintf("Report generated at: %s", report.GeneratedAt.Format("2006-01-02 15:04:05")))
	pdf.Ln(6)
	pdf.Cell(190, 6, "Generated by Google API Checker")

	if err := pdf.OutputFileAndClose(filename); err != nil {
		return fmt.Errorf("failed to save PDF: %v", err)
	}

	fmt.Printf("✅ PDF exported to: %s\n", filename)
	return nil
}

// ExportSummary exports a summary report
func ExportSummary(report *Report, options ExportOptions) error {
	filename := filepath.Join(options.OutputDir, fmt.Sprintf("summary_%s.txt", time.Now().Format("20060102_150405")))

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create summary file: %v", err)
	}
	defer file.Close()

	// Write summary
	fmt.Fprintf(file, "Google API Checker Summary Report\n")
	fmt.Fprintf(file, "Generated: %s\n\n", report.GeneratedAt.Format("2006-01-02 15:04:05"))

	fmt.Fprintf(file, "SUMMARY:\n")
	fmt.Fprintf(file, "  Total APIs: %d\n", report.Summary.TotalAPIs)
	fmt.Fprintf(file, "  Enabled: %d\n", report.Summary.EnabledCount)
	fmt.Fprintf(file, "  Disabled: %d\n", report.Summary.DisabledCount)
	fmt.Fprintf(file, "  Errors: %d\n", report.Summary.ErrorCount)
	fmt.Fprintf(file, "  Total Cost: $%.2f %s\n\n", report.Summary.TotalCost, report.Summary.Currency)

	if len(report.CostAnalysis.UnlimitedCostAPIs) > 0 {
		fmt.Fprintf(file, "UNLIMITED COST APIS (%d):\n", len(report.CostAnalysis.UnlimitedCostAPIs))
		for _, api := range report.CostAnalysis.UnlimitedCostAPIs {
			fmt.Fprintf(file, "  • %s\n", api.DisplayName)
		}
		fmt.Fprintf(file, "\n")
	}

	if len(report.CostAnalysis.HighCostAPIs) > 0 {
		fmt.Fprintf(file, "HIGH COST APIS (%d):\n", len(report.CostAnalysis.HighCostAPIs))
		for _, api := range report.CostAnalysis.HighCostAPIs {
			fmt.Fprintf(file, "  • %s: $%.2f/month\n", api.DisplayName, api.CostInfo.EstimatedCost)
		}
		fmt.Fprintf(file, "\n")
	}

	if len(report.Recommendations) > 0 {
		fmt.Fprintf(file, "RECOMMENDATIONS:\n")
		for _, rec := range report.Recommendations {
			fmt.Fprintf(file, "  • %s\n", rec)
		}
	}

	fmt.Printf("✅ Summary exported to: %s\n", filename)
	return nil
}
