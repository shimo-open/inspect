package metric

import (
	"fmt"
	"strings"
	"time"

	"github.com/gotomicro/cetus/l"
	"github.com/gotomicro/ego/core/econf"
	"github.com/gotomicro/ego/core/elog"
	"github.com/jinzhu/now"
	_ "inspect/pkg/fetch"
	"inspect/pkg/invoker"
	"inspect/pkg/model"
	"inspect/pkg/mysql"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

// FetchDataJob 每个指标，每个服务抓取对应数据
// 从 metric 表里读取指标，依次查询指标前一天的数据，并聚合计算，最后落库存储
func (h *Handler) FetchDataJob() error {
	svcs := h.Apps()

	err := model.ForeachPlugin()
	if err != nil {
		return err
	}

	var metrics model.ReportMetrics
	if err := econf.UnmarshalKey("metrics", &metrics); err != nil {
		return fmt.Errorf("prometheus config error: %w", err)
	}
	startTime := now.BeginningOfDay().Add(time.Duration(24) * time.Hour)
	for _, m := range metrics {
		plugin, err := model.Provider(m.Type)
		if err != nil {
			return err
		}

		client, ok := plugin.GetClient(m.TypeName)
		if !ok {
			elog.Warn("prometheus config not exist: " + m.TypeName)
			continue
		}

		switch m.TargetType {
		case "svc":
			for _, svc := range svcs {
				svcName := svc
				if strings.HasSuffix(svc, "-headless") {
					continue
				}
				vals, err := client.QueryMetric(startTime, m, svcName)
				if err != nil {
					continue
				}
				h.Save(startTime, m, svcName, vals)
			}
		case "none":
			vals, err := client.QueryMetric(startTime, m, "none")
			if err != nil {
				continue
			}
			h.Save(startTime, m, "none", vals)
		default:
			elog.Error("not support target type", elog.Any("targetType", m.TargetType))
		}
	}
	return nil
}

func (*Handler) Apps() []string {
	return econf.GetStringSlice("apps")
}

func (*Handler) Save(startTime time.Time, m *model.ReportMetric, targetName string, vals []float64) error {
	if len(vals) == 0 {
		return nil
	}
	measure := mysql.ReportMeasure{
		Metric:     m.Name,
		TargetType: m.TargetType,
		TargetName: targetName,
		Val:        AggrFuncMap[m.AggrFunc](vals),
		Ctime:      time.Now().Unix(),
		Time:       startTime.Add(-24 * time.Hour).Unix(),
	}
	err := mysql.ReportMeasureCreateMulti(invoker.Db, []mysql.ReportMeasure{measure})
	if err != nil {
		elog.Error("save measure to db error", l.A("measure", measure), l.E(err))
		return err
	}
	elog.Info("save measure to db", l.A("measuer", measure))
	return nil
}
