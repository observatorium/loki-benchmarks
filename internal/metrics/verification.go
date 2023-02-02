package metrics

import (
	"fmt"

	"github.com/onsi/gomega/gmeasure"
	"github.com/prometheus/common/model"
)

const (
	SecondsPerDay int = 60 * 60 * 24

	MegabytesPerSecondUnit = gmeasure.Units("MBps")
	GigabytesPerDayUnit    = gmeasure.Units("GBpd")

	LoadGeneratorAnnotation = gmeasure.Annotation("generator")
)

// This measurments are only meant to verify the configuration of the
// generator and are not actually important to the experiment.

func DistributorGiPDReceivedTotal(duration model.Duration) Measurement {
	return Measurement{
		Name: "Total Projected Bytes Received",
		Query: fmt.Sprintf(
			`sum(rate(loki_distributor_bytes_received_total[%s])) / %d * %d`,
			duration, BytesToGigabytesMultiplier, SecondsPerDay,
		),
		Unit:       GigabytesPerDayUnit,
		Annotation: DistributorAnnotation,
	}
}

func LoadNetworkTotal(pod string, duration model.Duration) Measurement {
	return Measurement{
		Name: "Total Bytes Transmitted",
		Query: fmt.Sprintf(
			`sum(rate(container_network_transmit_bytes_total{pod=~"%s-.*"}[%s])) / %d`,
			pod, duration, BytesToMegabytesMultiplier,
		),
		Unit:       MegabytesPerSecondUnit,
		Annotation: LoadGeneratorAnnotation,
	}
}

func LoadNetworkGiPDTotal(pod string, duration model.Duration) Measurement {
	return Measurement{
		Name: "Total Projected Bytes Transmitted",
		Query: fmt.Sprintf(
			`sum(rate(container_network_transmit_bytes_total{pod=~"%s-.*"}[%s])) / %d * %d`,
			pod, duration, BytesToGigabytesMultiplier, SecondsPerDay,
		),
		Unit:       GigabytesPerDayUnit,
		Annotation: LoadGeneratorAnnotation,
	}
}

func LokiStreamsInMemoryTotal(duration model.Duration) Measurement {
	return Measurement{
		Name: "Total Streams In Memory",
		Query: fmt.Sprintf(
			`sum(max_over_time(loki_ingester_memory_streams[%s]))`,
			duration,
		),
		Unit:       StreamsUnit,
		Annotation: IngesterAnnotation,
	}
}
