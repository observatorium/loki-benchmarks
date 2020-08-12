package config

import (
    "fmt"
    "time"
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
    return fmt.Sprintf("%s/loki/api/v1/query_range", lc.QueryFrontend)
}

type Writers struct {
    Replicas   int32 `yaml:"replicas"`
    Throughput int32 `yaml:"throughput"`
}

type Readers struct {
    Replicas       int32   `yaml:"replicas"`
    Query          string  `yaml:"query"`
    StartThreshold float64 `yaml:"startThreshold"`
}

type Samples struct {
    Interval time.Duration `yaml:"interval"`
}

type HighVolumeWrites struct {
    P99 float64 `yaml:"p99"`
    P50 float64 `yaml:"p50"`
    AVG float64 `yaml:"avg"`

    Writers *Writers `yaml:"writers,omitempty"`
    Readers *Readers `yaml:"readers,omitempty"`

    Samples Samples `yaml:"samples"`
}

type HighVolumeReads struct {
    P99 float64 `yaml:"p99"`
    P50 float64 `yaml:"p50"`
    AVG float64 `yaml:"avg"`

    Writers *Writers `yaml:"writers,omitempty"`
    Readers *Readers `yaml:"readers,omitempty"`

    Samples Samples `yaml:"samples"`
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
