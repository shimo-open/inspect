package fetch

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/gotomicro/cetus/l"
	"github.com/gotomicro/ego/core/econf"
	"github.com/gotomicro/ego/core/elog"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	modelmetric "inspect/pkg/model"
)

type Backend struct {
	Name         string `toml:"name"`
	Addr         string `toml:"addr"`
	AccessID     string `toml:"accessId"`
	AccessSecret string `toml:"accessSecret"`
}

type HandleConfig []Backend

type Client struct {
	v1.API
}

type PrometheusPlugin struct {
	clients map[string]Client
}

func init() {
	modelmetric.RegisterPlugin(&PrometheusPlugin{
		clients: make(map[string]Client),
	})
}

func (p *PrometheusPlugin) Init() error {
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
		p.clients[backend.Name] = Client{
			API: v1.NewAPI(cli),
		}
	}
	return nil
}

func (*PrometheusPlugin) Scheme() string {
	return "prometheus"
}

func (p *PrometheusPlugin) GetClient(clientName string) (client modelmetric.Fetch, flag bool) {
	client, flag = p.clients[clientName]
	return
}

// longRange 返回某日00:00:00到次日00:00:00时间区间
func longRange(yesterdayEnd time.Time) v1.Range {
	return v1.Range{
		Start: yesterdayEnd.Add(-24 * time.Hour),
		End:   yesterdayEnd,
		Step:  1 * time.Minute,
	}
}

func (p Client) queryRange(ctx context.Context, q string, r v1.Range) (model.Matrix, error) {
	result, warnings, err := p.QueryRange(ctx, q, r)
	if err != nil {
		return nil, fmt.Errorf("Error querying Prometheus: %v\n", err)
	}
	if len(warnings) > 0 {
		return nil, fmt.Errorf("Warnings: %v\n", warnings)
	}
	return result.(model.Matrix), nil
}

func (p Client) query(ctx context.Context, q string, t time.Time) (model.Vector, error) {
	result, warnings, err := p.Query(ctx, q, t)
	if err != nil {
		return nil, fmt.Errorf("Error querying Prometheus: %v\n", err)
	}
	if len(warnings) > 0 {
		return nil, fmt.Errorf("Warnings: %v\n", warnings)
	}
	return result.(model.Vector), nil
}

// QueryMetric 从prometheus中查询一个指标存到数据库中
func (p Client) QueryMetric(startTime time.Time, m *modelmetric.ReportMetric, targetName string) (vals []float64, err error) {
	takeCount := strings.Count(m.Query, "%s")
	taken := make([]interface{}, takeCount)
	for i := 0; i < takeCount; i++ {
		taken[i] = targetName
	}
	elog.Info("start query: ", l.A("query", fmt.Sprintf(m.Query, taken...)), l.A("@prometheus", m.TypeName))
	res, err := p.queryRange(context.Background(), fmt.Sprintf(m.Query, taken...), longRange(startTime))
	if err != nil {
		elog.Error("query prometheus fail", elog.FieldErr(err))
		return nil, nil
	}

	if len(res) == 0 {
		elog.Warn("empty response")
		return nil, nil
	}

	vals = make([]float64, 0)
	for _, v := range res {
		for _, val := range v.Values {
			vals = append(vals, float64(val.Value))
		}
	}

	return
}
