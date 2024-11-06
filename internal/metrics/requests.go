package metrics

import (
	"fmt"

	"github.com/onsi/gomega/gmeasure"
	"github.com/prometheus/common/model"
)

const (
	HTTPGetMethod  = "GET"
	HTTPPostMethod = "POST"

	HTTPPushRoute       = "loki_api_v1_push"
	HTTPQueryRangeRoute = "loki_api_v1_query_range"
	HTTPReadPathRoutes  = "loki_api_v1_series|api_prom_series|api_prom_query|api_prom_label|api_prom_label_name_values|loki_api_v1_query|loki_api_v1_query_range|loki_api_v1_labels|loki_api_v1_label_name_values"
)

const (
	GRPCMethod = "gRPC"

	GRPCPushRoute        = "/logproto.Pusher/Push"
	GRPCQuerySampleRoute = "/logproto.Querier/QuerySample"
	GRPCReadPathRoutes   = "/logproto.Querier/Query|/logproto.Querier/QuerySample|/logproto.Querier/Label|/logproto.Querier/Series|/logproto.Querier/GetChunkIDs"
)

const (
	IndexReadName  = "Index successful reads"
	IndexWriteName = "Index successful writes"

	WriteOperation = true
	ReadOperation  = false
)

func RequestRate(
	name, job, route, code string,
	duration model.Duration,
	annotation gmeasure.Annotation,
) Measurement {
	return Measurement{
		Name: fmt.Sprintf("%s request rate", name),
		Query: fmt.Sprintf(
			`sum(rate(loki_request_duration_seconds_count{job=~".*%s.*", route=~"%s", status_code=~"%s"}[%s]))`,
			job, route, code, duration,
		),
		Unit:       RequestsPerSecondUnit,
		Annotation: annotation,
	}
}

func RequestDurationAverage(
	name, job, method, route, code string,
	duration model.Duration,
	annotation gmeasure.Annotation,
) Measurement {
	numerator := fmt.Sprintf(
		`sum(rate(loki_request_duration_seconds_sum{job=~".*%s.*", method="%s", route=~"%s", status_code=~"%s"}[%s]))`,
		job, method, route, code, duration,
	)
	// clamp_min is used to avoid division by zero which breaks the reporting
	denomintator := fmt.Sprintf(
		`clamp_min(sum(rate(loki_request_duration_seconds_count{job=~".*%s.*", method="%s", route=~"%s", status_code=~"%s"}[%s])),0.01)`,
		job, method, route, code, duration,
	)

	return Measurement{
		Name:       fmt.Sprintf("%s request duration avg", name),
		Query:      fmt.Sprintf("(%s / %s) * %d", numerator, denomintator, SecondsToMillisecondsMultiplier),
		Unit:       MillisecondsUnit,
		Annotation: annotation,
	}
}

func RequestDurationQuantile(
	name, job, method, route, code string,
	percentile int,
	duration model.Duration,
	annotation gmeasure.Annotation,
) Measurement {
	return Measurement{
		Name: fmt.Sprintf("%s request duration P%d", name, percentile),
		// clamp_min is used to avoid NaN which breaks the reporting
		Query: fmt.Sprintf(
			`histogram_quantile(0.%d, sum by (job, le) (clamp_min(rate(loki_request_duration_seconds_bucket{job=~".*%s.*", method="%s", route=~"%s", status_code=~"%s"}[%s]), 0.01))) * %d`,
			percentile, job, method, route, code, duration, SecondsToMillisecondsMultiplier,
		),
		Unit:       MillisecondsUnit,
		Annotation: annotation,
	}
}

func RequestIndexRequestRate(name, job string, writeOperation bool, code string, duration model.Duration) Measurement {
	equalityOperator := "="
	if !writeOperation {
		equalityOperator = "!="
	}
	return Measurement{
		Name: fmt.Sprintf("%s request rate", name),
		Query: fmt.Sprintf(
			`sum(rate(loki_index_request_duration_seconds_count{job=~".*%s.*", operation%s"index_chunk", status_code=~"%s"}[%s]))`,
			job, equalityOperator, code, duration,
		),
		Unit:       RequestsPerSecondUnit,
		Annotation: IngesterAnnotation,
	}
}
