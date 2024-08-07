package mysql

// ReportMetric 读取查询语句
type ReportMetric struct {
	Name           string
	Desc           string
	Type           int    // 指标类型
	Query          string // 查询语句
	TargetType     string // 目标类型
	AggrFunc       string // max, sum, average, quantile90, quantile95, quantile99
	PrometheusName string // prometheus名称
}

type ReportMetrics []*ReportMetric