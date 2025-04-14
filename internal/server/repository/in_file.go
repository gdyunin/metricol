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
	// Const fileDefaultPerm defines the default file permissions.
	fileDefaultPerm = 0o600
	// Const dirDefaultPerm defines the default directory permissions.
	dirDefaultPerm = 0o750
	// Const makeDirTimeout specifies the timeout for directory creation.
	makeDirTimeout = 2 * time.Second
	// Const makeFileTimeout specifies the timeout for file creation.
	makeFileTimeout = 2 * time.Second
)

// InFileRepository represents a file-backed repository for metrics storage.
// It extends an in-memory repository by adding file synchronization capabilities.
type InFileRepository struct {
	*InMemoryRepository                    // Embedded in-memory repository.
	logger              *zap.SugaredLogger // Logger for repository operations.
	stopCh              chan struct{}      // Channel to signal stopping the auto-flush process.
	filepath            string             // Path of the storage file.
	autoFlushInterval   time.Duration      // Interval for automatically flushing data to the file.
	synchronized        bool               // Flag indicating whether the repository is in synchronized mode.
	restoreOnBuild      bool               // Flag indicating whether to restore data from file upon initialization.
}

// NewInFileRepository creates a new instance of InFileRepository.
// It initializes the underlying in-memory repository, sets up file path, and optionally restores data.
//
// Parameters:
//   - logger: Logger for repository operations.
//   - path: Directory path for the storage file.
//   - filename: Name of the storage file.
//   - interval: Auto-flush interval in seconds; if zero, the repository operates in synchronized mode.
//   - restore: Indicates if data should be restored from file during initialization.
//
// Returns:
//   - *InFileRepository: A pointer to the created InFileRepository instance.
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
// It first updates the in-memory repository and then flushes to file if in synchronized mode.
//
// Parameters:
//   - ctx: The context for the operation.
//   - metric: A pointer to the Metric to update.
//
// Returns:
//   - error: An error if the update fails.
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

// UpdateBatch adds or updates a batch of metrics in the repository.
//
// Parameters:
//   - ctx: The context for the operation.
//   - metrics: A pointer to the collection of Metrics to update.
//
// Returns:
//   - error: An error if any metric update fails.
func (r *InFileRepository) UpdateBatch(ctx context.Context, metrics *entity.Metrics) error {
	if metrics == nil {
		return errors.New("metrics should be non-nil, but got nil")
	}

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
// It retrieves all metrics, serializes them to JSON lines, and writes them to file.
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

// mustBuild initializes the repository by restoring data (if enabled),
// ensuring necessary directories and files exist, and starting auto-flush if required.
//
// Returns:
//   - *InFileRepository: A pointer to the fully initialized InFileRepository.
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

// shouldRestore restores metrics from the storage file into memory.
//
// Returns:
//   - error: An error if restoration fails.
func (r *InFileRepository) shouldRestore() error {
	if err := r.restore(); err != nil {
		return fmt.Errorf("failed to restore metrics: path=%s, error=%w", r.filepath, err)
	}
	return nil
}

// mustMakeDir ensures the directory for the storage file exists.
// It retries the directory creation and panics if it ultimately fails.
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
// It retries the file creation and panics if it ultimately fails.
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

// restore reads metrics from the storage file and loads them into the in-memory repository.
//
// Returns:
//   - error: An error if restoration fails.
func (r *InFileRepository) restore() error {
	file, err := os.Open(r.filepath)
	if err != nil {
		return fmt.Errorf("failed to open file for restoration: path=%s, error=%w", r.filepath, err)
	}
	defer func() {
		if err = file.Close(); err != nil {
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

// startAutoFlush starts a background process that periodically flushes metrics to the storage file.
// It continues until a stop signal is received via the stopCh channel.
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
