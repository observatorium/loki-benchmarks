package metrics

func (c *client) RequestDurationOkQueryRangeAvg(job, duration string) (float64, error) {
    return c.requestDurationAvg(job, "GET", "loki_api_v1_query_range", "2.*", duration)
}

func (c *client) RequestDurationOkQueryRangeP50(job, duration string) (float64, error) {
    return c.requestDurationQuantile(job, "GET", "loki_api_v1_query_range", "2.*", duration, 50)
}

func (c *client) RequestDurationOkQueryRangeP99(job, duration string) (float64, error) {
    return c.requestDurationQuantile(job, "GET", "loki_api_v1_query_range", "2.*", duration, 99)
}

func (c *client) RequestDurationOkPushAvg(job, duration string) (float64, error) {
    return c.requestDurationAvg(job, "POST", "loki_api_v1_push", "2.*", duration)
}

func (c *client) RequestDurationOkPushP50(job, duration string) (float64, error) {
    return c.requestDurationQuantile(job, "POST", "loki_api_v1_push", "2.*", duration, 50)
}

func (c *client) RequestDurationOkPushP99(job, duration string) (float64, error) {
    return c.requestDurationQuantile(job, "POST", "loki_api_v1_push", "2.*", duration, 99)
}
