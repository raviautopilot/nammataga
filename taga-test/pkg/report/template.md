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

## Test Details

| Test Case / Description | Suite / Type | Status | Duration |
| :--- | :--- | :--- | :--- |
{{range .Results}}| {{.Name}} | {{.Type}} | {{if eq .Status "passed"}}🟢 Passed{{else if eq .Status "failed"}}🔴 Failed{{else}}🟡 Skipped{{end}} | {{.DurationStr}} |
{{end}}

{{if gt .Summary.Failed 0}}
## Failures & Errors

{{range .Results}}{{if eq .Status "failed"}}
### ❌ {{.Name}}

- **Type**: {{.Type}}
- **Duration**: {{.DurationStr}}
- **Error Trace**:
```
{{.Error}}
```
{{if .Screenshot}}
- **Screenshot**:
  ![Screenshot]({{.Screenshot}})
{{end}}

---
{{end}}{{end}}{{end}}
