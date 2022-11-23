package metrics

import (
	"time"

	"github.com/onsi/gomega/gmeasure"
)

const (
	HTTPQueryRangeRoute = "loki_api_v1_query_range"
	HTTPPushRoute       = "loki_api_v1_push"
)

func RequestWritesQPS(job string, duration time.Duration, annotation gmeasure.Annotation) Measurement {
	return requestRate("2xx push", job, HTTPPushRoute, "2.*", duration, annotation)
}

func RequestDurationOkPushAvg(job string, duration time.Duration, annotation gmeasure.Annotation) Measurement {
	return requestDurationAvg("2xx push", job, "POST", HTTPPushRoute, "2.*", duration, annotation)
}

func RequestDurationOkPushPercentile(percentile int, job string, duration time.Duration, annotation gmeasure.Annotation) Measurement {
	return requestDurationQuantile("2xx push", job, "POST", HTTPPushRoute, "2.*", percentile, duration, annotation)
}

func RequestReadsQPS(job string, duration time.Duration, annotation gmeasure.Annotation) Measurement {
	route := "loki_api_v1_series|api_prom_series|api_prom_query|api_prom_label|api_prom_label_name_values|loki_api_v1_query|loki_api_v1_query_range|loki_api_v1_labels|loki_api_v1_label_name_values"
	return requestRate("2xx reads", job, route, "2.*", duration, annotation)
}

func RequestDurationOkQueryRangeAvg(job string, duration time.Duration, annotation gmeasure.Annotation) Measurement {
	return requestDurationAvg("2xx reads", job, "GET", HTTPQueryRangeRoute, "2.*", duration, annotation)
}

func RequestDurationOkQueryRangePercentile(percentile int, job string, duration time.Duration, annotation gmeasure.Annotation) Measurement {
	return requestDurationQuantile("2xx reads", job, "GET", HTTPQueryRangeRoute, "2.*", percentile, duration, annotation)
}

func RequestQueryRangeThroughput(job string, duration time.Duration, annotation gmeasure.Annotation) Measurement {
	endpoint := "loki_api_v1_query|loki_api_v1_query_range|api_prom_query|metrics"
	return requestThroughput("2xx reads", job, endpoint, "2.*", "range", "filter|metric", "slow", "4e+09", duration, annotation)
}
