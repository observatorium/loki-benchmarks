package metrics

import (
	"fmt"

	"github.com/onsi/gomega/gmeasure"
	"github.com/prometheus/common/model"
)

func requestRate(
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

func requestDurationAvg(
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

func requestDurationQuantile(
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
