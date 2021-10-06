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
	URL                   string                `yaml:"url"`
	Jobs                  map[string]MetricsJob `yaml:"jobs"`
	CadvisorJobs          map[string]MetricsJob `yaml:"cadvisorJobs"`
	EnableCadvisorMetrics bool                  `yaml:"enableCadvisorMetrics"`
}

type MetricsJob struct {
	Job        string `yaml:"job"`
	QueryLabel string `yaml:"queryLabel"`
}

func (m *Metrics) DistributorJob() MetricsJob {
	job, ok := m.Jobs["distributor"]
	if !ok {
		return MetricsJob{Job: "", QueryLabel: "job"}
	}

	return job
}

func (m *Metrics) CadvisorIngesterJob() MetricsJob {
	job, ok := m.CadvisorJobs["ingester"]
	if !ok {
		return MetricsJob{Job: "cadvisor_ingesters", QueryLabel: "job"}
	}

	return job
}

func (m *Metrics) CadvisorQuerierJob() MetricsJob {
	job, ok := m.CadvisorJobs["querier"]
	if !ok {
		return MetricsJob{Job: "cadvisor_querier", QueryLabel: "job"}
	}

	return job
}

func (m *Metrics) IngesterJob() MetricsJob {
	job, ok := m.Jobs["ingester"]
	if !ok {
		return MetricsJob{Job: "", QueryLabel: "job"}
	}

	return job
}

func (m *Metrics) QuerierJob() MetricsJob {
	job, ok := m.Jobs["querier"]
	if !ok {
		return MetricsJob{Job: "", QueryLabel: "job"}
	}

	return job
}

func (m *Metrics) QueryFrontendJob() MetricsJob {
	job, ok := m.Jobs["queryFrontend"]
	if !ok {
		return MetricsJob{Job: "", QueryLabel: "job"}
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
	Command  string            `yaml:"command"`
}

type Readers struct {
	Replicas       int32             `yaml:"replicas"`
	Args           map[string]string `yaml:"args"`
	Command        string            `yaml:"command"`
	StartThreshold float64           `yaml:"startThreshold"`
	Queries        map[string]string `yaml:"queries"`
}

type Samples struct {
	Interval time.Duration  `yaml:"interval"`
	Range    model.Duration `yaml:"range"`
	Total    int            `yaml:"total"`
}

type HighVolumeWrites struct {
	Enabled        bool            `yaml:"enabled"`
	Configurations []Configuration `yaml:"configurations"`
}

type HighVolumeReads struct {
	Enabled        bool            `yaml:"enabled"`
	Configurations []Configuration `yaml:"configurations"`
}

type Configuration struct {
	Description string   `yaml:"description"`
	Samples     Samples  `yaml:"samples"`
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
