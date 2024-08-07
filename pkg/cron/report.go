package cron

import (
	"context"

	"github.com/gotomicro/ego/task/ecron"
	"inspect/pkg/prom"
)

func PromCron() *ecron.Component {
	job := func(ctx context.Context) error {
		return prom.Handle.FetchDataJob()
	}
	cron := ecron.Load("cron.prometheus").Build(ecron.WithJob(job))
	return cron
}
