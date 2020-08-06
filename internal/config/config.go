package config

import "fmt"

type Logger struct {
    Name       string `yaml:"name"`
    Namespace  string `yaml:"namespace"`
    Image      string `yaml:"image"`
    TenantID   string `yaml:"tenantId"`
    Replicas   int32  `yaml:"replicas"`
    Throughput int32  `yaml:"throughput"`
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

type Loki struct {
    URL string `yaml:"url"`
}

func (lc *Loki) PushURL() string {
    return fmt.Sprintf("%s/push", lc.URL)
}

func (lc *Loki) QueryURL() string {
    return fmt.Sprintf("%s/query", lc.URL)
}

type Benchmark struct {
    Logger  *Logger  `yaml:"logger"`
    Metrics *Metrics `yaml:"metrics"`
    Loki    *Loki    `yaml:"loki"`
}
