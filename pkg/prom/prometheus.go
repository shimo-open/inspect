package prom

import (
	"fmt"
	"strings"

	"github.com/gotomicro/ego/core/econf"
	"github.com/gotomicro/ego/core/elog"
	"inspect/pkg/invoker"
	"inspect/pkg/mysql"
)

// FetchDataJob 每个指标，每个服务抓取对应数据
// 从 metric 表里读取指标，依次查询指标前一天的数据，并聚合计算，最后落库存储
func (ph Handler) FetchDataJob() error {
	svcs := ph.Apps()
	var metrics mysql.ReportMetrics
	if err := econf.UnmarshalKey("metrics", &metrics); err != nil {
		return fmt.Errorf("prometheus config error: %w", err)
	}
	for _, m := range metrics {
		if _, ok := ph[m.PrometheusName]; !ok {
			elog.Warn("prometheus config not exist: " + m.PrometheusName)
			continue
		}

		if m.TargetType == "" || m.TargetType == "none" {
			ph[m.PrometheusName].QueryAndSaveMeasure(invoker.Db, m, "none")
		} else {
			for _, svc := range svcs {
				svc := svc
				if strings.HasSuffix(svc, "-headless") {
					continue
				}
				ph[m.PrometheusName].QueryAndSaveMeasure(invoker.Db, m, svc)
			}
		}
	}
	return nil
}

func (ph Handler) Apps() []string {
	return econf.GetStringSlice("apps")
}
