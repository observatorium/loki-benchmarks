package metrics

func (c *client) RequestDurationOkGrpcQuerySampleAvg(job, duration string) (float64, error) {
    return c.requestDurationAvg(job, "gRPC", "/logproto.Querier/QuerySample", "success", duration)
}

func (c *client) RequestDurationOkGrpcQuerySampleP50(job, duration string) (float64, error) {
    return c.requestDurationQuantile(job, "gRPC", "/logproto.Querier/QuerySample", "success", duration, 50)
}

func (c *client) RequestDurationOkGrpcQuerySampleP99(job, duration string) (float64, error) {
    return c.requestDurationQuantile(job, "gRPC", "/logproto.Querier/QuerySample", "success", duration, 99)
}

func (c *client) RequestDurationOkGrpcPushAvg(job, duration string) (float64, error) {
    return c.requestDurationAvg(job, "gRPC", "/logproto.Pusher/Push", "success", duration)
}

func (c *client) RequestDurationOkGrpcPushP50(job, duration string) (float64, error) {
    return c.requestDurationQuantile(job, "gRPC", "/logproto.Pusher/Push", "success", duration, 50)
}

func (c *client) RequestDurationOkGrpcPushP99(job, duration string) (float64, error) {
    return c.requestDurationQuantile(job, "gRPC", "/logproto.Pusher/Push", "success", duration, 99)
}
