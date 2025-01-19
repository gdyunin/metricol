package repository

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gdyunin/metricol.git/internal/server/internal/entity"
	"github.com/gdyunin/metricol.git/pkg/retry"
	"github.com/labstack/gommon/log"

	"go.uber.org/zap"
)

const (
	fileDefaultPerm = 0o600           // Default file permissions.
	dirDefaultPerm  = 0o750           // Default directory permissions.
	makeDirTimeout  = 2 * time.Second // Timeout for make dir.
	makeFileTimeout = 2 * time.Second // Timeout for make new file.
)

// InFileRepository represents a file-backed repository for metrics storage.
// It extends InMemoryRepository with file synchronization capabilities.
type InFileRepository struct {
	*InMemoryRepository
	logger            *zap.SugaredLogger
	stopCh            chan struct{}
	filepath          string
	autoFlushInterval time.Duration
	synchronized      bool
	restoreOnBuild    bool
}

// NewInFileRepository creates a new instance of InFileRepository.
//
// Parameters:
//   - logger: Logger for repository operations.
//   - path: Directory path for the storage file.
//   - filename: Name of the storage file.
//   - interval: Auto-flush interval; 0 for synchronized mode.
//   - restore: Indicates if data should be restored from file during initialization.
//
// Returns:
//   - A pointer to the created InFileRepository instance.
func NewInFileRepository(
	logger *zap.SugaredLogger,
	path string,
	filename string,
	interval time.Duration,
	restore bool,
) *InFileRepository {
	ifr := InFileRepository{
		InMemoryRepository: NewInMemoryRepository(logger.Named("memory")),
		logger:             logger,
		synchronized:       interval == 0,
		stopCh:             make(chan struct{}),
		filepath:           filepath.Join(path, filename),
		restoreOnBuild:     restore,
		autoFlushInterval:  interval,
	}
	return ifr.mustBuild()
}

// Update adds or updates a metric in the repository.
//
// Parameters:
//   - metric: Pointer to the Metric to update.
//
// Returns:
//   - An error if the operation fails.
func (r *InFileRepository) Update(ctx context.Context, metric *entity.Metric) error {
	if err := r.InMemoryRepository.Update(ctx, metric); err != nil {
		return fmt.Errorf(
			"failed to update metric in memory: type=%s, name=%s, value=%v, error: %w",
			metric.Type,
			metric.Name,
			metric.Value,
			err,
		)
	}

	if r.synchronized {
		r.flush(ctx)
	}
	return nil
}

func (r *InFileRepository) UpdateBatch(ctx context.Context, metrics *entity.Metrics) error {
	if metrics == nil {
		return errors.New("metrics should be non-nil, but got nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	for _, m := range *metrics {
		if err := r.Update(ctx, m); err != nil {
			return fmt.Errorf("failed update one of metrics: %w", err)
		}
	}

	return nil
}

// Shutdown gracefully stops the auto-flush process.
func (r *InFileRepository) Shutdown() {
	r.stopCh <- struct{}{}
}

// flush writes all metrics to the storage file.
func (r *InFileRepository) flush(ctx context.Context) {
	metrics, err := r.All(ctx)
	if err != nil || metrics == nil {
		r.logger.Warnf("failed to retrieve metrics for flushing: error=%v", err)
		return
	}

	file, err := os.OpenFile(r.filepath, os.O_WRONLY, fileDefaultPerm)
	if err != nil {
		r.logger.Errorf("unable to open file for writing: path=%s, error=%v", r.filepath, err)
		return
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Errorf("File close error: %v", err)
		}
	}()

	var buf bytes.Buffer
	for _, m := range *metrics {
		data, err := json.Marshal(&m)
		if err != nil {
			r.logger.Warnf(
				"failed to serialize metric: type=%s, name=%s, value=%v, error: %v",
				m.Type,
				m.Name,
				m.Value,
				err,
			)
			continue
		}
		data = append(data, '\n')
		buf.Write(data)
	}

	writer := bufio.NewWriter(file)
	if _, err := writer.Write(buf.Bytes()); err != nil {
		r.logger.Errorf("failed to write metrics to file: path=%s, error=%v", r.filepath, err)
		return
	}

	if err := writer.Flush(); err != nil {
		r.logger.Errorf("failed to flush writer to file: path=%s, error=%v", r.filepath, err)
	}
}

