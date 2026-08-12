package audit

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"go.uber.org/zap"

	"taga-api/config"
)

// RunCleanup archives audit-log year/month directories that are older than the
// configured retention period (AUDIT_LOG_RETENTION_MONTHS env var, default 3).
//
// Safety rules:
//   - Never touches the current month.
//   - Never touches future months.
//   - Never touches directories outside "audit-logs/".
//   - Handles missing directories gracefully.
//   - Compresses expired months to .tar.gz and deletes original folders.
//   - Logs every compression and every error via the application logger.
func RunCleanup() {
	retention := retentionMonths()
	// Cutoff: the first moment of the month that is `retention` months ago.
	now := time.Now()
	cutoff := time.Date(now.Year(), now.Month()-time.Month(retention), 1, 0, 0, 0, 0, now.Location())

	if config.Logger != nil {
		config.Logger.Info("Audit log cleanup started",
			zap.Int("retention_months", retention),
			zap.String("cutoff", cutoff.Format("2006-01")),
		)
	}

	yearDirs, err := filepath.Glob("audit-logs/????")
	if err != nil || len(yearDirs) == 0 {
		if config.Logger != nil {
			config.Logger.Info("Audit log cleanup: no audit directories found")
		}
		return
	}

	for _, yearDir := range yearDirs {
		monthDirs, _ := filepath.Glob(filepath.Join(yearDir, "??"))
		for _, monthDir := range monthDirs {
			year := filepath.Base(yearDir)
			month := filepath.Base(monthDir)

			dirTime, err := time.Parse("2006/01", fmt.Sprintf("%s/%s", year, month))
			if err != nil {
				if config.Logger != nil {
					config.Logger.Warn("Audit cleanup: cannot parse directory date",
						zap.String("dir", monthDir))
				}
				continue
			}

			// Keep current and future months
			if !dirTime.Before(cutoff) {
				continue
			}

			archiveName := filepath.Join(yearDir, fmt.Sprintf("%s.tar.gz", month))

			// Check if already compressed (should not happen as monthDirs is matching '??')
			if _, err := os.Stat(archiveName); err == nil {
				// Already exists, just remove original dir
				_ = os.RemoveAll(monthDir)
				continue
			}

			if err := compressDir(monthDir, archiveName); err != nil {
				if config.Logger != nil {
					config.Logger.Error("Audit cleanup failed to compress directory",
						zap.String("dir", monthDir),
						zap.Error(err),
					)
				}
				continue
			}

			if err := os.RemoveAll(monthDir); err != nil {
				if config.Logger != nil {
					config.Logger.Error("Audit cleanup failed to delete directory after compression",
						zap.String("dir", monthDir),
						zap.Error(err),
					)
				}
			} else {
				if config.Logger != nil {
					config.Logger.Info("Audit cleanup compressed and removed expired directory",
						zap.String("dir", monthDir),
						zap.String("archive", archiveName),
					)
				}
			}
		}

		// Remove the year directory if it is now empty
		entries, _ := os.ReadDir(yearDir)
		if len(entries) == 0 {
			if err := os.Remove(yearDir); err != nil {
				if config.Logger != nil {
					config.Logger.Warn("Audit cleanup: failed to remove empty year dir",
						zap.String("dir", yearDir), zap.Error(err))
				}
			}
		}
	}

	if config.Logger != nil {
		config.Logger.Info("Audit log cleanup finished")
	}
}

// retentionMonths returns the configured retention period in months.
// Reads AUDIT_LOG_RETENTION_MONTHS from the environment; falls back to 3.
func retentionMonths() int {
	val := os.Getenv("AUDIT_LOG_RETENTION_MONTHS")
	if n, err := strconv.Atoi(val); err == nil && n > 0 {
		return n
	}
	return 3
}

// compressDir compresses a directory into a .tar.gz archive.
func compressDir(srcDir, destArchive string) error {
	out, err := os.Create(destArchive)
	if err != nil {
		return err
	}
	defer out.Close()

	gw := gzip.NewWriter(out)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	return filepath.Walk(srcDir, func(file string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(fi, fi.Name())
		if err != nil {
			return err
		}

		// Make the path relative to the srcDir parent
		relPath, err := filepath.Rel(filepath.Dir(srcDir), file)
		if err != nil {
			return err
		}
		header.Name = relPath

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		if !fi.Mode().IsRegular() {
			return nil
		}

		f, err := os.Open(file)
		if err != nil {
			return err
		}
		defer f.Close()

		if _, err := io.Copy(tw, f); err != nil {
			return err
		}

		return nil
	})
}
