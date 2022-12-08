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

type RequestPath int

const (
	WriteRequestPath RequestPath = 0
	ReadRequestPath  RequestPath = 1
)

type Client struct {
	api               v1.API
	timeout           time.Duration
	isCAdvisorEnabled bool
}

func NewClient(url, token string, timeout time.Duration, cadvisorEnabled bool) (*Client, error) {
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
		api:               v1.NewAPI(pc),
		timeout:           timeout,
		isCAdvisorEnabled: cadvisorEnabled,
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

func (c *Client) MeasureHTTPRequestMetrics(
	e *gmeasure.Experiment,
	path RequestPath,
	job string,
	sampleRange model.Duration,
	annotation gmeasure.Annotation,
) error {
	switch path {
	case WriteRequestPath:
		if err := c.Measure(e, RequestWritesQPS(job, sampleRange, annotation)); err != nil {
			return err
		}
		return c.measureCommonRequestMetrics(e, job, HTTPPostMethod, HTTPPushRoute, sampleRange, annotation)
	case ReadRequestPath:
		if err := c.Measure(e, RequestQueryRangeThroughput(job, sampleRange, annotation)); err != nil {
			return err
		}
		if err := c.Measure(e, RequestReadsQPS(job, sampleRange, annotation)); err != nil {
			return err
		}
		return c.measureCommonRequestMetrics(e, job, HTTPGetMethod, HTTPQueryRangeRoute, sampleRange, annotation)
	default:
		return fmt.Errorf("error unknown path specified: %d", path)
	}
}

func (c *Client) MeasureGRPCRequestMetrics(
	e *gmeasure.Experiment,
	path RequestPath,
	job string,
	sampleRange model.Duration,
	annotation gmeasure.Annotation,
) error {
	switch path {
	case WriteRequestPath:
		if err := c.Measure(e, RequestWritesGrpcQPS(job, sampleRange, annotation)); err != nil {
			return err
		}
		return c.measureCommonRequestMetrics(e, job, GRPCMethod, GRPCPushRoute, sampleRange, annotation)
	case ReadRequestPath:
		if err := c.Measure(e, RequestReadsGrpcQPS(job, sampleRange, annotation)); err != nil {
			return err
		}
		return c.measureCommonRequestMetrics(e, job, GRPCMethod, GRPCQuerySampleRoute, sampleRange, annotation)
	default:
		return fmt.Errorf("error unknown path specified: %d", path)
	}
}

func (c *Client) MeasureResourceUsageMetrics(
	e *gmeasure.Experiment,
	job string,
	sampleRange model.Duration,
	annotation gmeasure.Annotation,
) error {
	if err := c.Measure(e, ContainerCPU(job, sampleRange, annotation)); err != nil {
		return err
	}
	if err := c.Measure(e, PersistentVolumeUsedBytes(job, sampleRange, annotation)); err != nil {
		return err
	}

	if c.isCAdvisorEnabled {
		if err := c.Measure(e, ContainerMemoryWorkingSetBytes(job, sampleRange, annotation)); err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) MeasureIngestionVerificationMetrics(
	e *gmeasure.Experiment,
	deployment, distributor string,
	sampleRange model.Duration,
) error {
	if err := c.Measure(e, LoadNetworkTotal(deployment, sampleRange)); err != nil {
		return err
	}
	if err := c.Measure(e, LoadNetworkGiPDTotal(deployment, sampleRange)); err != nil {
		return err
	}
	if err := c.Measure(e, DistributorGiPDReceivedTotal(distributor, sampleRange)); err != nil {
		return err
	}
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

func (c *Client) measureCommonRequestMetrics(
	e *gmeasure.Experiment,
	job, method, route string,
	sampleRange model.Duration,
	annotation gmeasure.Annotation,
) error {
	name := fmt.Sprintf("2xx %s", route)
	code := "2.*"
	if method == GRPCMethod {
		name = fmt.Sprintf("successful GRPC %s", route)
		code = "success"
	}

	if err := c.Measure(e, RequestDurationAverage(name, job, method, route, code, sampleRange, annotation)); err != nil {
		return err
	}
	if err := c.Measure(e, RequestDurationQuantile(name, job, method, route, code, 90, sampleRange, annotation)); err != nil {
		return err
	}
	return nil
}
