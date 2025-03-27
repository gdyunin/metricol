// Package repository defines the interfaces and implementations for metric storage repositories.
// It provides methods for updating, retrieving, and checking the connection of metric data.
//
// The Repository interface specifies the basic operations for a metric repository, including Update,
// UpdateBatch, Find, All, and CheckConnection. This allows various implementations to be used
// interchangeably based on the application's needs.
//
// Implementations provided in this package include:
//
//   - InMemoryRepository:
//     A thread-safe, in-memory storage for metrics. It stores metrics in nested maps keyed
//     by metric type and name.
//
//   - InFileRepository:
//     A file-backed repository that extends InMemoryRepository by synchronizing metrics with a file on disk.
//     It supports auto-flushing to disk, data restoration on startup, and directory/file creation with retry logic.
//
//   - PostgreSQL:
//     A repository that persists metrics in a PostgreSQL database. It supports inserting/updating metrics,
//     batch operations via transactions, and automatic database migrations using embedded SQL files.
//     It also features connection checks with retry logic.
//
// These implementations provide flexible storage solutions for metrics in diverse environments.
package repository
