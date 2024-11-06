package metrics

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ViaQ/logerr/v2/kverrors"
	"github.com/onsi/gomega/gmeasure"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/config"
	"github.com/prometheus/common/model"
)

type RequestPath int

const (
	WriteRequestPath RequestPath = iota
	ReadRequestPath
)

var (
	errNilExperiment = errors.New("error measuring experiment: nil experiment")
	errUnknownPath   = errors.New("error unknown path specified")
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
		return nil, kverrors.Wrap(err, "failed to validate httpConfig")
	}

	rt, err := config.NewRoundTripperFromConfig(httpConfig, "benchmarks-metrics")
	if err != nil {
		return nil, kverrors.Wrap(err, "failed creating prometheus configuration")
	}

	pc, err := api.NewClient(api.Config{
		Address:      url,
		RoundTripper: rt,
	})
	if err != nil {
		return nil, kverrors.Wrap(err, "failed creating prometheus client")
	}

	return &Client{
		api:               v1.NewAPI(pc),
		timeout:           timeout,
		isCAdvisorEnabled: cadvisorEnabled,
	}, nil
}

func (c *Client) Measure(e *gmeasure.Experiment, data Measurement) error {
	if e == nil {
		return errNilExperiment
	}

	value, err := c.executeScalarQuery(data.Query)
	if err != nil {
		return kverrors.Wrap(err, "error measuring experiment")
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
		return c.measureCommonRequestMetrics(e, job, HTTPPostMethod, HTTPPushRoute, HTTPPushRoute, sampleRange, annotation)
	case ReadRequestPath:
		return c.measureCommonRequestMetrics(e, job, HTTPGetMethod, HTTPQueryRangeRoute, HTTPReadPathRoutes, sampleRange, annotation)
	default:
		return kverrors.Wrap(errUnknownPath, "path", path)
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
		return c.measureCommonRequestMetrics(e, job, GRPCMethod, GRPCPushRoute, GRPCPushRoute, sampleRange, annotation)
	case ReadRequestPath:
		return c.measureCommonRequestMetrics(e, job, GRPCMethod, GRPCQuerySampleRoute, GRPCReadPathRoutes, sampleRange, annotation)
	default:
		return kverrors.Wrap(errUnknownPath, "path", path)
	}
}

func (c *Client) MeasureIndexRequestMetrics(
	e *gmeasure.Experiment,
	path RequestPath,
	job string,
	sampleRange model.Duration,
) error {
	switch path {
	case WriteRequestPath:
		return c.Measure(e, RequestIndexRequestRate(IndexWriteName, job, WriteOperation, "2.*", sampleRange))
	case ReadRequestPath:
		return c.Measure(e, RequestIndexRequestRate(IndexReadName, job, ReadOperation, "2.*", sampleRange))
	default:
		return kverrors.Wrap(errUnknownPath, "path", path)
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

	if c.isCAdvisorEnabled {
		if err := c.Measure(e, ContainerMemoryWorkingSetBytes(job, sampleRange, annotation)); err != nil {
			return err
		}

		if err := c.Measure(e, ContainerGoMemstatsHeapInuse(job, sampleRange, annotation)); err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) MeasureVolumeUsageMetrics(
	e *gmeasure.Experiment,
	job string,
	sampleRange model.Duration,
	annotation gmeasure.Annotation,
) error {
	if err := c.Measure(e, PersistentVolumeUsedBytes(job, sampleRange, annotation)); err != nil {
		return err
	}
	return nil
}

func (c *Client) MeasureQueryMetrics(
	e *gmeasure.Experiment,
	job string,
	sampleRange model.Duration,
	annotation gmeasure.Annotation,
) error {
	if err := c.Measure(e, LogQLQueryLatencyAverage("2.*", job, sampleRange, annotation)); err != nil {
		return err
	}
	if err := c.Measure(e, LogQLQueryLatencyQuantile("2.*", job, DefaultPercentile, sampleRange, annotation)); err != nil {
		return err
	}
	if err := c.Measure(e, LogQLQueryMBpSProcessedAverage("2.*", job, sampleRange, annotation)); err != nil {
		return err
	}
	if err := c.Measure(e, LogQLQueryMBpSProcessedQuantile("2.*", job, DefaultPercentile, sampleRange, annotation)); err != nil {
		return err
	}

	return nil
}

func (c *Client) MeasureIngestionVerificationMetrics(
	e *gmeasure.Experiment,
	deployment string,
	sampleRange model.Duration,
) error {
	if err := c.Measure(e, LoadNetworkTotal(deployment, sampleRange)); err != nil {
		return err
	}
	if err := c.Measure(e, LoadNetworkGiPDTotal(deployment, sampleRange)); err != nil {
		return err
	}
	if err := c.Measure(e, DistributorGiPDReceivedTotal(sampleRange)); err != nil {
		return err
	}
	if err := c.Measure(e, LokiStreamsInMemoryTotal(sampleRange)); err != nil {
		return err
	}
	return nil
}

func (c *Client) MeasureLoadQuerierMetrics(
	e *gmeasure.Experiment,
	sampleRange model.Duration,
) error {
	if err := c.Measure(e, LogQLQueryRate(sampleRange)); err != nil {
		return err
	}
	if err := c.Measure(e, LogQLQueryDurationAverage(sampleRange)); err != nil {
		return err
	}
	if err := c.Measure(e, LogQLQueryDurationQuantile(DefaultPercentile, sampleRange)); err != nil {
		return err
	}
	return nil
}

func (c *Client) executeScalarQuery(query string) (float64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	res, _, err := c.api.Query(ctx, query, time.Now())
	if err != nil {
		return 0.0, kverrors.Wrap(err, "failed executing query", "query", query)
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
		return 0.0, kverrors.Wrap(err, "failed to parse result for query", "query", query)
	}
}

func (c *Client) measureCommonRequestMetrics(
	e *gmeasure.Experiment,
	job, method, route, pathRoutes string,
	sampleRange model.Duration,
	annotation gmeasure.Annotation,
) error {
	var name, code, badCode, requestRateName, badRequestRateName string

	if method == GRPCMethod {
		name = fmt.Sprintf("%s successful GRPC %s", annotation, route)
		code = "success"
		requestRateName = name
		if pathRoutes == GRPCReadPathRoutes {
			requestRateName = "successful GRPC reads"
		}
	} else {
		name = fmt.Sprintf("%s 2xx %s", annotation, route)
		code = "2.*"
		requestRateName = name
		if pathRoutes == HTTPReadPathRoutes {
			requestRateName = "2xx reads"
		}

		badCode = "5.*"
		badRequestRateName = fmt.Sprintf("%s 5xx %s", annotation, route)
		if pathRoutes == HTTPReadPathRoutes {
			badRequestRateName = "5xx reads"
		}
	}

	// Rate request of 200 or success
	if err := c.Measure(e, RequestRate(requestRateName, job, pathRoutes, code, sampleRange, annotation)); err != nil {
		return err
	}
	if method == HTTPPostMethod {
		if err := c.Measure(e, RequestRate(badRequestRateName, job, pathRoutes, badCode, sampleRange, annotation)); err != nil {
			return err
		}
	}
	if err := c.Measure(e, RequestDurationAverage(name, job, method, route, code, sampleRange, annotation)); err != nil {
		return err
	}
	if err := c.Measure(e, RequestDurationQuantile(name, job, method, route, code, DefaultPercentile, sampleRange, annotation)); err != nil {
		return err
	}
	return nil
}
