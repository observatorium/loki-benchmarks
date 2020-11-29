package config

import (
	"fmt"
	"time"

	"github.com/prometheus/common/model"
)

type Logger struct {
	Name      string `yaml:"name"`
	Namespace string `yaml:"namespace"`
	Image     string `yaml:"image"`
	TenantID  string `yaml:"tenantId"`
}

type Querier struct {
	Name      string `yaml:"name"`
	Namespace string `yaml:"namespace"`
	Image     string `yaml:"image"`
	TenantID  string `yaml:"tenantId"`
}

type Metrics struct {
	URL  string            `yaml:"url"`
	Jobs map[string]string `yaml:"jobs"`
}

func (m *Metrics) DistributorJob() string {
	job, ok := m.Jobs["distributor"]
	if !ok {
		return ""
	}

	return job
}

func (m *Metrics) IngesterJob() string {
	job, ok := m.Jobs["ingester"]
	if !ok {
		return ""
	}

	return job
}

func (m *Metrics) QuerierJob() string {
	job, ok := m.Jobs["querier"]
	if !ok {
		return ""
	}

	return job
}

func (m *Metrics) QueryFrontendJob() string {
	job, ok := m.Jobs["queryFrontend"]
	if !ok {
		return ""
	}

	return job
}

type Loki struct {
	Distributor   string `yaml:"distributor"`
	QueryFrontend string `yaml:"queryFrontend"`
}

func (lc *Loki) PushURL() string {
	return fmt.Sprintf("%s/loki/api/v1/push", lc.Distributor)
}

func (lc *Loki) QueryURL() string {
	return fmt.Sprintf("%s/loki/api/v1/query", lc.QueryFrontend)
}

func (lc *Loki) QueryRangeURL() string {
	return fmt.Sprintf("%s/loki/api/v1/query_range", lc.QueryFrontend)
}

type Writers struct {
	Replicas   int32 `yaml:"replicas"`
	Throughput int32 `yaml:"throughput"`
}

type Readers struct {
	Replicas       int32             `yaml:"replicas"`
	Queries        map[string]string `yaml:"queries"`
	StartThreshold float64           `yaml:"startThreshold"`
}

type Samples struct {
	Interval time.Duration  `yaml:"interval"`
	Range    model.Duration `yaml:"range"`
	Total    int            `yaml:"total"`
}

type HighVolumeWrites struct {
	Enabled bool     `yaml:"enabled"`
	Samples Samples  `yaml:"samples"`
	Readers *Readers `yaml:"readers,omitempty"`
	Writers *Writers `yaml:"writers,omitempty"`
}

type HighVolumeReads struct {
	Enabled bool     `yaml:"enabled"`
	Samples Samples  `yaml:"samples"`
	Readers *Readers `yaml:"readers,omitempty"`
	Writers *Writers `yaml:"writers,omitempty"`
}

type HighVolumeAggregate struct {
	Enabled bool     `yaml:"enabled"`
	Samples Samples  `yaml:"samples"`
	Readers *Readers `yaml:"readers,omitempty"`
	Writers *Writers `yaml:"writers,omitempty"`
}

type LogsBasedDashboard struct {
	Enabled bool     `yaml:"enabled"`
	Samples Samples  `yaml:"samples"`
	Readers *Readers `yaml:"readers,omitempty"`
	Writers *Writers `yaml:"writers,omitempty"`
}

type Scenarios struct {
	HighVolumeWrites    HighVolumeWrites    `yaml:"highVolumeWrites"`
	HighVolumeReads     HighVolumeReads     `yaml:"highVolumeReads"`
	HighVolumeAggregate HighVolumeAggregate `yaml:"highVolumeAggregate"`
	LogsBasedDashboard  LogsBasedDashboard  `yaml:"logsBasedDashboard"`
}

type Benchmark struct {
	Logger    *Logger    `yaml:"logger"`
	Querier   *Querier   `yaml:"querier"`
	Metrics   *Metrics   `yaml:"metrics"`
	Loki      *Loki      `yaml:"loki"`
	Scenarios *Scenarios `yaml:"scenarios"`
}
