package metrics

import "github.com/prometheus/common/model"

func (c *client) RequestDurationOkGrpcQuerySampleAvg(job string, duration model.Duration) (float64, error) {
	return c.requestDurationAvg(job, "gRPC", "/logproto.Querier/QuerySample", "success", duration)
}

func (c *client) RequestDurationOkGrpcQuerySampleP50(job string, duration model.Duration) (float64, error) {
	return c.requestDurationQuantile(job, "gRPC", "/logproto.Querier/QuerySample", "success", duration, 50)
}

func (c *client) RequestDurationOkGrpcQuerySampleP99(job string, duration model.Duration) (float64, error) {
	return c.requestDurationQuantile(job, "gRPC", "/logproto.Querier/QuerySample", "success", duration, 99)
}

func (c *client) RequestDurationOkGrpcPushAvg(job string, duration model.Duration) (float64, error) {
	return c.requestDurationAvg(job, "gRPC", "/logproto.Pusher/Push", "success", duration)
}

func (c *client) RequestDurationOkGrpcPushP50(job string, duration model.Duration) (float64, error) {
	return c.requestDurationQuantile(job, "gRPC", "/logproto.Pusher/Push", "success", duration, 50)
}

func (c *client) RequestDurationOkGrpcPushP99(job string, duration model.Duration) (float64, error) {
	return c.requestDurationQuantile(job, "gRPC", "/logproto.Pusher/Push", "success", duration, 99)
}

func (c *client) RequestReadsGrpcQPS(job string, duration model.Duration) (float64, error) {
	route := "/logproto.Querier/Query|/logproto.Querier/QuerySample|/logproto.Querier/Label|/logproto.Querier/Series"
	return c.requestQPS(job, route, "success", duration)
}

func (c *client) RequestWritesGrpcQPS(job string, duration model.Duration) (float64, error) {
	route := "/logproto.Pusher/Push"
	return c.requestQPS(job, route, "success", duration)
}
