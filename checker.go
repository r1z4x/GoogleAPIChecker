package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// APIResult represents the result of checking a single API
type APIResult struct {
	Name        string    `json:"name"`
	DisplayName string    `json:"display_name"`
	Status      string    `json:"status"`
	Enabled     bool      `json:"enabled"`
	CostInfo    CostInfo  `json:"cost_info"`
	CheckedAt   time.Time `json:"checked_at"`
	Error       string    `json:"error,omitempty"`
}

// CostInfo contains pricing and cost calculation information
type CostInfo struct {
	HasPricing     bool    `json:"has_pricing"`
	UnlimitedCost  bool    `json:"unlimited_cost"`
	EstimatedCost  float64 `json:"estimated_cost"`
	Currency       string  `json:"currency"`
	PricingDetails string  `json:"pricing_details"`
}

// GoogleAPIChecker handles the checking of Google APIs
type GoogleAPIChecker struct {
	token      string
	projectID  string
	threads    int
	client     *http.Client
	ctx        context.Context
	useRealAPI bool
}

// NewGoogleAPIChecker creates a new instance of the checker
func NewGoogleAPIChecker(token, projectID string, threads int) *GoogleAPIChecker {
	// Always use real API if token is provided
	useRealAPI := token != ""

	checker := &GoogleAPIChecker{
		token:      token,
		projectID:  projectID,
		threads:    threads,
		client:     &http.Client{Timeout: 30 * time.Second},
		ctx:        context.Background(),
		useRealAPI: useRealAPI,
	}

	return checker
}

// CheckAllAPIs performs the main checking operation with multithreading
func (c *GoogleAPIChecker) CheckAllAPIs() ([]APIResult, error) {
	fmt.Println("🔍 Discovering available Google APIs...")

	// Get list of all available APIs
	apis, err := c.getAvailableAPIs()
	if err != nil {
		return nil, fmt.Errorf("failed to get available APIs: %v", err)
	}

	fmt.Printf("📋 Found %d APIs to check\n", len(apis))

	// Create channels for work distribution and results collection
	jobs := make(chan string, len(apis))
	results := make(chan APIResult, len(apis))

	// Start worker goroutines
	var wg sync.WaitGroup
	for i := 0; i < c.threads; i++ {
		wg.Add(1)
		go c.worker(&wg, jobs, results)
	}

	// Send jobs to workers
	go func() {
		defer close(jobs)
		for _, api := range apis {
			jobs <- api
		}
	}()

	// Collect results
	go func() {
		wg.Wait()
		close(results)
	}()

	// Create progress bar
	progress := NewProgressBar(len(apis))

	// Gather all results
	var allResults []APIResult
	for result := range results {
		allResults = append(allResults, result)
		progress.Update()
	}

	// Complete progress bar
	progress.Complete()

	return allResults, nil
}

// worker processes API checking jobs
func (c *GoogleAPIChecker) worker(wg *sync.WaitGroup, jobs <-chan string, results chan<- APIResult) {
	defer wg.Done()

	for apiName := range jobs {
		result := c.checkSingleAPI(apiName)
		results <- result
	}
}

// checkSingleAPI checks the status and cost of a single API
func (c *GoogleAPIChecker) checkSingleAPI(apiName string) APIResult {
	result := APIResult{
		Name:      apiName,
		CheckedAt: time.Now(),
	}

	// Check if API is enabled
	enabled, err := c.isAPIEnabled(apiName)
	if err != nil {
		result.Error = err.Error()
		result.Status = "ERROR"
		return result
	}

	result.Enabled = enabled
	if enabled {
		result.Status = "ENABLED"
	} else {
		result.Status = "DISABLED"
	}

	// Get API display name
	result.DisplayName = c.getAPIDisplayName(apiName)

	// Check cost information
	costInfo, err := c.getCostInfo(apiName)
	if err != nil {
		result.CostInfo = CostInfo{
			HasPricing: false,
		}
	} else {
		result.CostInfo = costInfo
	}

	return result
}

