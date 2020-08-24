package metrics

func (c *client) RequestDurationOkReadsAvg(job, duration string) (float64, error) {
    return c.requestDurationAvg(job, "GET", "", "2.*", duration)
}

func (c *client) RequestDurationOkReadsP50(job, duration string) (float64, error) {
    return c.requestDurationQuantile(job, "GET", "", "2.*", duration, 50)
}

func (c *client) RequestDurationOkReadsP99(job, duration string) (float64, error) {
    return c.requestDurationQuantile(job, "GET", "", "2.*", duration, 99)
}

func (c *client) RequestDurationOkQueryRangeAvg(job, duration string) (float64, error) {
    return c.requestDurationAvg(job, "GET", "loki_api_v1_query_range", "2.*", duration)
}

func (c *client) RequestDurationOkQueryRangeP50(job, duration string) (float64, error) {
    return c.requestDurationQuantile(job, "GET", "loki_api_v1_query_range", "2.*", duration, 50)
}

func (c *client) RequestDurationOkQueryRangeP99(job, duration string) (float64, error) {
    return c.requestDurationQuantile(job, "GET", "loki_api_v1_query_range", "2.*", duration, 99)
}

func (c *client) RequestDurationOkWritesAvg(job, duration string) (float64, error) {
    return c.requestDurationAvg(job, "POST", "", "2.*", duration)
}

func (c *client) RequestDurationOkWritesP50(job, duration string) (float64, error) {
    return c.requestDurationQuantile(job, "POST", "", "2.*", duration, 50)
}

func (c *client) RequestDurationOkWritesP99(job, duration string) (float64, error) {
    return c.requestDurationQuantile(job, "POST", "", "2.*", duration, 99)
}
