package repository

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gdyunin/metricol.git/internal/server/internal/entity"
	"github.com/gdyunin/metricol.git/pkg/retry"

	"go.uber.org/zap"
)

const (
	FileDefaultPerm = 0o600 // Default file permissions.
	DirDefaultPerm  = 0o750 // Default directory permissions.
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
func (r *InFileRepository) Update(metric *entity.Metric) error {
	// TODO: Добавить работу с контекстом для контроля времени выполнения. Добавить контекст в сигнатуру метода.
	if err := r.InMemoryRepository.Update(metric); err != nil {
		return fmt.Errorf(
			"failed to update metric in memory: type=%s, name=%s, value=%v, error: %w",
			metric.Type,
			metric.Name,
			metric.Value,
			err,
		)
	}

	if r.synchronized {
		r.flush()
	}
	return nil
}

// Shutdown gracefully stops the auto-flush process.
func (r *InFileRepository) Shutdown() {
	r.stopCh <- struct{}{}
}

// flush writes all metrics to the storage file.
func (r *InFileRepository) flush() {
	// TODO: Добавить работу с контекстом для контроля времени выполнения. Добавить контекст в сигнатуру метода.
	metrics, err := r.All()
	if err != nil || metrics == nil {
		r.logger.Warnf("failed to retrieve metrics for flushing: error=%v", err)
		return
	}

	file, err := os.OpenFile(r.filepath, os.O_WRONLY, FileDefaultPerm)
	if err != nil {
		r.logger.Errorf("unable to open file for writing: path=%s, error=%v", r.filepath, err)
		return
	}
	defer func() { _ = file.Close() }()

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
	if err := retry.WithRetry(AttemptsDefaultCount, func() error {
		return os.MkdirAll(filepath.Dir(r.filepath), DirDefaultPerm)
	}); err != nil {
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
	if err := retry.WithRetry(AttemptsDefaultCount, func() error {
		file, err := os.OpenFile(r.filepath, os.O_CREATE|os.O_EXCL, FileDefaultPerm)
		if err != nil {
			if os.IsExist(err) {
				return nil
			}
			return fmt.Errorf("failed to create file: path=%s, error=%w", r.filepath, err)
		}
		_ = file.Close()
		return nil
	}); err != nil {
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
	defer func() { _ = file.Close() }()

	reader := bufio.NewScanner(file)
	for reader.Scan() {
		data := reader.Bytes()

		metric := entity.Metric{}
		if err = json.Unmarshal(data, &metric); err != nil {
			r.logger.Warnf("failed to deserialize metric data: raw=%s, error=%v", string(data), err)
			continue
		}

		if err = r.InMemoryRepository.Update(&metric); err != nil {
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
			r.flush()
		case <-r.stopCh:
			r.flush()
			return
		}
	}
}
