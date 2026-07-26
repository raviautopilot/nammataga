package service

import (
	"taga-api/config"
	"time"

	"go.uber.org/zap"
)

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
