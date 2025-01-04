package factory

import (
	"fmt"
	"time"

	"github.com/gdyunin/metricol.git/internal/server/backup"
	"github.com/gdyunin/metricol.git/internal/server/backup/managers/basic"
	"github.com/gdyunin/metricol.git/internal/server/entities"
)

const (
	ManagerTypeBasic = "basic"
)

func AbstractManagerFactory(managerType string, path string, filename string, interval time.Duration, restore bool, repo entities.MetricsRepository) (backup.ManagerAbstractFactory, error) {
	switch managerType {
	case ManagerTypeBasic:
		return basic.NewBackupManagerFactory(path, filename, interval, restore, repo), nil
	default:
		return nil, fmt.Errorf("unsupported manager type: '%s', please provide a valid manager type", managerType)
	}
}
