package report

import (
	_ "embed"
	"encoding/json"
	"html/template"
	"os"
	"path/filepath"
	"sync"
	texttemplate "text/template"
	"time"
)

//go:embed template.html
var htmlTemplate string

//go:embed template.md
var mdTemplate string


// TestResult records metadata for a single test case execution.
type TestResult struct {
	Name        string        `json:"name"`
	Type        string        `json:"type"` // "API" or "UI"
	Status      string        `json:"status"` // "passed", "failed", "skipped"
	Duration    time.Duration `json:"duration"`
	DurationStr string        `json:"duration_str"`
	Error       string        `json:"error,omitempty"`
	Screenshot  string        `json:"screenshot,omitempty"`
	Timestamp   time.Time     `json:"timestamp"`
}

// Summary collects overall metrics for the execution run.
type Summary struct {
	Total            int           `json:"total"`
	Passed           int           `json:"passed"`
	Failed           int           `json:"failed"`
	Skipped          int           `json:"skipped"`
	StartTime        time.Time     `json:"start_time"`
	EndTime          time.Time     `json:"end_time"`
	TotalDuration    time.Duration `json:"total_duration"`
	TotalDurationStr string        `json:"total_duration_str"`
}

// Reporter gathers results thread-safely during test runs.
type Reporter struct {
	mu      sync.Mutex
	Summary Summary      `json:"summary"`
	Results []TestResult `json:"results"`
}

var globalReporter = &Reporter{
	Summary: Summary{
		StartTime: time.Now(),
	},
}

// GetGlobalReporter returns the thread-safe global reporter instance.
func GetGlobalReporter() *Reporter {
	return globalReporter
}

// Record saves a single test case result in the reporter.
func (r *Reporter) Record(name string, testType string, status string, duration time.Duration, err string, screenshot string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	result := TestResult{
		Name:        name,
		Type:        testType,
		Status:      status,
		Duration:    duration,
		DurationStr: duration.Round(time.Millisecond).String(),
		Error:       err,
		Screenshot:  screenshot,
		Timestamp:   time.Now(),
	}
	r.Results = append(r.Results, result)

	r.Summary.Total++
	switch status {
	case "passed":
		r.Summary.Passed++
	case "failed":
		r.Summary.Failed++
	case "skipped":
		r.Summary.Skipped++
	}
}

// GenerateReports compiles test stats and outputs JSON and HTML report files.
func (r *Reporter) GenerateReports(outputDir string) error {
	r.mu.Lock()
	r.Summary.EndTime = time.Now()
	r.Summary.TotalDuration = r.Summary.EndTime.Sub(r.Summary.StartTime)
	r.Summary.TotalDurationStr = r.Summary.TotalDuration.Round(time.Millisecond).String()
	r.mu.Unlock()

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}

	// 1. JSON Report
	jsonPath := filepath.Join(outputDir, "report.json")
	jsonData, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(jsonPath, jsonData, 0644); err != nil {
		return err
	}

	// 2. HTML Report
	htmlPath := filepath.Join(outputDir, "report.html")
	tmpl, err := template.New("report").Parse(htmlTemplate)
	if err != nil {
		return err
	}

	htmlFile, err := os.Create(htmlPath)
	if err != nil {
		return err
	}
	htmlErr := tmpl.Execute(htmlFile, r)
	htmlFile.Close()
	if htmlErr != nil {
		return htmlErr
	}

	// 3. Markdown Report
	mdPath := filepath.Join(filepath.Dir(outputDir), "test-report.md")
	mdTmpl, err := texttemplate.New("markdownReport").Parse(mdTemplate)
	if err != nil {
		return err
	}

	mdFile, err := os.Create(mdPath)
	if err != nil {
		return err
	}
	mdErr := mdTmpl.Execute(mdFile, r)
	mdFile.Close()
	if mdErr != nil {
		return mdErr
	}

	return nil
}