// getAvailableAPIs returns a list of all available Google APIs
func (c *GoogleAPIChecker) getAvailableAPIs() ([]string, error) {
	// If we have real API access, try to get the actual list
	if c.useRealAPI {
		return c.getAvailableAPIsReal()
	}

	// Fallback to static list for testing
	return c.getAvailableAPIsStatic()
}

// getAvailableAPIsReal gets the actual list of APIs from Google Cloud
func (c *GoogleAPIChecker) getAvailableAPIsReal() ([]string, error) {
	var url string

	if c.projectID != "" {
		// Use Service Usage API with project ID
		url = fmt.Sprintf("https://serviceusage.googleapis.com/v1/projects/%s/services", c.projectID)
	} else {
		// Use Discovery API to get all available APIs
		url = "https://www.googleapis.com/discovery/v1/apis"
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Add("X-Goog-Api-Key", c.token)
	req.Header.Add("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get API list: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to get API list, status: %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse API list response: %v", err)
	}

	var apis []string

	if c.projectID != "" {
		// Parse Service Usage API response
		if services, ok := result["services"].([]interface{}); ok {
			for _, service := range services {
				if serviceMap, ok := service.(map[string]interface{}); ok {
					if name, ok := serviceMap["name"].(string); ok {
						apis = append(apis, name)
					}
				}
			}
		}
	} else {
		// Parse Discovery API response
		if items, ok := result["items"].([]interface{}); ok {
			for _, item := range items {
				if itemMap, ok := item.(map[string]interface{}); ok {
					if name, ok := itemMap["name"].(string); ok {
						apis = append(apis, name+".googleapis.com")
					}
				}
			}
		}
	}

	return apis, nil
}

// getAvailableAPIsStatic returns a static list of common Google APIs
func (c *GoogleAPIChecker) getAvailableAPIsStatic() ([]string, error) {
	apis := []string{
		"compute.googleapis.com",
		"storage.googleapis.com",
		"bigquery.googleapis.com",
		"pubsub.googleapis.com",
		"cloudfunctions.googleapis.com",
		"cloudrun.googleapis.com",
		"container.googleapis.com",
		"datastore.googleapis.com",
		"firestore.googleapis.com",
		"cloudsql.googleapis.com",
		"cloudbuild.googleapis.com",
		"cloudtasks.googleapis.com",
		"cloudscheduler.googleapis.com",
		"cloudkms.googleapis.com",
		"cloudiot.googleapis.com",
		"cloudtrace.googleapis.com",
		"clouddebugger.googleapis.com",
		"cloudprofiler.googleapis.com",
		"cloudmonitoring.googleapis.com",
		"cloudlogging.googleapis.com",
		"translate.googleapis.com",
		"vision.googleapis.com",
		"speech.googleapis.com",
		"language.googleapis.com",
		"ml.googleapis.com",
		"automl.googleapis.com",
		"dataflow.googleapis.com",
		"dataproc.googleapis.com",
		"dataprep.googleapis.com",
		"datalab.googleapis.com",
		"datacatalog.googleapis.com",
		"datastudio.googleapis.com",
		"analytics.googleapis.com",
		"analyticsadmin.googleapis.com",
		"searchconsole.googleapis.com",
		"webmasters.googleapis.com",
		"indexing.googleapis.com",
		"customsearch.googleapis.com",
		"pagespeedonline.googleapis.com",
		"siteverification.googleapis.com",
		"websecurityscanner.googleapis.com",
		"clouderrorreporting.googleapis.com",
		"cloudresourcemanager.googleapis.com",
		"iam.googleapis.com",
		"serviceusage.googleapis.com",
		"cloudbilling.googleapis.com",
		"billingbudgets.googleapis.com",
		"recommender.googleapis.com",
		"recommendationengine.googleapis.com",
		"retail.googleapis.com",
		"documentai.googleapis.com",
		"videointelligence.googleapis.com",
		"gameservices.googleapis.com",
		"playablelocations.googleapis.com",
		"places.googleapis.com",
		"geocoding.googleapis.com",
		"geolocation.googleapis.com",
		"maps.googleapis.com",
		"directions.googleapis.com",
		"distancematrix.googleapis.com",
		"elevation.googleapis.com",
		"timezone.googleapis.com",
		"staticmap.googleapis.com",
		"streetview.googleapis.com",
		"roads.googleapis.com",
		"fcm.googleapis.com",
		"firebase.googleapis.com",
		"firebaseappcheck.googleapis.com",
		"firebaseauth.googleapis.com",
		"firebasehosting.googleapis.com",
		"firebaseml.googleapis.com",
		"firebaserules.googleapis.com",
		"firebasestorage.googleapis.com",
		"identitytoolkit.googleapis.com",
		"securetoken.googleapis.com",
		"appengine.googleapis.com",
		"cloudapis.googleapis.com",
	}

	return apis, nil
}

