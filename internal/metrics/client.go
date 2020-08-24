package metrics

import (
    "context"
    "fmt"
    "time"

    "github.com/prometheus/client_golang/api"
    v1 "github.com/prometheus/client_golang/api/prometheus/v1"
    "github.com/prometheus/common/model"
)

type MetricType string

const (
    DistributorBytesReceivedTotal MetricType = "loki_distributor_bytes_received_total"
)

type Client interface {
    DistributorBytesReceivedTotal() (float64, error)

    // HTTP API
    RequestDurationOkReadsAvg(job, duration string) (float64, error)
    RequestDurationOkReadsP50(job, duration string) (float64, error)
    RequestDurationOkReadsP99(job, duration string) (float64, error)

    RequestDurationOkWritesAvg(job, duration string) (float64, error)
    RequestDurationOkWritesP50(job, duration string) (float64, error)
    RequestDurationOkWritesP99(job, duration string) (float64, error)

    RequestDurationOkQueryRangeAvg(job, duration string) (float64, error)
    RequestDurationOkQueryRangeP50(job, duration string) (float64, error)
    RequestDurationOkQueryRangeP99(job, duration string) (float64, error)

    // GRPC API
    RequestDurationOkGrpcQuerySampleAvg(job, duration string) (float64, error)
    RequestDurationOkGrpcQuerySampleP50(job, duration string) (float64, error)
    RequestDurationOkGrpcQuerySampleP99(job, duration string) (float64, error)

    RequestDurationOkGrpcPushAvg(job, duration string) (float64, error)
    RequestDurationOkGrpcPushP50(job, duration string) (float64, error)
    RequestDurationOkGrpcPushP99(job, duration string) (float64, error)
}

type client struct {
    api     v1.API
    timeout time.Duration
}

func NewClient(url string, timeout time.Duration) (Client, error) {
    pc, err := api.NewClient(api.Config{Address: url})
    if err != nil {
        return nil, fmt.Errorf("failed creating prometheus client: %w", err)
    }

    return &client{
        api:     v1.NewAPI(pc),
        timeout: timeout,
    }, nil
}

func (c *client) requestDurationAvg(job, method, route, code, duration string) (float64, error) {
    routeParam := ""
    if route != "" {
        routeParam = fmt.Sprintf(`route="%s",`, route)
    }

    query := fmt.Sprintf(
        `100 * (sum by (job) (rate(loki_request_duration_seconds_sum{job="%s", method="%s", %s status_code=~"%s"}[%s])) / sum by (job) (rate(loki_request_duration_seconds_count{job="%s", method="%s", %s status_code=~"%s"}[%s])))`,
        job, method, routeParam, code, duration,
        job, method, routeParam, code, duration,
    )

    return c.executeScalarQuery(query)
}

func (c *client) requestDurationQuantile(job, method, route, code, duration string, percentile int) (float64, error) {
    routeParam := ""
    if route != "" {
        routeParam = fmt.Sprintf(`route="%s",`, route)
    }

    query := fmt.Sprintf(
        `histogram_quantile(0.%d, sum by (job, le) (rate(loki_request_duration_seconds_bucket{job="%s", method="%s", %s status_code=~"%s"}[%s])))`,
        percentile, job, method, routeParam, code, duration,
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

    if res.Type() == model.ValScalar {
        value := res.(*model.Scalar)
        return float64(value.Value), nil
    }

    if res.Type() == model.ValVector {
        vec := res.(model.Vector)
        if vec.Len() == 0 {
            return 0.0, fmt.Errorf("empty result set for query: %s", query)
        }

        return float64(vec[0].Value), nil
    }

    return 0.0, fmt.Errorf("failed to parse result for query: %s", query)
}
