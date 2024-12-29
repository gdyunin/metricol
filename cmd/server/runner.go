package main

import (
	"log"
	"time"

	"github.com/gdyunin/metricol.git/internal/common"
	"github.com/gdyunin/metricol.git/internal/server/backup"
	"github.com/gdyunin/metricol.git/internal/server/consume"
	"github.com/gdyunin/metricol.git/internal/server/entities"
)

type runner struct {
	repo     entities.MetricRepository
	consumer consume.Consumer
	bm       backup.BackupManager
}

func newRunner(loggingLevel string, options ...func(*runner)) *runner {
	l := newLoggers(loggingLevel)
	appCfg := setupConfig(l.configParser)

	r := &runner{}
	r.repo = setupRepository(l.repository)
	r.consumer = setupConsumer(appCfg.ServerAddress, r.repo, l.consumer)
	r.bm = setupBackupManager(appCfg.FileStoragePath, time.Duration(appCfg.StoreInterval)*time.Second, appCfg.Restore, r.repo)

	return r.apply(options...)
}

func (r *runner) run() {
	r.bm.Restore()
	go r.bm.Start()

	sm := common.NewShutdownManager()
	sm.Add(r.bm.Stop)
	common.SetupGracefulShutdown(sm)

	if err := r.consumer.StartConsume(); err != nil {
		log.Fatal("")
	}
}

func (r *runner) apply(options ...func(*runner)) *runner {
	for _, o := range options {
		o(r)
	}
	return r
}
