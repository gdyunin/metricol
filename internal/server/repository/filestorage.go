package repository

import (
	"NewNewMetricol/internal/server/internal/entity"
	"NewNewMetricol/pkg/retry"
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"go.uber.org/zap"
)

const (
	FileDefaultPerm = 0o600
	DirDefaultPerm  = 0o750
)

type FileStorageRepository struct {
	InMemoryRepository
	ticker      *time.Ticker
	followChan  chan bool
	path        string
	interval    time.Duration
	needRestore bool
}

func NewFileStorageRepository(
	logger *zap.SugaredLogger,
	path string,
	filename string,
	interval time.Duration,
	restore bool,
) (*FileStorageRepository, error) {
	fsRepo := FileStorageRepository{
		InMemoryRepository: InMemoryRepository{
			counters: make(map[string]int64),
			gauges:   make(map[string]float64),
			mu:       &sync.RWMutex{},
			logger:   logger.Named("memory"),
		},
		path:        filepath.Join(path, filename),
		interval:    interval,
		needRestore: restore,
	}
	return fsRepo.build()
}

func (r *FileStorageRepository) Update(metric *entity.Metric) error {
	if err := r.InMemoryRepository.Update(metric); err != nil {
		return fmt.Errorf("failed save to memory: %w", err)
	}

	if r.interval == 0 {
		r.save()
	}
	return nil
}

func (r *FileStorageRepository) build() (*FileStorageRepository, error) {
	if err := retry.WithRetry(4, func() error {
		return os.MkdirAll(filepath.Dir(r.path), DirDefaultPerm)
	}); err != nil {
		return nil, err
	}

	if r.needRestore {
		r.mustRestore()
		r.needRestore = false
	}

	if r.interval != 0 {
		go func() {
			ticker := time.NewTicker(r.interval)
			defer ticker.Stop()

			for {
				<-ticker.C
				r.save()
			}
		}()
	}

	return r, nil
}

func (r *FileStorageRepository) Shutdown() {
	r.save()
}

func (r *FileStorageRepository) save() {
	metrics, err := r.All()
	if err != nil || metrics == nil {
		return
	}

	file, err := os.OpenFile(r.path, os.O_WRONLY|os.O_CREATE, FileDefaultPerm)
	if err != nil {
		return
	}
	defer func() { _ = file.Close() }()

	var buf bytes.Buffer
	for _, m := range *metrics {
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

func (r *FileStorageRepository) mustRestore() {
	file, err := os.Open(r.path)
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

		if err = r.InMemoryRepository.Update(&metric); err != nil {
			continue
		}
	}
}