// isAPIEnabled checks if a specific API is enabled using Google Cloud Service Usage API
func (c *GoogleAPIChecker) isAPIEnabled(apiName string) (bool, error) {
	// If we have a real API token, use real API calls
	if c.useRealAPI {
		return c.checkAPIEnabledReal(apiName)
	}

	// Fallback to simulation for testing
	return c.checkAPIEnabledSimulated(apiName)
}

// checkAPIEnabledReal checks API status using real Google Cloud Service Usage API
func (c *GoogleAPIChecker) checkAPIEnabledReal(apiName string) (bool, error) {
	var url string

	if c.projectID != "" {
		// Use Service Usage API with project ID
		url = fmt.Sprintf("https://serviceusage.googleapis.com/v1/projects/%s/services/%s", c.projectID, apiName)
	} else {
		// Use Discovery API to check if API exists
		url = fmt.Sprintf("https://www.googleapis.com/discovery/v1/apis/%s/v1", strings.TrimSuffix(apiName, ".googleapis.com"))
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, fmt.Errorf("failed to create request: %v", err)
	}

	// Add API key to request (Google Cloud API uses API key, not Bearer token)
	req.Header.Add("X-Goog-Api-Key", c.token)
	req.Header.Add("Content-Type", "application/json")

	// Make the actual HTTP request
	resp, err := c.client.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to make API request: %v", err)
	}
	defer resp.Body.Close()

	if c.projectID != "" {
		// Check if API is enabled based on response
		if resp.StatusCode == 200 {
			// Parse response body to check if service is enabled
			var result map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				return false, fmt.Errorf("failed to parse response: %v", err)
			}

			// Check if the service is enabled
			if state, ok := result["state"].(string); ok {
				return state == "ENABLED", nil
			}
			return true, nil // Default to enabled if state not found
		} else if resp.StatusCode == 404 {
			// Service not found, consider it disabled
			return false, nil
		} else {
			// Other error status codes
			return false, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
		}
	} else {
		// Without project ID, check if API is available (not necessarily enabled)
		if resp.StatusCode == 200 {
			// API exists and is available, but we can't determine if it's enabled without project ID
			// For now, we'll consider it as "available" but not necessarily "enabled"
			return false, nil // Consider as disabled since we can't verify actual enable status
		} else if resp.StatusCode == 404 {
			return false, nil // API not found
		} else {
			return false, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
		}
	}
}

// checkAPIEnabledSimulated provides simulated API status for testing
func (c *GoogleAPIChecker) checkAPIEnabledSimulated(apiName string) (bool, error) {
	time.Sleep(100 * time.Millisecond) // Simulate API call

	// Simulate some APIs being enabled and others disabled
	enabledAPIs := map[string]bool{
		"compute.googleapis.com":        true,
		"storage.googleapis.com":        true,
		"bigquery.googleapis.com":       true,
		"pubsub.googleapis.com":         false,
		"cloudfunctions.googleapis.com": true,
		"cloudrun.googleapis.com":       false,
		"container.googleapis.com":      true,
		"datastore.googleapis.com":      false,
		"firestore.googleapis.com":      true,
		"cloudsql.googleapis.com":       true,
	}

	if enabled, exists := enabledAPIs[apiName]; exists {
		return enabled, nil
	}

	// Default to enabled for unknown APIs
	return true, nil
}

