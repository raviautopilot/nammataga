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
	Name          string        `json:"name"`
	Type          string        `json:"type"`          // "API" or "UI"
	Category      string        `json:"category"`      // e.g. "05-health-api_test.go"
	Description   string        `json:"description"`   // Test scenario purpose / details
	Expected      string        `json:"expected"`      // Expected HTTP status / body / behavior
	Actual        string        `json:"actual"`        // Actual HTTP status / body / behavior (Got)
	Status        string        `json:"status"`        // "passed", "failed", "skipped"
	Duration      time.Duration `json:"duration"`
	DurationStr   string        `json:"duration_str"`
	Error         string        `json:"error,omitempty"`
	FailureReason string        `json:"failure_reason,omitempty"`
	Screenshot    string        `json:"screenshot,omitempty"`
	Screenshots   []string      `json:"screenshots,omitempty"`
	Timestamp     time.Time     `json:"timestamp"`
}

// CategoryGroup collects results grouped by test file or category.
type CategoryGroup struct {
	Category string       `json:"category"`
	Total    int          `json:"total"`
	Passed   int          `json:"passed"`
	Failed   int          `json:"failed"`
	Skipped  int          `json:"skipped"`
	Results  []TestResult `json:"results"`
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
	mu             sync.Mutex
	Summary        Summary         `json:"summary"`
	Results        []TestResult    `json:"results"`
	GroupedResults []CategoryGroup `json:"grouped_results"`
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
	r.RecordDetailed(name, testType, "General", "", "", "", status, duration, err, "", screenshot, nil)
}

// RecordWithCategory saves a single test case result along with its category/file name.
func (r *Reporter) RecordWithCategory(name string, testType string, category string, status string, duration time.Duration, err string, screenshot string) {
	r.RecordDetailed(name, testType, category, "", "", "", status, duration, err, "", screenshot, nil)
}

// RecordWithCategoryAndScreenshots saves a test result with a list of step screenshots.
func (r *Reporter) RecordWithCategoryAndScreenshots(name string, testType string, category string, status string, duration time.Duration, err string, screenshot string, screenshots []string) {
	r.RecordDetailed(name, testType, category, "", "", "", status, duration, err, "", screenshot, screenshots)
}

// RecordDetailed saves a test result with complete expected, actual, description, and failure reason details.
func (r *Reporter) RecordDetailed(name, testType, category, description, expected, actual, status string, duration time.Duration, errStr, failureReason, screenshot string, screenshots []string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if category == "" {
		category = "General"
	}
	if description == "" {
		description = name
	}

	result := TestResult{
		Name:          name,
		Type:          testType,
		Category:      category,
		Description:   description,
		Expected:      expected,
		Actual:        actual,
		Status:        status,
		Duration:      duration,
		DurationStr:   duration.Round(time.Millisecond).String(),
		Error:         errStr,
		FailureReason: failureReason,
		Screenshot:    screenshot,
		Screenshots:   screenshots,
		Timestamp:     time.Now(),
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

// GenerateReports compiles test stats and outputs JSON, HTML, and Markdown report files.
func (r *Reporter) GenerateReports(outputDir string) error {
	r.mu.Lock()
	r.Summary.EndTime = time.Now()
	r.Summary.TotalDuration = r.Summary.EndTime.Sub(r.Summary.StartTime)
	r.Summary.TotalDurationStr = r.Summary.TotalDuration.Round(time.Millisecond).String()
	r.mu.Unlock()

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}

	// Acquire a simple cross-process file lock
	lockPath := filepath.Join(outputDir, "report.lock")
	var lockFile *os.File
	var err error
	for i := 0; i < 50; i++ {
		lockFile, err = os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0666)
		if err == nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if err == nil {
		defer func() {
			lockFile.Close()
			os.Remove(lockPath)
		}()
	}

	jsonPath := filepath.Join(outputDir, "report.json")

	// Try to load and merge existing report
	var existingResults []TestResult
	var existingStartTime time.Time
	if data, err := os.ReadFile(jsonPath); err == nil {
		var existingReporter struct {
			Summary Summary      `json:"summary"`
			Results []TestResult `json:"results"`
		}
		if err := json.Unmarshal(data, &existingReporter); err == nil {
			existingResults = existingReporter.Results
			existingStartTime = existingReporter.Summary.StartTime
		}
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Merge results by name
	var mergedResults []TestResult
	seen := make(map[string]bool)

	// Helper to find index
	findResultIndex := func(results []TestResult, name string) int {
		for idx, res := range results {
			if res.Name == name {
				return idx
			}
		}
		return -1
	}

	// Add existing results, updated if they exist in the current run
	for _, res := range existingResults {
		if idx := findResultIndex(r.Results, res.Name); idx != -1 {
			mergedResults = append(mergedResults, r.Results[idx])
		} else {
			mergedResults = append(mergedResults, res)
		}
		seen[res.Name] = true
	}

	// Add remaining new results from the current run
	for _, res := range r.Results {
		if !seen[res.Name] {
			mergedResults = append(mergedResults, res)
			seen[res.Name] = true
		}
	}

	// Recalculate summary metrics
	var total, passed, failed, skipped int
	for _, res := range mergedResults {
		total++
		switch res.Status {
		case "passed":
			passed++
		case "failed":
			failed++
		case "skipped":
			skipped++
		}
	}

	r.Summary.Total = total
	r.Summary.Passed = passed
	r.Summary.Failed = failed
	r.Summary.Skipped = skipped
	if !existingStartTime.IsZero() && existingStartTime.Before(r.Summary.StartTime) {
		r.Summary.StartTime = existingStartTime
		r.Summary.TotalDuration = r.Summary.EndTime.Sub(r.Summary.StartTime)
		r.Summary.TotalDurationStr = r.Summary.TotalDuration.Round(time.Millisecond).String()
	}
	r.Results = mergedResults

	// Build GroupedResults by Category / File
	groupMap := make(map[string]*CategoryGroup)
	var groupOrder []string

	for _, res := range mergedResults {
		cat := res.Category
		if cat == "" {
			cat = "General / Uncategorized"
		}
		grp, exists := groupMap[cat]
		if !exists {
			grp = &CategoryGroup{
				Category: cat,
			}
			groupMap[cat] = grp
			groupOrder = append(groupOrder, cat)
		}
		grp.Results = append(grp.Results, res)
		grp.Total++
		switch res.Status {
		case "passed":
			grp.Passed++
		case "failed":
			grp.Failed++
		case "skipped":
			grp.Skipped++
		}
	}

	var groupedResults []CategoryGroup
	for _, cat := range groupOrder {
		groupedResults = append(groupedResults, *groupMap[cat])
	}
	r.GroupedResults = groupedResults

	// 1. JSON Report
	jsonData, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(jsonPath, jsonData, 0644); err != nil {
		return err
	}

	// 2. HTML Report
	htmlPath := filepath.Join(outputDir, "report.html")
	tmpl, err := template.New("report").Funcs(template.FuncMap{
		"basename": func(s string) string { return filepath.Base(s) },
	}).Parse(htmlTemplate)
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
