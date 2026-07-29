# E2E Test Execution Report

## Summary

| Metric | Value |
| :--- | :--- |
| **Total Tests** | {{.Summary.Total}} |
| **Passed** | {{.Summary.Passed}} |
| **Failed** | {{.Summary.Failed}} |
| **Skipped** | {{.Summary.Skipped}} |
| **Start Time** | {{.Summary.StartTime.Format "2006-01-02 15:04:05"}} |
| **End Time** | {{.Summary.EndTime.Format "2006-01-02 15:04:05"}} |
| **Total Duration** | {{.Summary.TotalDurationStr}} |

---

## Detailed Test Results Categorized by Test File

{{range .GroupedResults}}
### 📄 `{{.Category}}` (Passed: {{.Passed}}, Failed: {{.Failed}}, Total: {{.Total}})

| Test Case Name & Purpose | Expected Result | Actual Result (Got) | Status | Duration |
| :--- | :--- | :--- | :--- | :--- |
{{range .Results}}| **{{.Name}}**<br>_{{.Description}}_ | `{{if .Expected}}{{.Expected}}{{else}}200 OK Response{{end}}` | `{{if .Actual}}{{.Actual}}{{else}}{{.Status}}{{end}}` | {{if eq .Status "passed"}}🟢 Passed{{else if eq .Status "failed"}}🔴 Failed{{else}}🟡 Skipped{{end}} | {{.DurationStr}} |
{{end}}

{{end}}

{{if gt .Summary.Failed 0}}
---

## Failed Tests & Diagnoses

{{range .GroupedResults}}{{range .Results}}{{if eq .Status "failed"}}
### ❌ [{{.Category}}] {{.Name}}

- **Description / Purpose**: {{.Description}}
- **Expected Result**: `{{.Expected}}`
- **Actual Result (Got)**: `{{.Actual}}`
- **Duration**: {{.DurationStr}}
- **Failure Reason**:
```
{{if .FailureReason}}{{.FailureReason}}{{else}}{{.Error}}{{end}}
```
{{if .Screenshot}}
- **Failure Screenshot**:
  ![Screenshot]({{.Screenshot}})
{{end}}

---
{{end}}{{end}}{{end}}
{{end}}
