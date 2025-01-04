package basic

import (
	"bufio"
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/gdyunin/metricol.git/internal/common/patterns"
	"github.com/gdyunin/metricol.git/internal/server/backup"
	"github.com/gdyunin/metricol.git/internal/server/entities"
)

const (
	// FileDefaultPerm defines default permissions for backup files.
	FileDefaultPerm = 0o600
	// DirDefaultPerm defines default permissions for backup directories.
	DirDefaultPerm = 0o750
)

// BackupManagerFactory creates instances of BackupManager with specified parameters.
type BackupManagerFactory struct {
	path     string
	filename string
	interval time.Duration
	restore  bool
	repo     entities.MetricsRepository
}

// NewBackupManagerFactory initializes and returns a new BackupManagerFactory.
//
// Parameters:
//   - path: Directory path for backup files.
//   - filename: Name of the backup file.
//   - interval: Interval for regular backups.
//   - restore: Whether restoration from backups is enabled.
//   - repo: Metrics repository to manage.
//
// Returns:
//   - A pointer to a BackupManagerFactory instance.
func NewBackupManagerFactory(path, filename string, interval time.Duration, restore bool, repo entities.MetricsRepository) *BackupManagerFactory {
	return &BackupManagerFactory{
		path:     path,
		filename: filename,
		interval: interval,
		restore:  restore,
		repo:     repo,
	}
}

// CreateManager creates and returns a new BackupManager instance.
func (f *BackupManagerFactory) CreateManager() backup.Manager {
	return NewBackupManager(f.path, f.filename, f.interval, f.restore, f.repo)
}

// BackupManager manages backup and restore operations for metrics.
type BackupManager struct {
	repo        entities.MetricsRepository
	ticker      *time.Ticker
	followChan  chan bool
	path        string
	interval    time.Duration
	needRestore bool
}

// NewBackupManager initializes and returns a new BackupManager.
//
// Parameters:
//   - path: Directory path for backup files.
//   - filename: Name of the backup file.
//   - interval: Interval for regular backups.
//   - restore: Whether restoration from backups is enabled.
//   - repo: Metrics repository to manage.
//
// Returns:
//   - A pointer to a BackupManager instance.
func NewBackupManager(path, filename string, interval time.Duration, restore bool, repo entities.MetricsRepository) *BackupManager {
	return &BackupManager{
		path:        filepath.Join(path, filename),
		interval:    interval,
		repo:        repo,
		needRestore: restore,
	}
}

// Start begins the backup process. If an interval is set, regular backups occur; otherwise, backups are event-driven.
func (b *BackupManager) Start() {
	if b.interval == 0 {
		b.syncBackup()
	} else {
		b.regularBackup()
	}
}

// Stop halts the backup process and performs a final backup.
func (b *BackupManager) Stop() {
	if b.followChan != nil {
		close(b.followChan)
	}
	if b.ticker != nil {
		b.ticker.Stop()
	}
	b.backup()
}

// OnNotify triggers an immediate backup in response to an observed event.
func (b *BackupManager) OnNotify() {
	if b.followChan == nil {
		b.followChan = make(chan bool, 1)
	}
	b.followChan <- true
}

// Restore restores metrics from a backup if restoration is enabled.
func (b *BackupManager) Restore() {
	if b.needRestore {
		b.mustRestore()
	}
}

// syncBackup manages backups triggered by observed changes.
func (b *BackupManager) syncBackup() {
	sbj, ok := b.repo.(patterns.ObserveSubject)
	if !ok {
		return
	}

	if err := sbj.RegisterObserver(b); err != nil {
		return
	}

	for range b.followChan {
		b.backup()
	}
}

// regularBackup performs backups at regular intervals.
func (b *BackupManager) regularBackup() {
	b.ticker = time.NewTicker(b.interval)
	defer b.ticker.Stop()

	for range b.ticker.C {
		b.backup()
	}
}

// backup saves all metrics to the configured backup file.
func (b *BackupManager) backup() {
	metrics, err := b.repo.All()
	if err != nil {
		return
	}

	err = os.MkdirAll(filepath.Dir(b.path), DirDefaultPerm)
	if err != nil {
		return
	}

	file, err := os.OpenFile(b.path, os.O_WRONLY|os.O_CREATE, FileDefaultPerm)
	if err != nil {
		return
	}
	defer func() { _ = file.Close() }()

	var buf bytes.Buffer
	for _, m := range metrics {
		data, err := json.Marshal(&m)
		if err != nil {
			continue
		}
		data = append(data, '\n')
		buf.Write(data)
	}

	writer := bufio.NewWriter(file)
	if _, err = writer.Write(buf.Bytes()); err != nil {
		return
	}

	_ = writer.Flush()
}

// mustRestore restores metrics from the backup file.
func (b *BackupManager) mustRestore() {
	file, err := os.Open(b.path)
	if err != nil {
		return
	}
	defer func() { _ = file.Close() }()

	reader := bufio.NewScanner(file)
	for {
		if !reader.Scan() {
			return
		}
		data := reader.Bytes()

		metric := entities.Metric{}
		if err = json.Unmarshal(data, &metric); err != nil {
			continue
		}

		if err := metric.AfterJSONUnmarshalling(); err != nil {
			continue
		}

		_ = b.repo.Update(&metric)
	}
}
