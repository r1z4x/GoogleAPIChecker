package main

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// ProgressBar represents a progress bar for API checking
type ProgressBar struct {
	total        int
	current      int
	mu           sync.Mutex
	startTime    time.Time
	spinner      []string
	spinnerIndex int
}

// NewProgressBar creates a new progress bar
func NewProgressBar(total int) *ProgressBar {
	return &ProgressBar{
		total:        total,
		current:      0,
		startTime:    time.Now(),
		spinner:      []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		spinnerIndex: 0,
	}
}

// Update updates the progress bar
func (p *ProgressBar) Update() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.current++
	p.spinnerIndex = (p.spinnerIndex + 1) % len(p.spinner)

	// Calculate progress percentage
	percentage := float64(p.current) / float64(p.total) * 100

	// Calculate elapsed time
	elapsed := time.Since(p.startTime)

	// Calculate estimated time remaining
	var eta time.Duration
	if p.current > 0 {
		eta = time.Duration(float64(elapsed) * float64(p.total-p.current) / float64(p.current))
	}

	// Create progress bar
	barWidth := 30
	filled := int(float64(barWidth) * percentage / 100)
	bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)

	// Clear line and print progress
	fmt.Printf("\r%s Scanning APIs... [%s] %d/%d (%.1f%%) | Elapsed: %s | ETA: %s",
		p.spinner[p.spinnerIndex],
		bar,
		p.current,
		p.total,
		percentage,
		formatDuration(elapsed),
		formatDuration(eta))
}

// Complete marks the progress as complete
func (p *ProgressBar) Complete() {
	p.mu.Lock()
	defer p.mu.Unlock()

	elapsed := time.Since(p.startTime)

	// Clear line and print completion message
	fmt.Printf("\r✅ Scanning completed! %d APIs checked in %s\n", p.total, formatDuration(elapsed))
}

// formatDuration formats duration in a human-readable way
func formatDuration(d time.Duration) string {
	if d < time.Second {
		return "0s"
	}
	if d < time.Minute {
		return fmt.Sprintf("%.0fs", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%.0fm", d.Minutes())
	}
	return fmt.Sprintf("%.0fh", d.Hours())
}

// LoadingSpinner shows a simple loading spinner
func LoadingSpinner(message string, done chan bool) {
	spinner := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	i := 0

	for {
		select {
		case <-done:
			fmt.Printf("\r%s Done!\n", strings.Repeat(" ", len(message)+10))
			return
		default:
			fmt.Printf("\r%s %s", spinner[i], message)
			time.Sleep(100 * time.Millisecond)
			i = (i + 1) % len(spinner)
		}
	}
}

// StatusUpdate shows a status update with timestamp
func StatusUpdate(message string) {
	timestamp := time.Now().Format("15:04:05")
	fmt.Printf("[%s] %s\n", timestamp, message)
}

// ClearLine clears the current line
func ClearLine() {
	fmt.Printf("\r%s", strings.Repeat(" ", 100))
	fmt.Printf("\r")
}
