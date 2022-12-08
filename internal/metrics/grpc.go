package metrics

import (
	"github.com/onsi/gomega/gmeasure"
	"github.com/prometheus/common/model"
)

const (
	GRPCMethod = "gRPC"

	GRPCPushRoute        = "/logproto.Pusher/Push"
	GRPCQuerySampleRoute = "/logproto.Querier/QuerySample"
)

func RequestWritesGrpcQPS(job string, duration model.Duration, annotation gmeasure.Annotation) Measurement {
	return requestRate("successful GRPC push", job, GRPCPushRoute, "success", duration, annotation)
}

func RequestReadsGrpcQPS(job string, duration model.Duration, annotation gmeasure.Annotation) Measurement {
	route := "/logproto.Querier/Query|/logproto.Querier/QuerySample|/logproto.Querier/Label|/logproto.Querier/Series|/logproto.Querier/GetChunkIDs"
	return requestRate("successful GRPC reads", job, route, "success", duration, annotation)
}
