package metrics

import "fmt"

func (c *client) RequestDurationOkWritesAvg(job, duration string) (float64, error) {
    query := fmt.Sprintf(
        `100 * (sum by (job) (rate(loki_request_duration_seconds_sum{job="%s", method="POST", status_code=~"2.*"}[%s])) / sum by (job) (rate(loki_request_duration_seconds_count{job="%s", method="POST", status_code=~"2.*"}[%s])))`,
        job,
        duration,
        job,
        duration,
    )

    return c.executeScalarQuery(query)
}

func (c *client) RequestDurationOkWritesP50(job, duration string) (float64, error) {
    return c.requestDurationOkWrites(job, duration, 50)
}

func (c *client) RequestDurationOkWritesP99(job, duration string) (float64, error) {
    return c.requestDurationOkWrites(job, duration, 99)
}

func (c *client) requestDurationOkWrites(job, duration string, percentile int) (float64, error) {
    query := fmt.Sprintf(
        `histogram_quantile(0.%d, sum by (job, le) (rate(loki_request_duration_seconds_bucket{job="%s", method="POST", status_code=~"2.*"}[%s])))`,
        percentile,
        job,
        duration,
    )

    return c.executeScalarQuery(query)
}
