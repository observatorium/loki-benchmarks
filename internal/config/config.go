package config

import (
	"time"

	"github.com/onsi/gomega/gmeasure"
	"github.com/prometheus/common/model"
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
	Writers     Writer  `yaml:"writers"`
	samples     *Sample `yaml:"samples,omitempty"`
}

func (w *HighVolumeWrites) SamplingConfiguration() (gmeasure.SamplingConfig, model.Duration) {
	samples := w.samples
	if samples == nil {
		samples = &Sample{
			Total:    10,
			Interval: time.Minute * 3,
		}
	}

	return gmeasure.SamplingConfig{
		N:                   samples.Total,
		Duration:            samples.Interval * time.Duration(samples.Total+1),
		MinSamplingInterval: samples.Interval,
	}, model.Duration(samples.Interval)
}

type HighVolumeReads struct {
	Enabled        bool    `yaml:"enabled"`
	Description    string  `yaml:"description"`
	StartThreshold float64 `yaml:"startThreshold"`
	Readers        Reader  `yaml:"readers"`
	samples        *Sample `yaml:"samples,omitempty"`
	generator      *Writer `yaml:"generator,omitempty"`
}

func (r *HighVolumeReads) SamplingConfiguration() (gmeasure.SamplingConfig, model.Duration) {
	samples := r.samples
	if samples == nil {
		samples = &Sample{
			Total:    10,
			Interval: time.Minute,
		}
	}

	return gmeasure.SamplingConfig{
		N:                   samples.Total,
		Duration:            samples.Interval * time.Duration(samples.Total+1),
		MinSamplingInterval: samples.Interval,
	}, model.Duration(samples.Interval)
}

func (r *HighVolumeReads) LogGenerator() Writer {
	if r.generator == nil {
		return Writer{
			Replicas: 10,
			Args: map[string]string{
				"source":         "application",
				"log-lines-rate": "500",
			},
		}
	}
	return *r.generator
}

type Sample struct {
	Total    int           `yaml:"total"`
	Interval time.Duration `yaml:"interval"`
}

type Writer struct {
	Replicas int32             `yaml:"replicas"`
	Args     map[string]string `yaml:"args"`
}

type Reader struct {
	Replicas   int32             `yaml:"replicas"`
	Queries    map[string]string `yaml:"queries"`
	QueryRange string            `yaml:"queryRange"`
}
