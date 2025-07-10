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

// generateHTMLReport creates an HTML table report
func generateHTMLReport(results []APIResult, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create HTML file: %v", err)
	}
	defer file.Close()

	// Calculate statistics
	var enabledCount, disabledCount, errorCount int
	var totalCost float64
	for _, result := range results {
		if result.Error != "" {
			errorCount++
		} else if result.Enabled {
			enabledCount++
			if result.CostInfo.HasPricing {
				totalCost += result.CostInfo.EstimatedCost
			}
		} else {
			disabledCount++
		}
	}

	// Generate HTML content
	htmlContent := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Google API Checker Report</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <script defer src="https://unpkg.com/alpinejs@3.x.x/dist/cdn.min.js"></script>
</head>
<body class="bg-gray-100 min-h-screen">
    <script id="apidata" type="application/json">%s</script>
    <div class="container mx-auto px-4 py-8" x-data="apiChecker()" x-init="init()">
        <div class="max-w-7xl mx-auto">
            <!-- Header -->
            <div class="bg-gradient-to-r from-blue-600 to-purple-600 text-white rounded-lg p-8 mb-8 text-center">
                <h1 class="text-4xl font-bold mb-2">🔍 Google API Checker Report</h1>
                <p class="text-lg opacity-90">Generated on %s</p>
            </div>
            <!-- Stats Cards -->
            <div class="grid grid-cols-1 md:grid-cols-5 gap-6 mb-8">
                <div class="bg-white rounded-lg p-6 shadow-md border-l-4 border-blue-500">
                    <div class="text-3xl font-bold text-blue-600" x-text="stats.total"></div>
                    <div class="text-gray-600 mt-2">Total APIs</div>
                </div>
                <div class="bg-white rounded-lg p-6 shadow-md border-l-4 border-green-500">
                    <div class="text-3xl font-bold text-green-600" x-text="stats.enabled"></div>
                    <div class="text-gray-600 mt-2">Enabled</div>
                </div>
                <div class="bg-white rounded-lg p-6 shadow-md border-l-4 border-red-500">
                    <div class="text-3xl font-bold text-red-600" x-text="stats.disabled"></div>
                    <div class="text-gray-600 mt-2">Disabled</div>
                </div>
                <div class="bg-white rounded-lg p-6 shadow-md border-l-4 border-yellow-500">
                    <div class="text-3xl font-bold text-yellow-600" x-text="stats.errors"></div>
                    <div class="text-gray-600 mt-2">Errors</div>
                </div>
                <div class="bg-white rounded-lg p-6 shadow-md border-l-4 border-purple-500">
                    <div class="text-3xl font-bold text-purple-600" x-text="'$' + (typeof stats.totalCost === 'number' ? stats.totalCost.toFixed(2) : '0.00')"></div>
                    <div class="text-gray-600 mt-2">Total Cost (USD)</div>
                </div>
            </div>
            <!-- Search Box -->
            <div class="mb-6">
                <input 
                    type="text" 
                    x-model="searchTerm"
                    placeholder="Search APIs..." 
                    class="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                >
            </div>
            <!-- Tabs -->
            <div class="flex space-x-2 mb-6">
                <button 
                    @click="activeTab = 'all'"
                    :class="activeTab === 'all' ? 'bg-blue-600 text-white' : 'bg-gray-200 text-gray-700'"
                    class="px-6 py-3 rounded-lg font-medium transition-colors"
                >
                    All APIs
                </button>
                <button 
                    @click="activeTab = 'enabled'"
                    :class="activeTab === 'enabled' ? 'bg-green-600 text-white' : 'bg-gray-200 text-gray-700'"
                    class="px-6 py-3 rounded-lg font-medium transition-colors"
                >
                    Enabled
                </button>
                <button 
                    @click="activeTab = 'disabled'"
                    :class="activeTab === 'disabled' ? 'bg-red-600 text-white' : 'bg-gray-200 text-gray-700'"
                    class="px-6 py-3 rounded-lg font-medium transition-colors"
                >
                    Disabled
                </button>
                <button 
                    @click="activeTab = 'errors'"
                    :class="activeTab === 'errors' ? 'bg-yellow-600 text-white' : 'bg-gray-200 text-gray-700'"
                    class="px-6 py-3 rounded-lg font-medium transition-colors"
                >
                    Errors
                </button>
            </div>
            <!-- Results Count -->
            <div class="mb-4 text-gray-600">
                Showing <span class="font-semibold" x-text="filteredApis.length"></span> of <span class="font-semibold" x-text="stats.total"></span> APIs
            </div>
            <!-- Table -->
            <div class="bg-white rounded-lg shadow-md overflow-hidden">
                <div class="overflow-x-auto">
                    <table class="w-full">
                        <thead class="bg-gray-50">
                            <tr>
                                <th class="px-6 py-4 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">API Name</th>
                                <th class="px-6 py-4 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Display Name</th>
                                <th class="px-6 py-4 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
                                <th class="px-6 py-4 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Cost (USD)</th>
                                <th class="px-6 py-4 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Pricing Details</th>
                                <th class="px-6 py-4 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Checked At</th>
                            </tr>
                        </thead>
                        <tbody class="bg-white divide-y divide-gray-200">
                            <template x-for="(api, idx) in filteredApis" :key="api.name + idx">
                                <tr class="hover:bg-gray-50">
                                    <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900" x-text="api.name"></td>
                                    <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900" x-text="api.displayName"></td>
                                    <td class="px-6 py-4 whitespace-nowrap">
                                        <span 
                                            :class="{
                                                'bg-green-100 text-green-800': api.status === 'ENABLED',
                                                'bg-red-100 text-red-800': api.status === 'DISABLED',
                                                'bg-yellow-100 text-yellow-800': api.status === 'ERROR'
                                            }"
                                            class="px-2 py-1 text-xs font-medium rounded-full"
                                            x-text="api.status"
                                        ></span>
                                    </td>
                                    <td class="px-6 py-4 whitespace-nowrap text-sm">
                                        <span 
                                            :class="{
                                                'text-red-600 font-bold': api.costInfo.estimatedCost > 50,
                                                'text-yellow-600 font-bold': api.costInfo.estimatedCost > 10 && api.costInfo.estimatedCost <= 50,
                                                'text-green-600': api.costInfo.estimatedCost <= 10
                                            }"
                                            x-text="'$' + (typeof api.costInfo.estimatedCost === 'number' ? api.costInfo.estimatedCost.toFixed(2) : '0.00')"
                                        ></span>
                                    </td>
                                    <td class="px-6 py-4 text-sm text-gray-900" x-text="api.costInfo.pricingDetails"></td>
                                    <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500" x-text="new Date(api.checkedAt).toLocaleString()"></td>
                                </tr>
                            </template>
                        </tbody>
                    </table>
                </div>
            </div>
        </div>
    </div>
    <script>
    function apiChecker() {
        return {
            apis: [],
            activeTab: 'all',
            searchTerm: '',
            get filteredApis() {
                return this.apis.filter(api => {
                    const matchesSearch = !this.searchTerm || 
                        api.name.toLowerCase().includes(this.searchTerm.toLowerCase()) ||
                        api.displayName.toLowerCase().includes(this.searchTerm.toLowerCase());
                    if (this.activeTab === 'all') return matchesSearch;
                    if (this.activeTab === 'enabled') return matchesSearch && api.status === 'ENABLED';
                    if (this.activeTab === 'disabled') return matchesSearch && api.status === 'DISABLED';
                    if (this.activeTab === 'errors') return matchesSearch && api.status === 'ERROR';
                    return matchesSearch;
                });
            },
            get stats() {
                const total = this.apis.length;
                const enabled = this.apis.filter(api => api.status === 'ENABLED').length;
                const disabled = this.apis.filter(api => api.status === 'DISABLED').length;
                const errors = this.apis.filter(api => api.status === 'ERROR').length;
                const totalCost = this.apis.reduce((sum, api) => sum + (api.costInfo.estimatedCost || 0), 0);
                return { total, enabled, disabled, errors, totalCost };
            },
            init() {
                this.apis = JSON.parse(document.getElementById('apidata').textContent);
            }
        }
    }
    </script>
</body>
</html>`, generateJSONData(results), time.Now().Format("2006-01-02 15:04:05"))

	_, err = file.WriteString(htmlContent)
	return err
}

// generateJSONData converts API results to JSON for Alpine.js
func generateJSONData(results []APIResult) string {
	type APIData struct {
		Name        string    `json:"name"`
		DisplayName string    `json:"displayName"`
		Status      string    `json:"status"`
		Enabled     bool      `json:"enabled"`
		CostInfo    CostInfo  `json:"costInfo"`
		CheckedAt   time.Time `json:"checkedAt"`
		Error       string    `json:"error,omitempty"`
	}

	var apiData []APIData
	for _, result := range results {
		apiData = append(apiData, APIData{
			Name:        result.Name,
			DisplayName: result.DisplayName,
			Status:      result.Status,
			Enabled:     result.Enabled,
			CostInfo:    result.CostInfo,
			CheckedAt:   result.CheckedAt,
			Error:       result.Error,
		})
	}

	jsonData, err := json.Marshal(apiData)
	if err != nil {
		return "[]"
	}
	return string(jsonData)
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
