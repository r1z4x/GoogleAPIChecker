package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"
)

// Report represents the analysis report
type Report struct {
	Summary         SummaryInfo  `json:"summary"`
	EnabledAPIs     []APIResult  `json:"enabled_apis"`
	DisabledAPIs    []APIResult  `json:"disabled_apis"`
	CostAnalysis    CostAnalysis `json:"cost_analysis"`
	Recommendations []string     `json:"recommendations"`
	GeneratedAt     time.Time    `json:"generated_at"`
}

// SummaryInfo contains summary statistics
type SummaryInfo struct {
	TotalAPIs     int     `json:"total_apis"`
	EnabledCount  int     `json:"enabled_count"`
	DisabledCount int     `json:"disabled_count"`
	ErrorCount    int     `json:"error_count"`
	TotalCost     float64 `json:"total_cost"`
	Currency      string  `json:"currency"`
}

// CostAnalysis contains detailed cost information
type CostAnalysis struct {
	TotalEstimatedCost float64            `json:"total_estimated_cost"`
	UnlimitedCostAPIs  []APIResult        `json:"unlimited_cost_apis"`
	HighCostAPIs       []APIResult        `json:"high_cost_apis"`
	CostBreakdown      map[string]float64 `json:"cost_breakdown"`
}

// GenerateReport creates a comprehensive analysis report
func GenerateReport(results []APIResult) *Report {
	report := &Report{
		GeneratedAt: time.Now(),
	}

	// Separate APIs by status
	var enabledAPIs, disabledAPIs []APIResult
	var errorCount int
	var totalCost float64
	var unlimitedCostAPIs, highCostAPIs []APIResult
	costBreakdown := make(map[string]float64)

	for _, result := range results {
		if result.Error != "" {
			errorCount++
			continue
		}

		if result.Enabled {
			enabledAPIs = append(enabledAPIs, result)

			// Calculate costs
			if result.CostInfo.HasPricing {
				totalCost += result.CostInfo.EstimatedCost
				costBreakdown[result.DisplayName] = result.CostInfo.EstimatedCost

				// Check for unlimited cost APIs
				if result.CostInfo.UnlimitedCost {
					unlimitedCostAPIs = append(unlimitedCostAPIs, result)
				}

				// Check for high cost APIs (>$50)
				if result.CostInfo.EstimatedCost > 50.0 {
					highCostAPIs = append(highCostAPIs, result)
				}
			}
		} else {
			disabledAPIs = append(disabledAPIs, result)
		}
	}

	// Sort APIs by cost (highest first)
	sort.Slice(highCostAPIs, func(i, j int) bool {
		return highCostAPIs[i].CostInfo.EstimatedCost > highCostAPIs[j].CostInfo.EstimatedCost
	})

	// Sort unlimited cost APIs by name
	sort.Slice(unlimitedCostAPIs, func(i, j int) bool {
		return unlimitedCostAPIs[i].DisplayName < unlimitedCostAPIs[j].DisplayName
	})

	// Create summary
	report.Summary = SummaryInfo{
		TotalAPIs:     len(results),
		EnabledCount:  len(enabledAPIs),
		DisabledCount: len(disabledAPIs),
		ErrorCount:    errorCount,
		TotalCost:     totalCost,
		Currency:      "USD",
	}

	report.EnabledAPIs = enabledAPIs
	report.DisabledAPIs = disabledAPIs
	report.CostAnalysis = CostAnalysis{
		TotalEstimatedCost: totalCost,
		UnlimitedCostAPIs:  unlimitedCostAPIs,
		HighCostAPIs:       highCostAPIs,
		CostBreakdown:      costBreakdown,
	}

	// Generate recommendations
	report.Recommendations = generateRecommendations(report)

	return report
}

