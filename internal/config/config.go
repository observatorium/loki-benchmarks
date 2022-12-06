package config

import (
	"time"

	"github.com/onsi/gomega/gmeasure"
)

type Benchmark struct {
	Generator *Generator `yaml:"generator"`
	Querier   *Querier   `yaml:"querier"`
	Metrics   *Metrics   `yaml:"metrics"`
	Scenarios *Scenarios `yaml:"scenarios"`
}

type Generator struct {
	Namespace      string `yaml:"namespace"`
	ServiceAccount string `yaml:"serviceAccount,omitempty"`
	Image          string `yaml:"image"`
	Tenant         string `yaml:"tenant"`
	PushURL        string `yaml:"pushURL"`
}

type Querier struct {
	Namespace      string `yaml:"namespace"`
	ServiceAccount string `yaml:"serviceAccount,omitempty"`
	Image          string `yaml:"image"`
	Tenant         string `yaml:"tenant"`
	PullURL        string `yaml:"pullURL"`
}

type Metrics struct {
	URL                   string `yaml:"url"`
	Jobs                  *Jobs  `yaml:"jobs"`
	EnableCadvisorMetrics bool   `yaml:"enableCadvisorMetrics"`
}

type Jobs struct {
	Distributor   string `yaml:"distributor"`
	Ingester      string `yaml:"ingester"`
	Querier       string `yaml:"querier"`
	QueryFrontend string `yaml:"queryFrontend"`
}

type Scenarios struct {
	HighVolumeWrites *HighVolumeWrites `yaml:"highVolumeWrites,omitempty"`
	HighVolumeReads  *HighVolumeReads  `yaml:"highVolumeReads,omitempty"`
}

type HighVolumeWrites struct {
	Enabled     bool    `yaml:"enabled"`
	Description string  `yaml:"description"`
	Samples     Samples `yaml:"samples"`
	Writers     Writers `yaml:"writers"`
}

type HighVolumeReads struct {
	Enabled        bool    `yaml:"enabled"`
	Description    string  `yaml:"description"`
	Samples        Samples `yaml:"samples"`
	StartThreshold float64 `yaml:"startThreshold"`
	Generator      Writers `yaml:"generator"`
	Readers        Readers `yaml:"readers"`
}

type Samples struct {
	Total    int           `yaml:"total"`
	Interval time.Duration `yaml:"interval"`
}

func (s Samples) SamplingConfiguration() gmeasure.SamplingConfig {
	return gmeasure.SamplingConfig{
		N:                   s.Total,
		Duration:            s.Interval * time.Duration(s.Total+1),
		MinSamplingInterval: s.Interval,
	}
}

type Writers struct {
	Replicas int32             `yaml:"replicas"`
	Args     map[string]string `yaml:"args"`
}

type Readers struct {
	Replicas   int32             `yaml:"replicas"`
	Queries    map[string]string `yaml:"queries"`
	QueryRange string            `yaml:"queryRange"`
}
