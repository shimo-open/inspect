package cron

import (
	"context"

	"github.com/gotomicro/ego/task/ecron"
	"inspect/pkg/metric"
)

func HandleCron() *ecron.Component {
	job := func(ctx context.Context) error {
		handler := metric.NewHandler()
		return handler.FetchDataJob()
	}
	cron := ecron.Load("cron.prometheus").Build(ecron.WithJob(job))
	return cron
}
