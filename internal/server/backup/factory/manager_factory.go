package factory

import (
	"fmt"
	"time"

	"github.com/gdyunin/metricol.git/internal/server/backup"
	"github.com/gdyunin/metricol.git/internal/server/backup/managers/basic"
	"github.com/gdyunin/metricol.git/internal/server/entities"
)

const (
	// ManagerTypeBasic represents the identifier for the basic backup manager type.
	ManagerTypeBasic = "basic"
)

// AbstractManagerFactory creates a backup manager factory based on the specified manager type.
//
// Parameters:
//   - managerType: A string representing the type of backup manager to create (e.g., "basic").
//   - path: The directory path where backups will be stored.
//   - filename: The name of the backup file.
//   - interval: The interval duration between backups.
//   - restore: A boolean flag indicating whether to restore data from an existing backup during initialization.
//   - repo: The metrics repository to be managed by the backup manager.
//
// Returns:
//   - An implementation of the backup.ManagerAbstractFactory interface for the specified manager type.
//   - An error if the manager type is unsupported.
func AbstractManagerFactory(
	managerType string,
	path string,
	filename string,
	interval time.Duration,
	restore bool,
	repo entities.MetricsRepository,
) (backup.ManagerAbstractFactory, error) {
	switch managerType {
	case ManagerTypeBasic:
		return basic.NewBackupManagerFactory(path, filename, interval, restore, repo), nil
	default:
		// Return an error if the manager type is unsupported.
		return nil, fmt.Errorf(
			"unsupported manager type: '%s', please provide a valid manager type",
			managerType,
		)
	}
}
