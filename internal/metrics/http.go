package metrics

import "github.com/prometheus/common/model"

func (c *client) RequestDurationOkQueryAvg(job string, duration model.Duration) (float64, error) {
	return c.requestDurationAvg(job, "GET", "loki_api_v1_query", "2.*", duration)
}

func (c *client) RequestDurationOkQueryP50(job string, duration model.Duration) (float64, error) {
	return c.requestDurationQuantile(job, "GET", "loki_api_v1_query", "2.*", duration, 50)
}

func (c *client) RequestDurationOkQueryP99(job string, duration model.Duration) (float64, error) {
	return c.requestDurationQuantile(job, "GET", "loki_api_v1_query", "2.*", duration, 99)
}

func (c *client) RequestDurationOkQueryRangeAvg(job string, duration model.Duration) (float64, error) {
	return c.requestDurationAvg(job, "GET", "loki_api_v1_query_range", "2.*", duration)
}

func (c *client) RequestDurationOkQueryRangeP50(job string, duration model.Duration) (float64, error) {
	return c.requestDurationQuantile(job, "GET", "loki_api_v1_query_range", "2.*", duration, 50)
}

func (c *client) RequestDurationOkQueryRangeP99(job string, duration model.Duration) (float64, error) {
	return c.requestDurationQuantile(job, "GET", "loki_api_v1_query_range", "2.*", duration, 99)
}

func (c *client) RequestDurationOkPushAvg(job string, duration model.Duration) (float64, error) {
	return c.requestDurationAvg(job, "POST", "loki_api_v1_push", "2.*", duration)
}

func (c *client) RequestDurationOkPushP50(job string, duration model.Duration) (float64, error) {
	return c.requestDurationQuantile(job, "POST", "loki_api_v1_push", "2.*", duration, 50)
}

func (c *client) RequestDurationOkPushP99(job string, duration model.Duration) (float64, error) {
	return c.requestDurationQuantile(job, "POST", "loki_api_v1_push", "2.*", duration, 99)
}
