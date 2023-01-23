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

func (s *Scenarios) IsWriteTestRunnable() bool {
	if s == nil {
		return false
	}

	if s.HighVolumeWrites == nil {
		return false
	}

	return s.HighVolumeWrites.Enabled
}

func (s *Scenarios) IsReadTestRunnable() bool {
	if s == nil {
		return false
	}

	if s.HighVolumeReads == nil {
		return false
	}

	return s.HighVolumeReads.Enabled
}

type HighVolumeWrites struct {
	Enabled     bool    `yaml:"enabled"`
	Description string  `yaml:"description"`
	Writers     *Writer `yaml:"writers"`
	samples     *Sample `yaml:"samples,omitempty"`
}

func (w *HighVolumeWrites) SamplingConfiguration() (gmeasure.SamplingConfig, model.Duration) {
	samples := &Sample{
		Total:    10,
		Interval: time.Minute * 3,
	}

	if w != nil {
		if w.samples != nil {
			samples = w.samples
		}
	}

	return gmeasure.SamplingConfig{
		N:                   samples.Total,
		Duration:            samples.Interval * time.Duration(samples.Total+1),
		MinSamplingInterval: samples.Interval,
	}, model.Duration(samples.Interval)
}

type HighVolumeReads struct {
	Enabled     bool    `yaml:"enabled"`
	Description string  `yaml:"description"`
	Readers     *Reader `yaml:"readers"`
	samples     *Sample `yaml:"samples,omitempty"`
	generator   *Writer `yaml:"generator,omitempty"`
}

func (r *HighVolumeReads) SamplingConfiguration() (gmeasure.SamplingConfig, model.Duration) {
	samples := &Sample{
		Total:    15,
		Interval: time.Minute,
	}

	if r != nil {
		if r.samples != nil {
			samples = r.samples
		}
	}

	return gmeasure.SamplingConfig{
		N:                   samples.Total,
		Duration:            samples.Interval * time.Duration(samples.Total+1),
		MinSamplingInterval: samples.Interval,
	}, model.Duration(samples.Interval)
}

func (r *HighVolumeReads) LogGenerator() *Writer {
	writer := &Writer{
		Replicas: 15,
		Args: map[string]string{
			"log-type":        "application",
			"logs-per-second": "500",
		},
	}

	if r != nil {
		if r.samples != nil {
			writer = r.generator
		}
	}

	return writer
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