// generateRecommendations creates actionable recommendations based on the analysis
func generateRecommendations(report *Report) []string {
	var recommendations []string

	// Check for unlimited cost APIs
	if len(report.CostAnalysis.UnlimitedCostAPIs) > 0 {
		recommendations = append(recommendations,
			fmt.Sprintf("⚠️  CRITICAL: Found %d APIs with unlimited cost potential. Review and set usage limits immediately:", len(report.CostAnalysis.UnlimitedCostAPIs)))

		for _, api := range report.CostAnalysis.UnlimitedCostAPIs {
			recommendations = append(recommendations,
				fmt.Sprintf("   - %s: %s", api.DisplayName, api.CostInfo.PricingDetails))
		}
	}

	// Check for high cost APIs
	if len(report.CostAnalysis.HighCostAPIs) > 0 {
		recommendations = append(recommendations,
			fmt.Sprintf("💰 High cost APIs detected (%d APIs with >$50 estimated cost):", len(report.CostAnalysis.HighCostAPIs)))

		for _, api := range report.CostAnalysis.HighCostAPIs {
			recommendations = append(recommendations,
				fmt.Sprintf("   - %s: $%.2f/month", api.DisplayName, api.CostInfo.EstimatedCost))
		}
	}

	// Check total cost
	if report.Summary.TotalCost > 500 {
		recommendations = append(recommendations,
			fmt.Sprintf("💸 Total estimated monthly cost is high: $%.2f. Consider reviewing usage patterns.", report.Summary.TotalCost))
	}

	// Check for disabled APIs that might be needed
	if len(report.DisabledAPIs) > 0 {
		recommendations = append(recommendations,
			fmt.Sprintf("🔒 %d APIs are currently disabled. Review if any are needed for your application.", len(report.DisabledAPIs)))
	}

	// General recommendations
	recommendations = append(recommendations,
		"📊 Set up billing alerts and budget limits in Google Cloud Console")
	recommendations = append(recommendations,
		"🔍 Regularly monitor API usage and costs")
	recommendations = append(recommendations,
		"⚡ Consider using quotas and rate limiting for high-cost APIs")

	return recommendations
}

// SaveReport saves the report to a JSON file
func SaveReport(report *Report, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create report file: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(report); err != nil {
		return fmt.Errorf("failed to encode report: %v", err)
	}

	return nil
}

// PrintReport prints a formatted report to the console with colors and validation
func PrintReport(report *Report) {
	// ANSI color codes
	const (
		reset    = "\033[0m"
		bold     = "\033[1m"
		red      = "\033[31m"
		green    = "\033[32m"
		yellow   = "\033[33m"
		blue     = "\033[34m"
		magenta  = "\033[35m"
		cyan     = "\033[36m"
		white    = "\033[37m"
		bgRed    = "\033[41m"
		bgYellow = "\033[43m"
	)

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Printf(bold + cyan + "📊 GOOGLE API CHECKER - ANALYSIS REPORT" + reset + "\n")
	fmt.Println(strings.Repeat("=", 80))

	// Summary
	fmt.Printf("\n" + bold + "📈 SUMMARY:" + reset + "\n")
	fmt.Printf("   Total APIs checked: %s%d%s\n", blue, report.Summary.TotalAPIs, reset)
	fmt.Printf("   Enabled APIs: %s%d%s\n", green, report.Summary.EnabledCount, reset)
	fmt.Printf("   Disabled APIs: %s%d%s\n", yellow, report.Summary.DisabledCount, reset)
	fmt.Printf("   Errors: %s%d%s\n", red, report.Summary.ErrorCount, reset)
	fmt.Printf("   Total estimated monthly cost: %s$%.2f %s%s\n", magenta, report.Summary.TotalCost, report.Summary.Currency, reset)

	// Cost Analysis
	if len(report.CostAnalysis.UnlimitedCostAPIs) > 0 {
		fmt.Printf("\n"+bgRed+white+bold+"⚠️  UNLIMITED COST APIS (%d):"+reset+"\n", len(report.CostAnalysis.UnlimitedCostAPIs))
		for _, api := range report.CostAnalysis.UnlimitedCostAPIs {
			fmt.Printf(bold+red+"   • %s"+reset+"\n", api.DisplayName)
			fmt.Printf("     %s%s%s\n", yellow, api.CostInfo.PricingDetails, reset)
		}
	}

	if len(report.CostAnalysis.HighCostAPIs) > 0 {
		fmt.Printf("\n" + bgYellow + bold + "💰 HIGH COST APIS (>$50/month):" + reset + "\n")
		for _, api := range report.CostAnalysis.HighCostAPIs {
			fmt.Printf(bold+magenta+"   • %s: $%.2f/month"+reset+"\n", api.DisplayName, api.CostInfo.EstimatedCost)
		}
	}

	// Recommendations
	if len(report.Recommendations) > 0 {
		fmt.Printf("\n" + bold + blue + "💡 RECOMMENDATIONS:" + reset + "\n")
		for _, rec := range report.Recommendations {
			fmt.Printf("   %s%s%s\n", green, rec, reset)
		}
	}

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Printf("Report generated at: %s\n", report.GeneratedAt.Format("2006-01-02 15:04:05"))
	fmt.Println(strings.Repeat("=", 80))
}
