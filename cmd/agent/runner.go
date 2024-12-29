package main

import (
	"log"
	"time"

	"github.com/gdyunin/metricol.git/internal/agent/collect"
	"github.com/gdyunin/metricol.git/internal/agent/entities"
	"github.com/gdyunin/metricol.git/internal/agent/orchestrate"
	"github.com/gdyunin/metricol.git/internal/agent/produce"
)

type runner struct {
	repo         entities.MetricsRepository
	collector    collect.Collector
	producer     produce.Producer
	orchestrator orchestrate.Orchestrator
}

func newRunner(loggingLevel string, options ...func(*runner)) *runner {
	l := newLoggers(loggingLevel)
	appCfg := setupConfig(l.configParser)

	r := &runner{}
	r.repo = setupRepository(l.repository)
	r.collector = setupCollector(time.Duration(appCfg.PollInterval)*time.Second, r.repo, l.collector)
	r.producer = setupProducer(time.Duration(appCfg.ReportInterval)*time.Second, appCfg.ServerAddress, r.repo, l.producer)
	r.orchestrator = setupOrchestrator(r.collector, r.producer, l.orchestrator)

	return r.apply(options...)
}

func (r *runner) run() {
	if err := r.orchestrator.StartAll(); err != nil {
		log.Fatal("")
	}
}

func (r *runner) apply(options ...func(*runner)) *runner {
	for _, o := range options {
		o(r)
	}
	return r
}
