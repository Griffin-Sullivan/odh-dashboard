package repositories

import (
	"log/slog"
)

// Repositories struct is a single convenient container to hold and represent all our repositories.
type Repositories struct {
	HealthCheck            *HealthCheckRepository
	MaaSModels             *MaaSModelsRepository
}

// NewRepositories creates domain-specific repositories.
func NewRepositories() *Repositories {
	return &Repositories{
		HealthCheck:            NewHealthCheckRepository(),
		MaaSModels:             NewMaaSModelsRepository(),
	}
}

// NewRepositoriesWithMCP creates repositories with MCP client factory
func NewRepositoriesWithMCP(logger *slog.Logger) *Repositories {
	repos := NewRepositories()
	return repos
}
