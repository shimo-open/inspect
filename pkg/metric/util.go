package metric

import (
	"github.com/influxdata/tdigest"
)

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
