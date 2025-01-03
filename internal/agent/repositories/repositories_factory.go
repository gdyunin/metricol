package repositories

import (
	"fmt"

	"github.com/gdyunin/metricol.git/internal/agent/entities"
	"go.uber.org/zap"
)

const (
	RepoTypeInMemory = "in memory repository"
)

func AbstractRepositoriesFactory(repoType string, logger *zap.SugaredLogger) (entities.RepositoryAbstractFactory, error) {
	switch repoType {
	case RepoTypeInMemory:
		return NewInMemoryRepositoryFactory(logger), nil
	default:
		return nil, fmt.Errorf("unsupported repository type: %s", repoType)
	}
}
