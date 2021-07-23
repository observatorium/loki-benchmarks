package metrics

import "github.com/prometheus/common/model"

func (c *client) RequestDurationOkGrpcQuerySampleAvg(label, job string, duration model.Duration) (float64, error) {
	return c.requestDurationAvg(label, job, "gRPC", "/logproto.Querier/QuerySample", "success", duration)
}

func (c *client) RequestDurationOkGrpcQuerySampleP50(label, job string, duration model.Duration) (float64, error) {
	return c.requestDurationQuantile(label, job, "gRPC", "/logproto.Querier/QuerySample", "success", duration, 50)
}

func (c *client) RequestDurationOkGrpcQuerySampleP99(label, job string, duration model.Duration) (float64, error) {
	return c.requestDurationQuantile(label, job, "gRPC", "/logproto.Querier/QuerySample", "success", duration, 99)
}

func (c *client) RequestDurationOkGrpcPushAvg(label, job string, duration model.Duration) (float64, error) {
	return c.requestDurationAvg(label, job, "gRPC", "/logproto.Pusher/Push", "success", duration)
}

func (c *client) RequestDurationOkGrpcPushP50(label, job string, duration model.Duration) (float64, error) {
	return c.requestDurationQuantile(label, job, "gRPC", "/logproto.Pusher/Push", "success", duration, 50)
}

func (c *client) RequestDurationOkGrpcPushP99(label, job string, duration model.Duration) (float64, error) {
	return c.requestDurationQuantile(label, job, "gRPC", "/logproto.Pusher/Push", "success", duration, 99)
}

func (c *client) RequestReadsGrpcQPS(label, job string, duration model.Duration) (float64, error) {
	route := "/logproto.Querier/Query|/logproto.Querier/QuerySample|/logproto.Querier/Label|/logproto.Querier/Series|/logproto.Querier/GetChunkIDs"
	return c.requestQPS(label, job, route, "success", duration)
}

func (c *client) RequestWritesGrpcQPS(label, job string, duration model.Duration) (float64, error) {
	route := "/logproto.Pusher/Push"
	return c.requestQPS(label, job, route, "success", duration)
}
