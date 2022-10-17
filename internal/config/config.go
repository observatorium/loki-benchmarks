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
	URL                   string            `yaml:"url"`
	Jobs                  map[string]string `yaml:"jobs"`
	CadvisorJobs          map[string]string `yaml:"cadvisorJobs"`
	EnableCadvisorMetrics bool              `yaml:"enableCadvisorMetrics"`
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

type Writers struct {
	Replicas int32             `yaml:"replicas"`
	Args     map[string]string `yaml:"args"`
}

type Readers struct {
	Replicas int32             `yaml:"replicas"`
	Args     map[string]string `yaml:"args"`
	Queries  map[string]string `yaml:"queries"`
}

type Samples struct {
	Interval time.Duration  `yaml:"interval"`
	Range    model.Duration `yaml:"range"`
	Total    int            `yaml:"total"`
}

type HighVolumeWrites struct {
	Enabled        bool            `yaml:"enabled"`
	Samples        Samples         `yaml:"samples"`
	Configurations []Configuration `yaml:"configurations"`
}

type HighVolumeReads struct {
	Enabled        bool            `yaml:"enabled"`
	Generator      *Writers        `yaml:"generator"`
	StartThreshold float64         `yaml:"startThreshold"`
	Samples        Samples         `yaml:"samples"`
	Configurations []Configuration `yaml:"configurations"`
}

type Configuration struct {
	Description string   `yaml:"description"`
	Readers     *Readers `yaml:"readers,omitempty"`
	Writers     *Writers `yaml:"writers,omitempty"`
}

type Scenarios struct {
	HighVolumeWrites HighVolumeWrites `yaml:"highVolumeWrites"`
	HighVolumeReads  HighVolumeReads  `yaml:"highVolumeReads"`
}

type Benchmark struct {
	Logger    *Logger    `yaml:"logger"`
	Querier   *Querier   `yaml:"querier"`
	Metrics   *Metrics   `yaml:"metrics"`
	Loki      *Loki      `yaml:"loki"`
	Scenarios *Scenarios `yaml:"scenarios"`
}
