package metrics

import (
	"fmt"

	"github.com/onsi/gomega/gmeasure"
	"github.com/prometheus/common/model"
)

const (
	HTTPGetMethod  = "GET"
	HTTPPostMethod = "POST"

	HTTPQueryRangeRoute = "loki_api_v1_query_range"
	HTTPPushRoute       = "loki_api_v1_push"

	HTTPReadPathRoutes = "loki_api_v1_series|api_prom_series|api_prom_query|api_prom_label|api_prom_label_name_values|loki_api_v1_query|loki_api_v1_query_range|loki_api_v1_labels|loki_api_v1_label_name_values"

	GRPCMethod = "gRPC"

	GRPCPushRoute        = "/logproto.Pusher/Push"
	GRPCQuerySampleRoute = "/logproto.Querier/QuerySample"

	GRPCReadPathRoutes = "/logproto.Querier/Query|/logproto.Querier/QuerySample|/logproto.Querier/Label|/logproto.Querier/Series|/logproto.Querier/GetChunkIDs"
)

func RequestRate(
	name, job, route, code string,
	duration model.Duration,
	annotation gmeasure.Annotation,
) Measurement {
	return Measurement{
		Name: fmt.Sprintf("%s request rate", name),
		Query: fmt.Sprintf(
			`sum(irate(loki_request_duration_seconds_count{job=~".*%s.*", route=~"%s", status_code=~"%s"}[%s]))`,
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
		`sum(irate(loki_request_duration_seconds_sum{job=~".*%s.*", method="%s", route=~"%s", status_code=~"%s"}[%s]))`,
		job, method, route, code, duration,
	)
	denomintator := fmt.Sprintf(
		`sum(irate(loki_request_duration_seconds_count{job=~".*%s.*", method="%s", route=~"%s", status_code=~"%s"}[%s]))`,
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
		Query: fmt.Sprintf(
			`histogram_quantile(0.%d, sum by (job, le) (irate(loki_request_duration_seconds_bucket{job=~".*%s.*", method="%s", route=~"%s", status_code=~"%s"}[%s]))) * %d`,
			percentile, job, method, route, code, duration, SecondsToMillisecondsMultiplier,
		),
		Unit:       MillisecondsUnit,
		Annotation: annotation,
	}
}

func RequestQueryRangeThroughput(job string, duration model.Duration, annotation gmeasure.Annotation) Measurement {
	endpoint := "loki_api_v1_query|loki_api_v1_query_range|api_prom_query|metrics"
	return requestThroughput("2xx reads", job, endpoint, "2.*", "range", "filter|metric", "slow", "4e+09", duration, annotation)
}

func requestThroughput(
	name, job, endpoint, code, queryRange, metricType, latencyType, le string,
	duration model.Duration,
	annotation gmeasure.Annotation,
) Measurement {
	numerator := fmt.Sprintf(
		`sum by (namespace, job) (irate(loki_logql_querystats_bytes_processed_per_seconds_bucket{status_code=~"%s", endpoint=~"%s", range="%s", type=~"%s", job=~".*%s.*", latency_type="%s", le="%s"}[%s]))`,
		code, endpoint, queryRange, metricType, job, latencyType, le, duration,
	)
	denomintator := fmt.Sprintf(
		`sum by (namespace, job) (irate(loki_logql_querystats_bytes_processed_per_seconds_count{status_code=~"%s", endpoint=~"%s", range="%s", type=~"%s", job=~".*%s.*"}[%s]))`,
		code, endpoint, queryRange, metricType, job, duration,
	)

	return Measurement{
		Name:       fmt.Sprintf("%s throughput", name),
		Query:      fmt.Sprintf("(%s / %s) * %d", numerator, denomintator, SecondsToMillisecondsMultiplier),
		Unit:       MillisecondsUnit,
		Annotation: annotation,
	}
}

func requestBoltDBShipperQPS(name, job, operation, code string, duration model.Duration) Measurement {
	return Measurement{
		Name: fmt.Sprintf("%s request rate", name),
		Query: fmt.Sprintf(
			`sum(irate(loki_boltdb_shipper_request_duration_seconds_count{job=~".*%s.*", operation="%s", status_code=~"%s"}[%s]))`,
			job, operation, code, duration,
		),
		Unit:       RequestsPerSecondUnit,
		Annotation: IngesterAnnotation,
	}
}
