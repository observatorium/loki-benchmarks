package metrics

import "github.com/prometheus/common/model"

func (c *client) RequestDurationOkQueryRangeAvg(job string, duration model.Duration) (float64, error) {
	route := "loki_api_v1_query_range"
	return c.requestDurationAvg(job, "GET", route, "2.*", duration)
}

func (c *client) RequestDurationOkQueryRangeP50(job string, duration model.Duration) (float64, error) {
	route := "loki_api_v1_query_range"
	return c.requestDurationQuantile(job, "GET", route, "2.*", duration, 50)
}

func (c *client) RequestDurationOkQueryRangeP99(job string, duration model.Duration) (float64, error) {
	route := "loki_api_v1_query_range"
	return c.requestDurationQuantile(job, "GET", route, "2.*", duration, 99)
}

func (c *client) RequestDurationOkPushAvg(job string, duration model.Duration) (float64, error) {
	route := "loki_api_v1_push"
	return c.requestDurationAvg(job, "POST", route, "2.*", duration)
}

func (c *client) RequestDurationOkPushP50(job string, duration model.Duration) (float64, error) {
	route := "loki_api_v1_push"
	return c.requestDurationQuantile(job, "POST", route, "2.*", duration, 50)
}

func (c *client) RequestDurationOkPushP99(job string, duration model.Duration) (float64, error) {
	route := "loki_api_v1_push"
	return c.requestDurationQuantile(job, "POST", route, "2.*", duration, 99)
}

func (c *client) RequestReadsQPS(job string, duration model.Duration) (float64, error) {
	route := "loki_api_v1_series|api_prom_series|api_prom_query|api_prom_label|api_prom_label_name_values|loki_api_v1_query|loki_api_v1_query_range|loki_api_v1_labels|loki_api_v1_label_name_values"
	return c.requestQPS(job, route, "2.*", duration)
}

func (c *client) RequestWritesQPS(job string, duration model.Duration) (float64, error) {
	route := "loki_api_v1_push"
	return c.requestQPS(job, route, "2.*", duration)
}

func (c *client) RequestQueryRangeThroughput(job string, duration model.Duration) (float64, error) {
	endpoint := "loki_api_v1_query|loki_api_v1_query_range|api_prom_query|metrics"
	return c.requestThroughput(job, endpoint, "2.*", "range", "filter|metric", "slow", "4e+09", duration)
}
