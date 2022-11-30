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

func (c *Client) DistributorBytesReceivedTotal(duration model.Duration) (float64, error) {
	// This method breaks the convention of the Measurment -> Measure structure because
	// it is used out of the scope of an experiement.

	query := fmt.Sprintf(
		`sum(max_over_time(loki_distributor_bytes_received_total[%s]) - min_over_time(loki_distributor_bytes_received_total[%s]))`,
		duration, duration,
	)
	return c.executeScalarQuery(query)
}

func DistributorGiPDReceivedTotal(job string, duration model.Duration) Measurement {
	return Measurement{
		Name: "Total Projected Bytes Received",
		Query: fmt.Sprintf(
			`sum(rate(loki_distributor_bytes_received_total{job=~".*%s.*"}[%s])) / %d * %d`,
			job, duration, BytesToGigabytesMultiplier, SecondsPerDay,
		),
		Unit:       GigabytesPerDayUnit,
		Annotation: DistributorAnnotation,
	}
}

func LoadNetworkTotal(job string, duration model.Duration) Measurement {
	return Measurement{
		Name: "Total Bytes Transmitted",
		Query: fmt.Sprintf(
			`sum(rate(container_network_transmit_bytes_total{pod=~"%s-.*"}[%s])) / %d`,
			job, duration, BytesToMegabytesMultiplier,
		),
		Unit:       MegabytesPerSecondUnit,
		Annotation: LoadGeneratorAnnotation,
	}
}

func LoadNetworkGiPDTotal(job string, duration model.Duration) Measurement {
	return Measurement{
		Name: "Total Projected Bytes Transmitted",
		Query: fmt.Sprintf(
			`sum(rate(container_network_transmit_bytes_total{pod=~".*%s.*"}[%s])) / %d * %d`,
			job, duration, BytesToGigabytesMultiplier, SecondsPerDay,
		),
		Unit:       GigabytesPerDayUnit,
		Annotation: LoadGeneratorAnnotation,
	}
}
