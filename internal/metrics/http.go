package metrics

import (
	"github.com/onsi/gomega/gmeasure"
	"github.com/prometheus/common/model"
)

const (
	HTTPGetMethod  = "GET"
	HTTPPostMethod = "POST"

	HTTPQueryRangeRoute = "loki_api_v1_query_range"
	HTTPPushRoute       = "loki_api_v1_push"
)

func RequestWritesQPS(job string, duration model.Duration, annotation gmeasure.Annotation) Measurement {
	return requestRate("2xx push", job, HTTPPushRoute, "2.*", duration, annotation)
}

func RequestReadsQPS(job string, duration model.Duration, annotation gmeasure.Annotation) Measurement {
	route := "loki_api_v1_series|api_prom_series|api_prom_query|api_prom_label|api_prom_label_name_values|loki_api_v1_query|loki_api_v1_query_range|loki_api_v1_labels|loki_api_v1_label_name_values"
	return requestRate("2xx reads", job, route, "2.*", duration, annotation)
}

func RequestQueryRangeThroughput(job string, duration model.Duration, annotation gmeasure.Annotation) Measurement {
	endpoint := "loki_api_v1_query|loki_api_v1_query_range|api_prom_query|metrics"
	return requestThroughput("2xx reads", job, endpoint, "2.*", "range", "filter|metric", "slow", "4e+09", duration, annotation)
}
