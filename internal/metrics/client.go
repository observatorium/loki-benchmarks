package metrics

import (
	"context"
	"fmt"
	"time"

	"github.com/onsi/gomega/gmeasure"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/config"
	"github.com/prometheus/common/model"
)

type Client struct {
	api     v1.API
	timeout time.Duration
}

func NewClient(url, token string, timeout time.Duration) (*Client, error) {
	httpConfig := config.HTTPClientConfig{
		TLSConfig: config.TLSConfig{
			InsecureSkipVerify: true,
		},
	}

	if token != "" {
		httpConfig.Authorization = &config.Authorization{
			Credentials: config.Secret(token),
		}
	}

	if err := httpConfig.Validate(); err != nil {
		return nil, fmt.Errorf("failed to validate httpConfig: %w", err)
	}

	rt, err := config.NewRoundTripperFromConfig(httpConfig, "benchmarks-metrics")
	if err != nil {
		return nil, fmt.Errorf("failed creating prometheus configuration: %w", err)
	}

	pc, err := api.NewClient(api.Config{
		Address:      url,
		RoundTripper: rt,
	})
	if err != nil {
		return nil, fmt.Errorf("failed creating prometheus client: %w", err)
	}

	return &Client{
		api:     v1.NewAPI(pc),
		timeout: timeout,
	}, nil
}

func (c *Client) Measure(e *gmeasure.Experiment, data Measurement) error {
	if e == nil {
		return fmt.Errorf("error measuring experiment: nil experiment")
	}

	value, err := c.executeScalarQuery(data.Query)
	if err != nil {
		return fmt.Errorf("error measuring experiment: %s", err)
	}

	e.RecordValue(data.Name, value, data.Unit, data.Annotation, gmeasure.Precision(4))
	return nil
}

func (c *Client) executeScalarQuery(query string) (float64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	res, _, err := c.api.Query(ctx, query, time.Now())
	if err != nil {
		return 0.0, fmt.Errorf("failed executing query %q: %w", query, err)
	}

	switch res.Type() {
	case model.ValScalar:
		value := res.(*model.Scalar)
		return float64(value.Value), nil
	case model.ValVector:
		vec := res.(model.Vector)
		if vec.Len() == 0 {
			return 0.0, nil
		}
		return float64(vec[0].Value), nil
	default:
		return 0.0, fmt.Errorf("failed to parse result for query: %s", query)
	}
}