// mustBuild initializes the repository, restoring data and starting auto-flush if necessary.
//
// Returns:
//   - A pointer to the initialized InFileRepository.
func (r *InFileRepository) mustBuild() *InFileRepository {
	if r.restoreOnBuild {
		if err := r.shouldRestore(); err != nil {
			r.logger.Warnf("Restore skipped with error: %v", err)
		}
	}

	r.mustMakeDir()
	r.mustMakeFile()
	if !r.synchronized {
		go r.startAutoFlush()
	}

	return r
}

// shouldRestore restores metrics from the storage file.
func (r *InFileRepository) shouldRestore() error {
	if err := r.restore(); err != nil {
		return fmt.Errorf("failed to restore metrics: path=%s, error=%w", r.filepath, err)
	}
	return nil
}

// mustMakeDir ensures the directory for the storage file exists.
// Panics if the directory cannot be created after retries.
func (r *InFileRepository) mustMakeDir() {
	ctx, cancel := context.WithTimeout(context.Background(), makeDirTimeout)
	defer cancel()

	if err := retry.WithRetry(
		ctx,
		r.logger,
		"making dirs for persistent storage",
		defaultAttemptsDefaultCount,
		func() error {
			return os.MkdirAll(filepath.Dir(r.filepath), dirDefaultPerm)
		},
	); err != nil {
		panic(fmt.Sprintf(
			"failed to create directory after retries: path=%s, error=%v",
			filepath.Dir(r.filepath),
			err,
		))
	}
}

// mustMakeFile ensures the storage file exists.
// Panics if the file cannot be created after retries.
func (r *InFileRepository) mustMakeFile() {
	ctx, cancel := context.WithTimeout(context.Background(), makeFileTimeout)
	defer cancel()

	if err := retry.WithRetry(
		ctx,
		r.logger,
		"create file for persistent storage",
		defaultAttemptsDefaultCount,
		func() error {
			file, err := os.OpenFile(r.filepath, os.O_CREATE|os.O_EXCL, fileDefaultPerm)
			if err != nil {
				if os.IsExist(err) {
					return nil
				}
				return fmt.Errorf("failed to create file: path=%s, error=%w", r.filepath, err)
			}
			_ = file.Close()
			return nil
		},
	); err != nil {
		panic(fmt.Sprintf("unable to create file after retries: path=%s, error=%v", r.filepath, err))
	}
}

// restore reads metrics from the storage file and loads them into memory.
//
// Returns:
//   - An error if the restoration fails.
func (r *InFileRepository) restore() error {
	file, err := os.Open(r.filepath)
	if err != nil {
		return fmt.Errorf("failed to open file for restoration: path=%s, error=%w", r.filepath, err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Errorf("File close error: %v", err)
		}
	}()

	reader := bufio.NewScanner(file)
	for reader.Scan() {
		data := reader.Bytes()

		metric := entity.Metric{}
		if err = json.Unmarshal(data, &metric); err != nil {
			r.logger.Warnf("failed to deserialize metric data: raw=%s, error=%v", string(data), err)
			continue
		}

		if err = r.InMemoryRepository.Update(context.TODO(), &metric); err != nil {
			r.logger.Warnf(
				"failed to load metric into memory: type=%s, name=%s, value=%v, error=%v",
				metric.Type,
				metric.Name,
				metric.Value,
				err,
			)
			continue
		}
	}

	return nil
}

// startAutoFlush starts a background process to periodically flush metrics to the storage file.
func (r *InFileRepository) startAutoFlush() {
	ticker := time.NewTicker(r.autoFlushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), r.autoFlushInterval)
			r.flush(ctx)
			cancel()
		case <-r.stopCh:
			r.flush(context.TODO())
			return
		}
	}
}
