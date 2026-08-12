package service

import (
	"taga-api/config"
	"taga-api/service/audit"
	"time"

	"go.uber.org/zap"
)

// StartAuditCleanupScheduler runs audit log retention cleanup once per month.
// It honours the AUDIT_LOG_RETENTION_MONTHS env var (default 3).
func StartAuditCleanupScheduler() {
	config.Logger.Info("Starting audit log cleanup scheduler")
	go func() {
		// Run once on startup so retention is enforced immediately after deployment.
		audit.RunCleanup()

		for {
			now := time.Now()
			// Schedule next run at the 1st of the next month at 02:00.
			next := time.Date(now.Year(), now.Month()+1, 1, 2, 0, 0, 0, now.Location())
			time.Sleep(time.Until(next))
			audit.RunCleanup()
		}
	}()
}

// StartScheduler starts a daily job that checks for renewal reminders.
func StartScheduler() {
	config.Logger.Info("Starting renewal reminder scheduler")
	go func() {
		for {
			now := time.Now()
			// Schedule next run at 8:00 AM tomorrow
			next := time.Date(now.Year(), now.Month(), now.Day()+1, 8, 0, 0, 0, now.Location())
			duration := next.Sub(now)
			time.Sleep(duration)

			// Run the check
			if err := SendRemindersIfDue(); err != nil {
				config.Logger.Error("Scheduler error", zap.Error(err))
			}
		}
	}()
}
