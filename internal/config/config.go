package config

import (
	"time"
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
	HighVolumeWrites HighVolumeWrites `yaml:"highVolumeWrites"`
	HighVolumeReads  HighVolumeReads  `yaml:"highVolumeReads"`
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

type Samples struct {
	Total    int           `yaml:"total"`
	Interval time.Duration `yaml:"interval"`
}

type Configuration struct {
	Description string   `yaml:"description"`
	Readers     *Readers `yaml:"readers,omitempty"`
	Writers     *Writers `yaml:"writers,omitempty"`
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
