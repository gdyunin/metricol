package main

import (
	"NewNewMetricol/internal/server/delivery"
	"NewNewMetricol/internal/server/repository"

	"go.uber.org/zap"
)

func main() {
	es := delivery.NewEchoServer("localhost:8080", repository.NewInMemoryRepository(zap.NewExample().Sugar()), zap.NewExample().Sugar())
	es.Start()
}
