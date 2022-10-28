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

type MetricType string
type queryFunc func(job string, duration model.Duration) (float64, error)

const (
	SecondsPerDay              int    = 60 * 60 * 24
	BytesToGigabytesMultiplier string = "1000000000"
	BytesToMegabytesMultiplier string = "1000000"
)

type Client interface {
	DistributorBytesReceivedTotal(duration model.Duration) (float64, error)
	DistributorGiPDReceivedTotal(job string, duration model.Duration) (float64, error)
	DistributorGiPDDiscardedTotal(job string, duration model.Duration) (float64, error)

	// Load
	LoadNetworkTotal(job string, duration model.Duration) (float64, error)
	LoadNetworkGiPDTotal(job string, duration model.Duration) (float64, error)

	// HTTP API
	RequestDurationOkQueryRangeAvg(job string, duration model.Duration) (float64, error)
	RequestDurationOkQueryRangeP50(job string, duration model.Duration) (float64, error)
	RequestDurationOkQueryRangeP99(job string, duration model.Duration) (float64, error)

	RequestDurationOkPushAvg(job string, duration model.Duration) (float64, error)
	RequestDurationOkPushP50(job string, duration model.Duration) (float64, error)
	RequestDurationOkPushP99(job string, duration model.Duration) (float64, error)

	RequestReadsQPS(job string, duration model.Duration) (float64, error)
	RequestWritesQPS(job string, duration model.Duration) (float64, error)

	// GRPC API
	RequestDurationOkGrpcQuerySampleAvg(job string, duration model.Duration) (float64, error)
	RequestDurationOkGrpcQuerySampleP50(job string, duration model.Duration) (float64, error)
	RequestDurationOkGrpcQuerySampleP99(job string, duration model.Duration) (float64, error)

	RequestDurationOkGrpcPushAvg(job string, duration model.Duration) (float64, error)
	RequestDurationOkGrpcPushP50(job string, duration model.Duration) (float64, error)
	RequestDurationOkGrpcPushP99(job string, duration model.Duration) (float64, error)

	RequestReadsGrpcQPS(job string, duration model.Duration) (float64, error)
	RequestWritesGrpcQPS(job string, duration model.Duration) (float64, error)

	RequestQueryRangeThroughput(job string, duration model.Duration) (float64, error)

	// Store API
	RequestBoltDBShipperReadsQPS(job string, duration model.Duration) (float64, error)
	RequestBoltDBShipperWritesQPS(job string, duration model.Duration) (float64, error)

	// Container API
	// NOTE: Container API functions requires cadvisor to be deployed and functional
	ContainerCPU(job string, duration model.Duration) (float64, error)
	ContainerMemoryWorkingSetBytes(job string, duration model.Duration) (float64, error)

	// Process API
	ProcessCPUSeconds(job string, duration model.Duration) (float64, error)

	Measure(e *gmeasure.Experiment, f queryFunc, name, job string, defaultRange model.Duration) error
}

type client struct {
	api     v1.API
	timeout time.Duration
}

func NewClient(url, token string, timeout time.Duration) (Client, error) {
	httpConfig := config.HTTPClientConfig{
		BearerToken: config.Secret(token),
		TLSConfig: config.TLSConfig{
			InsecureSkipVerify: true,
		},
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

	return &client{
		api:     v1.NewAPI(pc),
		timeout: timeout,
	}, nil
}

func (c *client) Measure(e *gmeasure.Experiment, f queryFunc, name, job string, defaultRange model.Duration) error {
	measure, err := f(job, defaultRange)
	if err != nil {
		return fmt.Errorf("queryMetric %s failed job: %s err: %w", name, job, err)
	}

	e.RecordValue(fmt.Sprintf("%s - %s", job, name), measure)

	return nil
}

func (c *client) ProcessCPUSeconds(job string, duration model.Duration) (float64, error) {
	query := fmt.Sprintf(
		`sum(increase(process_cpu_seconds_total{pod=~".*%s.*"}[%s]))`,
		job, duration,
	)

	return c.executeScalarQuery(query)
}

func (c *client) ContainerCPU(job string, duration model.Duration) (float64, error) {
	query := fmt.Sprintf(
		`sum(avg_over_time(pod:container_cpu_usage:sum{pod=~".*%s.*"}[%s]))`,
		job, duration,
	)

	return c.executeScalarQuery(query)
}

func (c *client) ContainerMemoryWorkingSetBytes(job string, duration model.Duration) (float64, error) {
	query := fmt.Sprintf(
		`sum(avg_over_time(container_memory_working_set_bytes{pod=~".*%s.*", container=""}[%s]) / %s)`,
		job, duration, BytesToGigabytesMultiplier,
	)

	return c.executeScalarQuery(query)
}

func (c *client) requestRate(job, route, code string, duration model.Duration) (float64, error) {
	query := fmt.Sprintf(
		`sum(irate(loki_request_duration_seconds_count{job=~".*%s.*", route=~"%s", status_code=~"%s"}[%s]))`,
		job, route, code, duration,
	)

	return c.executeScalarQuery(query)
}

func (c *client) requestDurationAvg(job, method, route, code string, duration model.Duration) (float64, error) {
	query := fmt.Sprintf(
		`(sum(irate(loki_request_duration_seconds_sum{job=~".*%s.*", method="%s", route=~"%s", status_code=~"%s"}[%s])) / sum(irate(loki_request_duration_seconds_count{job=~".*%s.*", method="%s", route=~"%s", status_code=~"%s"}[%s])))`,
		job, method, route, code, duration,
		job, method, route, code, duration,
	)

	return c.executeScalarQuery(query)
}

func (c *client) requestDurationQuantile(job, method, route, code string, duration model.Duration, percentile int) (float64, error) {
	query := fmt.Sprintf(
		`histogram_quantile(0.%d, sum by (job, le) (irate(loki_request_duration_seconds_bucket{job=~".*%s.*", method="%s", route=~"%s", status_code=~"%s"}[%s])))`,
		percentile, job, method, route, code, duration,
	)

	return c.executeScalarQuery(query)
}

func (c *client) requestThroughput(job, endpoint, code, queryRange, metricType, latencyType, le string, duration model.Duration) (float64, error) {
	query := fmt.Sprintf(
		`(sum by (namespace, job) (irate(loki_logql_querystats_bytes_processed_per_seconds_bucket{status_code=~"%s", endpoint=~"%s", range="%s", type=~"%s", job=~".*%s.*", latency_type="%s", le="%s"}[%s])) / sum by (namespace, job) (irate(loki_logql_querystats_bytes_processed_per_seconds_count{status_code=~"%s", endpoint=~"%s", range="%s", type=~"%s", job=~".*%s.*"}[%s])))`,
		code, endpoint, queryRange, metricType, job, latencyType, le, duration,
		code, endpoint, queryRange, metricType, job, duration,
	)

	return c.executeScalarQuery(query)
}

func (c *client) requestBoltDBShipperQPS(job, operation, code string, duration model.Duration) (float64, error) {
	query := fmt.Sprintf(
		`sum(irate(loki_boltdb_shipper_request_duration_seconds_count{job=~".*%s.*", operation="%s", status_code=~"%s"}[%s]))`,
		job, operation, code, duration,
	)

	return c.executeScalarQuery(query)
}

func (c *client) executeScalarQuery(query string) (float64, error) {
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
