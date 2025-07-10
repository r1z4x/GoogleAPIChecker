# Google API Checker

A Go CLI application that checks the status of all Google API products using multithreading and provides cost analysis with recommendations.

## Features

- 🔍 **Multithreaded API Scanning**: Checks all Google API products concurrently
- 💰 **Cost Analysis**: Calculates estimated costs and identifies unlimited cost risks
- 📊 **Comprehensive Reporting**: Generates detailed reports with recommendations
- ⚠️ **Risk Detection**: Identifies APIs with unlimited cost potential
- 🎯 **CLI Interface**: Easy-to-use command line interface
- 📤 **Export Features**: CSV, PDF, and text export capabilities

## Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd GoogleAPIChecker
```

2. Install dependencies:
```bash
go mod tidy
```

3. Build the application:
```bash
go build -o googleapichecker
```

## Usage

### Basic Usage

```bash
./googleapichecker --token YOUR_GOOGLE_API_TOKEN
```

### Advanced Usage

```bash
./googleapichecker \
  --token YOUR_GOOGLE_API_TOKEN \
  --threads 20 \
  --output results.json \
  --export both \
  --export-dir ./reports
```

### Command Line Options

- `--token, -t`: Google API token (required)
- `--threads, -n`: Number of concurrent threads (default: 10)
- `--output, -o`: Output file path (default: results.json)
- `--export, -e`: Export format: csv, pdf, both
- `--export-dir, -d`: Export directory (default: current directory)

## Output Files

The application generates several output files:

1. **Results File** (`results.json`): Raw API checking results
2. **Report File** (`results_report.json`): Analyzed report with recommendations
3. **CSV Export** (`google_api_checker_YYYYMMDD_HHMMSS.csv`): Detailed results in CSV format
4. **PDF Export** (`google_api_checker_YYYYMMDD_HHMMSS.pdf`): Professional PDF report
5. **Summary Export** (`summary_YYYYMMDD_HHMMSS.txt`): Text summary report

### Sample Report Output

```
================================================================================
📊 GOOGLE API CHECKER - ANALYSIS REPORT
================================================================================

📈 SUMMARY:
   Total APIs checked: 85
   Enabled APIs: 45
   Disabled APIs: 40
   Errors: 0
   Total estimated monthly cost: $290.00 USD

⚠️  UNLIMITED COST APIS (2):
   • BigQuery API
     ⚠️ WARNING: No usage limits - potential unlimited costs
   • Cloud Firestore API
     ⚠️ WARNING: No usage limits - potential unlimited costs

💰 HIGH COST APIS (>$50/month):
   • Compute Engine API: $150.00/month
   • Maps JavaScript API: $100.00/month

💡 RECOMMENDATIONS:
   ⚠️  CRITICAL: Found 2 APIs with unlimited cost potential. Review and set usage limits immediately:
   - BigQuery API: ⚠️ WARNING: No usage limits - potential unlimited costs
   - Cloud Firestore API: ⚠️ WARNING: No usage limits - potential unlimited costs
   💰 High cost APIs detected (2 APIs with >$50 estimated cost):
   - Compute Engine API: $150.00/month
   - Maps JavaScript API: $100.00/month
   💸 Total estimated monthly cost is high: $290.00. Consider reviewing usage patterns.
   🔒 40 APIs are currently disabled. Review if any are needed for your application.
   📊 Set up billing alerts and budget limits in Google Cloud Console
   🔍 Regularly monitor API usage and costs
   ⚡ Consider using quotas and rate limiting for high-cost APIs

================================================================================
Report generated at: 2024-01-15 14:30:25
================================================================================
```

## Cost Analysis Features

### Unlimited Cost Detection
The application identifies APIs that have no usage limits and could potentially incur unlimited costs. These are marked with warnings and require immediate attention.

### High Cost API Detection
APIs with estimated monthly costs above $50 are flagged for review.

### Cost Breakdown
Detailed cost analysis for each API including:
- Estimated monthly cost
- Pricing details
- Currency information

## Multithreading

The application uses Go's goroutines for concurrent API checking:
- Configurable number of worker threads
- Efficient resource utilization
- Progress tracking during execution

## API Coverage

The application checks a comprehensive list of Google APIs including:
- Compute Engine APIs
- Storage and Database APIs
- Machine Learning APIs
- Analytics APIs
- Maps and Location APIs
- Firebase APIs
- And many more...

## Security

- API tokens are handled securely
- No sensitive information is logged
- Results are saved locally

## Error Handling

- Graceful handling of API errors
- Detailed error reporting
- Continuation of checking process even if some APIs fail

## Requirements

- Go 1.21 or higher
- Google Cloud API token with appropriate permissions
- Internet connection for API access

## Development

### Project Structure

```
GoogleAPIChecker/
├── main.go          # CLI entry point
├── checker.go       # Core API checking logic
├── report.go        # Report generation and analysis
├── go.mod           # Go module file
└── README.md        # This file
```

### Adding New APIs

To add new APIs to the checker, modify the `getAvailableAPIs()` function in `checker.go`.

### Customizing Cost Analysis

To customize cost analysis, modify the `getCostInfo()` function in `checker.go`.

## License

This project is licensed under the MIT License.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## Support

For issues and questions, please create an issue in the repository. 