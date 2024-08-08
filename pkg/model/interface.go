package model

import (
	"fmt"
	"time"
)

type Plugin interface {
	Init() (err error)
	Scheme() string
	GetClient(string) (Fetch, bool)
}

type Fetch interface {
	QueryMetric(startTime time.Time, m *ReportMetric, targetName string) (vals []float64, err error)
}

var (
	registry = make(map[string]Plugin)
)

// RegisterPlugin registers a dataSource creator function to the registry
func RegisterPlugin(builder Plugin) {
	registry[builder.Scheme()] = builder
}

func ForeachPlugin() error {
	for _, plugin := range registry {
		err := plugin.Init()
		if err != nil {
			return fmt.Errorf(plugin.Scheme()+" init fail, err: %w", err)
		}
	}
	return nil
}

// Provider 根据协议找到Plugin
func Provider(scheme string) (Plugin, error) {
	logger, ok := registry[scheme]
	if !ok {
		return nil, fmt.Errorf("not found " + scheme)
	}
	return logger, nil
}
