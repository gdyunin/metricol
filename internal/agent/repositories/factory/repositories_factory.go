package factory

import (
	"fmt"

	"github.com/gdyunin/metricol.git/internal/agent/entities"
	"github.com/gdyunin/metricol.git/internal/agent/repositories"
	"go.uber.org/zap"
)

const (
	// RepoTypeInMemory specifies the type identifier for an in-memory repository implementation.
	RepoTypeInMemory = "in memory repository"
)

// AbstractRepositoriesFactory creates a repository factory based on the specified repository type.
//
// Parameters:
//   - repoType: The type of repository to create (e.g., "in memory repository").
//   - logger: A logger instance for logging repository-related activities.
//
// Returns:
//   - An implementation of entities.RepositoryAbstractFactory for the specified repository type.
//   - An error if the repository type is unsupported.
func AbstractRepositoriesFactory(repoType string, logger *zap.SugaredLogger) (entities.RepositoryAbstractFactory, error) {
	switch repoType {
	case RepoTypeInMemory:
		// Create and return an in-memory repository factory.
		return repositories.NewInMemoryRepositoryFactory(logger), nil
	default:
		// Return an error if the repository type is unsupported.
		return nil, fmt.Errorf("unsupported repository type: '%s', please provide a valid repository type", repoType)
	}
}