// getAPIDisplayName returns the display name for an API
func (c *GoogleAPIChecker) getAPIDisplayName(apiName string) string {
	displayNames := map[string]string{
		"compute.googleapis.com":        "Compute Engine API",
		"storage.googleapis.com":        "Cloud Storage API",
		"bigquery.googleapis.com":       "BigQuery API",
		"pubsub.googleapis.com":         "Cloud Pub/Sub API",
		"cloudfunctions.googleapis.com": "Cloud Functions API",
		"cloudrun.googleapis.com":       "Cloud Run API",
		"container.googleapis.com":      "Kubernetes Engine API",
		"datastore.googleapis.com":      "Cloud Datastore API",
		"firestore.googleapis.com":      "Cloud Firestore API",
		"cloudsql.googleapis.com":       "Cloud SQL API",
		"cloudbuild.googleapis.com":     "Cloud Build API",
		"cloudtasks.googleapis.com":     "Cloud Tasks API",
		"cloudscheduler.googleapis.com": "Cloud Scheduler API",
		"cloudkms.googleapis.com":       "Cloud KMS API",
		"cloudiot.googleapis.com":       "Cloud IoT API",
		"translate.googleapis.com":      "Cloud Translation API",
		"vision.googleapis.com":         "Cloud Vision API",
		"speech.googleapis.com":         "Cloud Speech API",
		"language.googleapis.com":       "Natural Language API",
		"ml.googleapis.com":             "Machine Learning API",
		"automl.googleapis.com":         "AutoML API",
		"dataflow.googleapis.com":       "Dataflow API",
		"dataproc.googleapis.com":       "Dataproc API",
		"analytics.googleapis.com":      "Google Analytics API",
		"maps.googleapis.com":           "Maps JavaScript API",
		"firebase.googleapis.com":       "Firebase API",
		"appengine.googleapis.com":      "App Engine API",
	}

	if displayName, exists := displayNames[apiName]; exists {
		return displayName
	}

	// Return a formatted version of the API name if no display name is found
	return apiName
}

