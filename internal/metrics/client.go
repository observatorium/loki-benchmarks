package metrics

import (
    "context"
    "fmt"
    "time"

    "github.com/prometheus/client_golang/api"
    v1 "github.com/prometheus/client_golang/api/prometheus/v1"
    "github.com/prometheus/common/model"
)

type Client interface {
    RequestDurationOkWritesP99(job, duration string) (float64, error)
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

func (c *client) RequestDurationOkWritesP99(job, duration string) (float64, error) {
    query := fmt.Sprintf(
        `histogram_quantile(0.99, sum by (job, le) (rate(loki_request_duration_seconds_bucket{job="%s", method="POST", status_code=~"2.*"}[%s])))`,
        job,
        duration,
    )

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
            return 0.0, fmt.Errorf("empty result set for job: %s", job)
        }

        return float64(vec[0].Value), nil
    }

    return 0.0, fmt.Errorf("failed to parse result for job: %s", job)
}
