package prom

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/gotomicro/cetus/l"
	"github.com/gotomicro/ego/core/econf"
	"github.com/gotomicro/ego/core/elog"
	"github.com/influxdata/tdigest"
	"github.com/jinzhu/now"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"gorm.io/gorm"
	"inspect/pkg/mysql"
)

type Handler map[string]client

var Handle = make(Handler)

type Backend struct {
	Name         string `toml:"name"`
	Addr         string `toml:"addr"`
	AccessID     string `toml:"accessId"`
	AccessSecret string `toml:"accessSecret"`
}

type HandleConfig []Backend

type client struct {
	v1.API
}

type Err struct {
	svc string
	err error
}

func (e Err) Error() string {
	return fmt.Sprintf("svc:%s, err:%s", e.svc, e.err)
}

func Init() error {
	config := HandleConfig{}
	if err := econf.UnmarshalKey("prometheus", &config); err != nil {
		return fmt.Errorf("prometheus config error: %w", err)
	}
	for _, backend := range config {
		u, err := url.Parse(backend.Addr)
		if err != nil {
			return fmt.Errorf("prometheus config addr error: %w", err)
		}
		u.User = url.UserPassword(backend.AccessID, backend.AccessSecret)
		cli, err := api.NewClient(api.Config{
			Address: u.String(),
		})
		if err != nil {
			return fmt.Errorf("connect prometheus error:  %w", err)
		}
		Handle[backend.Name] = client{
			API: v1.NewAPI(cli),
		}
	}
	return nil
}

func Average(vals []float64) float64 {
	var sum float64
	for _, x := range vals {
		sum += x
	}
	return sum / float64(len(vals))
}

func Max(vars []float64) float64 {
	ret := vars[0]
	for _, x := range vars[1:] {
		if ret < x {
			ret = x
		}
	}
	return ret
}

func Quantile75(vals []float64) float64 {
	td := tdigest.NewWithCompression(1000)
	for _, x := range vals {
		td.Add(x, 1)
	}
	return td.Quantile(0.75)
}

func Quantile90(vals []float64) float64 {
	td := tdigest.NewWithCompression(1000)
	for _, x := range vals {
		td.Add(x, 1)
	}
	return td.Quantile(0.90)
}

func Quantile95(vals []float64) float64 {
	td := tdigest.NewWithCompression(1000)
	for _, x := range vals {
		td.Add(x, 1)
	}
	return td.Quantile(0.95)
}

func Quantile99(vals []float64) float64 {
	td := tdigest.NewWithCompression(1000)
	for _, x := range vals {
		td.Add(x, 1)
	}
	return td.Quantile(0.99)
}

// longRange 返回某日00:00:00到次日00:00:00时间区间
func longRange(yesterdayEnd time.Time) v1.Range {
	return v1.Range{
		Start: yesterdayEnd.Add(-24 * time.Hour),
		End:   yesterdayEnd,
		Step:  1 * time.Minute,
	}
}

func (p client) queryRange(ctx context.Context, q string, r v1.Range) (model.Matrix, error) {
	result, warnings, err := p.QueryRange(ctx, q, r)
	if err != nil {
		return nil, fmt.Errorf("Error querying Prometheus: %v\n", err)
	}
	if len(warnings) > 0 {
		return nil, fmt.Errorf("Warnings: %v\n", warnings)
	}
	return result.(model.Matrix), nil
}

func (p client) query(ctx context.Context, q string, t time.Time) (model.Vector, error) {
	result, warnings, err := p.Query(ctx, q, t)
	if err != nil {
		return nil, fmt.Errorf("Error querying Prometheus: %v\n", err)
	}
	if len(warnings) > 0 {
		return nil, fmt.Errorf("Warnings: %v\n", warnings)
	}
	return result.(model.Vector), nil
}

// ￥0.21：0.38/2G, 0.19/1G, 0.62/1U

type AggrFunc func(vals []float64) float64

var AggrFuncMap = map[string]AggrFunc{
	"max":        Max,
	"average":    Average,
	"quantile75": Quantile75,
	"quantile90": Quantile90,
	"quantile95": Quantile95,
	"quantile99": Quantile99,
}

// QueryAndSaveMeasure 从prometheus中查询一个指标存到数据库中
func (p client) QueryAndSaveMeasure(orm *gorm.DB, m *mysql.ReportMetric, targetName string) (val float64, err error) {
	takeCount := strings.Count(m.Query, "%s")
	taken := make([]interface{}, takeCount)
	for i := 0; i < takeCount; i++ {
		taken[i] = targetName
	}

	offset := 1
	startTime := now.BeginningOfDay().Add(time.Duration(24*offset) * time.Hour)

	elog.Info("start query: ", l.A("query", fmt.Sprintf(m.Query, taken...)), l.A("@prometheus", m.PrometheusName))
	res, err := p.queryRange(context.Background(), fmt.Sprintf(m.Query, taken...), longRange(startTime))

	if err != nil {
		elog.Error("query prometheus fail", elog.FieldErr(err))
		return 0, nil
	}

	if len(res) == 0 {
		elog.Warn("empty response")
		return 0, nil
	}

	vals := make([]float64, 0)
	for _, v := range res {
		for _, val := range v.Values {
			vals = append(vals, float64(val.Value))
		}
	}

	targetType := m.TargetType
	if targetType == "" {
		targetType = "none"
	}

	measure := mysql.ReportMeasure{
		Metric:     m.Name,
		TargetType: m.TargetType,
		TargetName: targetName,
		Val:        AggrFuncMap[m.AggrFunc](vals),
		Ctime:      time.Now().Unix(),
		Time:       startTime.Add(-24 * time.Hour).Unix(),
	}
	err = mysql.ReportMeasureCreateMulti(orm, []mysql.ReportMeasure{measure})
	if err != nil {
		elog.Error("save measure to db error", l.A("measure", measure), l.E(err))
	} else {
		elog.Info("save measure to db", l.A("measuer", measure))
	}
	return
}