// getCostInfo retrieves cost information for an API
func (c *GoogleAPIChecker) getCostInfo(apiName string) (CostInfo, error) {
	// In a real implementation, you would query the Cloud Billing API
	// For now, we'll simulate cost information

	time.Sleep(50 * time.Millisecond) // Simulate API call

	// Simulate cost data for different APIs
	costData := map[string]CostInfo{
		"compute.googleapis.com": {
			HasPricing:     true,
			UnlimitedCost:  false,
			EstimatedCost:  150.0,
			Currency:       "USD",
			PricingDetails: "Pay per use - $0.05 per hour for standard instances",
		},
		"storage.googleapis.com": {
			HasPricing:     true,
			UnlimitedCost:  false,
			EstimatedCost:  25.0,
			Currency:       "USD",
			PricingDetails: "Pay per use - $0.02 per GB per month",
		},
		"bigquery.googleapis.com": {
			HasPricing:     true,
			UnlimitedCost:  true,
			EstimatedCost:  0.0,
			Currency:       "USD",
			PricingDetails: "⚠️ WARNING: No usage limits - potential unlimited costs",
		},
		"pubsub.googleapis.com": {
			HasPricing:     true,
			UnlimitedCost:  false,
			EstimatedCost:  10.0,
			Currency:       "USD",
			PricingDetails: "Pay per use - $0.40 per million messages",
		},
		"cloudfunctions.googleapis.com": {
			HasPricing:     true,
			UnlimitedCost:  false,
			EstimatedCost:  5.0,
			Currency:       "USD",
			PricingDetails: "Pay per use - $0.40 per million invocations",
		},
		"firestore.googleapis.com": {
			HasPricing:     true,
			UnlimitedCost:  true,
			EstimatedCost:  0.0,
			Currency:       "USD",
			PricingDetails: "⚠️ WARNING: No usage limits - potential unlimited costs",
		},
		"maps.googleapis.com": {
			HasPricing:     true,
			UnlimitedCost:  false,
			EstimatedCost:  100.0,
			Currency:       "USD",
			PricingDetails: "Pay per use - $5.00 per 1000 requests",
		},
		// Additional unlimited cost APIs for testing
		"datastore.googleapis.com": {
			HasPricing:     true,
			UnlimitedCost:  true,
			EstimatedCost:  0.0,
			Currency:       "USD",
			PricingDetails: "⚠️ WARNING: No usage limits - potential unlimited costs",
		},
		"cloudsql.googleapis.com": {
			HasPricing:     true,
			UnlimitedCost:  false,
			EstimatedCost:  75.0,
			Currency:       "USD",
			PricingDetails: "Pay per use - $0.10 per hour for standard instances",
		},
		"container.googleapis.com": {
			HasPricing:     true,
			UnlimitedCost:  false,
			EstimatedCost:  50.0,
			Currency:       "USD",
			PricingDetails: "Pay per use - $0.10 per hour for standard clusters",
		},
		"vision.googleapis.com": {
			HasPricing:     true,
			UnlimitedCost:  false,
			EstimatedCost:  30.0,
			Currency:       "USD",
			PricingDetails: "Pay per use - $1.50 per 1000 requests",
		},
		"speech.googleapis.com": {
			HasPricing:     true,
			UnlimitedCost:  false,
			EstimatedCost:  20.0,
			Currency:       "USD",
			PricingDetails: "Pay per use - $0.006 per 15 seconds",
		},
		"translate.googleapis.com": {
			HasPricing:     true,
			UnlimitedCost:  false,
			EstimatedCost:  15.0,
			Currency:       "USD",
			PricingDetails: "Pay per use - $20 per million characters",
		},
		"ml.googleapis.com": {
			HasPricing:     true,
			UnlimitedCost:  true,
			EstimatedCost:  0.0,
			Currency:       "USD",
			PricingDetails: "⚠️ WARNING: No usage limits - potential unlimited costs",
		},
		"automl.googleapis.com": {
			HasPricing:     true,
			UnlimitedCost:  true,
			EstimatedCost:  0.0,
			Currency:       "USD",
			PricingDetails: "⚠️ WARNING: No usage limits - potential unlimited costs",
		},
		"dataflow.googleapis.com": {
			HasPricing:     true,
			UnlimitedCost:  false,
			EstimatedCost:  200.0,
			Currency:       "USD",
			PricingDetails: "Pay per use - $0.06 per vCPU per hour",
		},
		"dataproc.googleapis.com": {
			HasPricing:     true,
			UnlimitedCost:  false,
			EstimatedCost:  120.0,
			Currency:       "USD",
			PricingDetails: "Pay per use - $0.10 per vCPU per hour",
		},
		"analytics.googleapis.com": {
			HasPricing:     true,
			UnlimitedCost:  false,
			EstimatedCost:  8.0,
			Currency:       "USD",
			PricingDetails: "Pay per use - $0.50 per 1000 requests",
		},
		"firebase.googleapis.com": {
			HasPricing:     true,
			UnlimitedCost:  true,
			EstimatedCost:  0.0,
			Currency:       "USD",
			PricingDetails: "⚠️ WARNING: No usage limits - potential unlimited costs",
		},
		"appengine.googleapis.com": {
			HasPricing:     true,
			UnlimitedCost:  false,
			EstimatedCost:  40.0,
			Currency:       "USD",
			PricingDetails: "Pay per use - $0.05 per instance hour",
		},
	}

	if costInfo, exists := costData[apiName]; exists {
		return costInfo, nil
	}

	// Default cost info for unknown APIs
	return CostInfo{
		HasPricing:     false,
		UnlimitedCost:  false,
		EstimatedCost:  0.0,
		Currency:       "USD",
		PricingDetails: "No pricing information available",
	}, nil
}

// SaveResults saves the results to a JSON file
func (c *GoogleAPIChecker) SaveResults(results []APIResult, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create output file: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(results); err != nil {
		return fmt.Errorf("failed to encode results: %v", err)
	}

	return nil
}
