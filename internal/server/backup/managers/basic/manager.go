package basic

import (
	"bufio"
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/gdyunin/metricol.git/internal/common"
	"github.com/gdyunin/metricol.git/internal/server/entities"
)

const (
	FileDefaultPerm = 0o600
	DirDefaultPerm  = 0o750
)

type BackupManager struct {
	repo        entities.MetricRepository
	ticker      *time.Ticker
	followChan  chan bool
	path        string
	interval    time.Duration
	needRestore bool
}

func NewBackupManager(
	path string,
	filename string,
	interval time.Duration,
	restore bool,
	repo entities.MetricRepository,
) *BackupManager {
	return &BackupManager{
		path:        filepath.Join(path, filename),
		interval:    interval,
		repo:        repo,
		needRestore: restore,
	}
}

func (b *BackupManager) StartBackup() {
	if b.interval == 0 {
		b.syncBackup()
	} else {
		b.regularBackup()
	}
}

func (b *BackupManager) StopBackup() {
	if b.followChan != nil {
		close(b.followChan)
	}
	b.ticker.Stop()
	b.backup()
}

func (b *BackupManager) OnNotify() {
	if b.followChan == nil {
		b.followChan = make(chan bool, 1)
	}
	b.followChan <- true
}

func (b *BackupManager) syncBackup() {
	sbj, ok := b.repo.(common.ObserveSubject)
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

func (b *BackupManager) regularBackup() {
	b.ticker = time.NewTicker(b.interval)
	defer b.ticker.Stop()

	for range b.ticker.C {
		b.backup()
	}
}

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
	_, err = writer.Write(buf.Bytes())
	if err != nil {
		return
	}

	if err := writer.Flush(); err != nil {
		return
	}
}

func (b *BackupManager) Restore() {
	if b.needRestore {
		b.mustRestore()
	}
}

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

		if err = b.repo.Update(&metric); err != nil {
			continue
		}
	}
}
