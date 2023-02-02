package metrics

import (
	"fmt"

	"github.com/onsi/gomega/gmeasure"
	"github.com/prometheus/common/model"
)

const (
	LogQLAnnotation = gmeasure.Annotation("logql")
)

func LogQLQueryRate(duration model.Duration) Measurement {
	return Measurement{
		Name: "LogQL query rate",
		Query: fmt.Sprintf(
			`sum(rate(logql_query_duration_seconds_count[%s]))`,
			duration,
		),
		Unit:       QueriesPerSecondUnit,
		Annotation: LogQLAnnotation,
	}
}

func LogQLQueryDurationAverage(duration model.Duration) Measurement {
	numerator := fmt.Sprintf(
		`sum(rate(logql_query_duration_seconds_sum[%s]))`,
		duration,
	)

	denomintator := fmt.Sprintf(
		`sum(rate(logql_query_duration_seconds_count[%s]))`,
		duration,
	)

	return Measurement{
		Name:       "LogQL query duration avg",
		Query:      fmt.Sprintf(`(%s / %s) * %d`, numerator, denomintator, SecondsToMillisecondsMultiplier),
		Unit:       MillisecondsUnit,
		Annotation: LogQLAnnotation,
	}
}

func LogQLQueryDurationQuantile(percentile int, duration model.Duration) Measurement {
	return Measurement{
		Name: fmt.Sprintf("LogQL query duration P%d", percentile),
		Query: fmt.Sprintf(
			`histogram_quantile(0.%d, sum by (le) (rate(logql_query_duration_seconds_bucket[%s]))) * %d`,
			percentile, duration, SecondsToMillisecondsMultiplier,
		),
		Unit:       MillisecondsUnit,
		Annotation: LogQLAnnotation,
	}
}

func LogQLQueryLatencyAverage(
	code, pod string,
	duration model.Duration,
	annotation gmeasure.Annotation,
) Measurement {
	numerator := fmt.Sprintf(
		`sum(rate(loki_logql_querystats_latency_seconds_sum{pod=~"%s.*", status_code=~"%s"}[%s]))`,
		pod, code, duration,
	)

	denomintator := fmt.Sprintf(
		`sum(rate(loki_logql_querystats_latency_seconds_count{pod=~"%s.*", status_code=~"%s"}[%s]))`,
		pod, code, duration,
	)

	return Measurement{
		Name:       "LogQL query latency avg",
		Query:      fmt.Sprintf(`(%s / %s) * %d`, numerator, denomintator, SecondsToMillisecondsMultiplier),
		Unit:       MillisecondsUnit,
		Annotation: annotation,
	}
}

func LogQLQueryLatencyQuantile(
	code, pod string,
	percentile int,
	duration model.Duration,
	annotation gmeasure.Annotation,
) Measurement {
	return Measurement{
		Name: fmt.Sprintf("LogQL query latency P%d", percentile),
		Query: fmt.Sprintf(
			`histogram_quantile(0.%d, sum by (job, le) (rate(loki_logql_querystats_latency_seconds_bucket{pod=~"%s.*", status_code=~"%s"}[%s]))) * %d`,
			percentile, pod, code, duration, SecondsToMillisecondsMultiplier,
		),
		Unit:       MillisecondsUnit,
		Annotation: annotation,
	}
}

func LogQLQueryMBpSProcessedAverage(
	code, pod string,
	duration model.Duration,
	annotation gmeasure.Annotation,
) Measurement {
	numerator := fmt.Sprintf(
		`sum(rate(loki_logql_querystats_bytes_processed_per_seconds_sum{pod=~"%s.*", status_code=~"%s"}[%s]))`,
		pod, code, duration,
	)

	denomintator := fmt.Sprintf(
		`sum(rate(loki_logql_querystats_bytes_processed_per_seconds_count{pod=~"%s.*", status_code=~"%s"}[%s]))`,
		pod, code, duration,
	)

	return Measurement{
		Name:       "LogQL query MBps processed avg",
		Query:      fmt.Sprintf(`(%s / %s) / %d`, numerator, denomintator, BytesToMegabytesMultiplier),
		Unit:       MegabytesPerSecondUnit,
		Annotation: annotation,
	}
}

func LogQLQueryMBpSProcessedQuantile(
	code, pod string,
	percentile int,
	duration model.Duration,
	annotation gmeasure.Annotation,
) Measurement {
	return Measurement{
		Name: fmt.Sprintf("LogQL query MBps processed P%d", percentile),
		Query: fmt.Sprintf(
			`histogram_quantile(0.%d, sum by (job, le) (rate(loki_logql_querystats_bytes_processed_per_seconds_bucket{pod=~"%s.*", status_code=~"%s"}[%s]))) / %d`,
			percentile, pod, code, duration, BytesToMegabytesMultiplier,
		),
		Unit:       MegabytesPerSecondUnit,
		Annotation: annotation,
	}
}
