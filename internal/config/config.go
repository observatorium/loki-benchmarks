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
	IndexGateway  string `yaml:"indexGateway"`
}

type Scenarios struct {
	IngestionPath *IngestionPath `yaml:"ingestionPath,omitempty"`
	QueryPath     *QueryPath     `yaml:"queryPath,omitempty"`
}

func (s *Scenarios) IsWriteTestEnabled() bool {
	if s == nil {
		return false
	}

	if s.IngestionPath == nil {
		return false
	}

	return s.IngestionPath.Enabled
}

func (s *Scenarios) IsReadTestEnabled() bool {
	if s == nil {
		return false
	}

	if s.QueryPath == nil {
		return false
	}

	return s.QueryPath.Enabled
}

type IngestionPath struct {
	Enabled     bool    `yaml:"enabled"`
	Description string  `yaml:"description"`
	Writers     *Writer `yaml:"writers"`
	samples     *Sample `yaml:"samples,omitempty"`
}

func (w *IngestionPath) SamplingConfiguration() (gmeasure.SamplingConfig, model.Duration) {
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

type QueryPath struct {
	Enabled     bool    `yaml:"enabled"`
	Description string  `yaml:"description"`
	Readers     *Reader `yaml:"readers"`
	samples     *Sample `yaml:"samples,omitempty"`
	generator   *Writer `yaml:"generator,omitempty"`
}

func (r *QueryPath) SamplingConfiguration() (gmeasure.SamplingConfig, model.Duration) {
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

func (r *QueryPath) LogGenerator() *Writer {
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
