package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
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
	token   string
	threads int
	client  *http.Client
	ctx     context.Context
}

// NewGoogleAPIChecker creates a new instance of the checker
func NewGoogleAPIChecker(token string, threads int) *GoogleAPIChecker {
	return &GoogleAPIChecker{
		token:   token,
		threads: threads,
		client:  &http.Client{Timeout: 30 * time.Second},
		ctx:     context.Background(),
	}
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
	// This is a simplified list - in a real implementation, you would
	// query the Google Cloud Service Usage API to get the actual list
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
		"cloudtrace.googleapis.com",
		"clouddebugger.googleapis.com",
		"cloudprofiler.googleapis.com",
		"cloudmonitoring.googleapis.com",
		"cloudlogging.googleapis.com",
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
	}

	return apis, nil
}

// isAPIEnabled checks if a specific API is enabled
func (c *GoogleAPIChecker) isAPIEnabled(apiName string) (bool, error) {
	// In a real implementation, you would use the Google Cloud Service Usage API
	// For now, we'll simulate the check
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
