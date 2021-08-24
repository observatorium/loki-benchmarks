package metrics

import "github.com/prometheus/common/model"

func (c *client) RequestDurationOkQueryAvg(label, job string, duration model.Duration) (float64, error) {
	route := "loki_api_v1_query|metrics"
	return c.requestDurationAvg(label, job, "GET", route, "2.*", duration)
}

func (c *client) RequestDurationOkQueryP50(label, job string, duration model.Duration) (float64, error) {
	route := "loki_api_v1_query|metrics"
	return c.requestDurationQuantile(label, job, "GET", route, "2.*", duration, 50)
}

func (c *client) RequestDurationOkQueryP99(label, job string, duration model.Duration) (float64, error) {
	route := "loki_api_v1_query|metrics"
	return c.requestDurationQuantile(label, job, "GET", route, "2.*", duration, 99)
}

func (c *client) RequestDurationOkQueryRangeAvg(label, job string, duration model.Duration) (float64, error) {
	route := "loki_api_v1_query_range|metrics"
	return c.requestDurationAvg(label, job, "GET", route, "2.*", duration)
}

func (c *client) RequestDurationOkQueryRangeP50(label, job string, duration model.Duration) (float64, error) {
	route := "loki_api_v1_query_range|metrics"
	return c.requestDurationQuantile(label, job, "GET", route, "2.*", duration, 50)
}

func (c *client) RequestDurationOkQueryRangeP99(label, job string, duration model.Duration) (float64, error) {
	route := "loki_api_v1_query_range|metrics"
	return c.requestDurationQuantile(label, job, "GET", route, "2.*", duration, 99)
}

func (c *client) RequestDurationOkPushAvg(label, job string, duration model.Duration) (float64, error) {
	route := "loki_api_v1_push|metrics"
	return c.requestDurationAvg(label, job, "POST", route, "2.*", duration)
}

func (c *client) RequestDurationOkPushP50(label, job string, duration model.Duration) (float64, error) {
	route := "loki_api_v1_push|metrics"
	return c.requestDurationQuantile(label, job, "POST", route, "2.*", duration, 50)
}

func (c *client) RequestDurationOkPushP99(label, job string, duration model.Duration) (float64, error) {
	route := "loki_api_v1_push|metrics"
	return c.requestDurationQuantile(label, job, "POST", route, "2.*", duration, 99)
}

func (c *client) RequestReadsQPS(label, job string, duration model.Duration) (float64, error) {
	route := "loki_api_v1_series|api_prom_series|api_prom_query|api_prom_label|api_prom_label_name_values|loki_api_v1_query|loki_api_v1_query_range|loki_api_v1_labels|loki_api_v1_label_name_values|metrics"
	return c.requestQPS(label, job, route, "2.*", duration)
}

func (c *client) RequestWritesQPS(label, job string, duration model.Duration) (float64, error) {
	route := "loki_api_v1_push|metrics"
	return c.requestQPS(label, job, route, "2.*|429", duration)
}
