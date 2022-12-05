package metrics

import (
	"github.com/onsi/gomega/gmeasure"
	"github.com/prometheus/common/model"
)

const (
	GRPCPushRoute        = "/logproto.Pusher/Push"
	GRPCQuerySampleRoute = "/logproto.Querier/QuerySample"
)

func RequestWritesGrpcQPS(job string, duration model.Duration, annotation gmeasure.Annotation) Measurement {
	return requestRate("successful GRPC push", job, GRPCPushRoute, "success", duration, annotation)
}

func RequestDurationOkGrpcPushAvg(job string, duration model.Duration, annotation gmeasure.Annotation) Measurement {
	return requestDurationAvg("successful GRPC push", job, "gRPC", GRPCPushRoute, "success", duration, annotation)
}

func RequestDurationOkGrpcPushPercentile(percentile int, job string, duration model.Duration, annotation gmeasure.Annotation) Measurement {
	return requestDurationQuantile("successful GRPC push", job, "gRPC", GRPCPushRoute, "success", percentile, duration, annotation)
}

func RequestReadsGrpcQPS(job string, duration model.Duration, annotation gmeasure.Annotation) Measurement {
	route := "/logproto.Querier/Query|/logproto.Querier/QuerySample|/logproto.Querier/Label|/logproto.Querier/Series|/logproto.Querier/GetChunkIDs"
	return requestRate("successful GRPC reads", job, route, "success", duration, annotation)
}

func RequestDurationOkGrpcQuerySampleAvg(job string, duration model.Duration, annotation gmeasure.Annotation) Measurement {
	return requestDurationAvg("successful GRPC reads", job, "gRPC", GRPCQuerySampleRoute, "success", duration, annotation)
}

func RequestDurationOkGrpcQuerySamplePercentile(percentile int, job string, duration model.Duration, annotation gmeasure.Annotation) Measurement {
	return requestDurationQuantile("successful GRPC reads", job, "gRPC", GRPCQuerySampleRoute, "success", percentile, duration, annotation)
}
