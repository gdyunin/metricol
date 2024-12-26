package basebackup

import (
	"bufio"
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/gdyunin/metricol.git/internal/common"
	"github.com/gdyunin/metricol.git/internal/server/entity"
)

const (
	FileDefaultPerm = 0o600
	DirDefaultPerm  = 0o750
)

type BaseBackupper struct {
	repo        entity.MetricRepository
	ticker      *time.Ticker
	followChan  chan bool
	path        string
	interval    time.Duration
	needRestore bool
}

func NewBaseBackupper(
	path string,
	filename string,
	interval time.Duration,
	restore bool,
	repo entity.MetricRepository,
) *BaseBackupper {
	return &BaseBackupper{
		path:        filepath.Join(path, filename),
		interval:    interval,
		repo:        repo,
		needRestore: restore,
	}
}

func (b *BaseBackupper) StartBackup() {
	if b.interval == 0 {
		b.syncBackup()
	} else {
		b.regularBackup()
	}
}

func (b *BaseBackupper) StopBackup() {
	if b.followChan != nil {
		close(b.followChan)
	}
	b.ticker.Stop()
	b.backup()
}

func (b *BaseBackupper) OnNotify() {
	if b.followChan == nil {
		b.followChan = make(chan bool, 1)
	}
	b.followChan <- true
}

func (b *BaseBackupper) syncBackup() {
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

func (b *BaseBackupper) regularBackup() {
	b.ticker = time.NewTicker(b.interval)
	defer b.ticker.Stop()

	for range b.ticker.C {
		b.backup()
	}
}

func (b *BaseBackupper) backup() {
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

func (b *BaseBackupper) Restore() {
	if b.needRestore {
		b.mustRestore()
	}
}

func (b *BaseBackupper) mustRestore() {
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

		metric := entity.Metric{}
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
