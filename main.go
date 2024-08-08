package main

import (
	"github.com/gotomicro/ego"
	"github.com/gotomicro/ego/core/elog"
	"github.com/gotomicro/ego/task/ejob"
	"inspect/pkg/cron"
	"inspect/pkg/invoker"
	"inspect/pkg/job"
)

func main() {
	if err := ego.New().Invoker(
		invoker.Init,
		//metric.Init,
	).Job(ejob.Job("install", job.RunInstall)).Cron(cron.HandleCron()).Run(); err != nil {
		elog.Panic("Start up error: " + err.Error())
	}
}
