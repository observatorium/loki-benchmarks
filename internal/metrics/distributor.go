package metrics

import (
	"fmt"

	"github.com/prometheus/common/model"
)

func (c *client) DistributorBytesReceivedTotal(duration model.Duration) (float64, error) {
	query := fmt.Sprintf(
		`sum(max_over_time(loki_distributor_bytes_received_total[%s]) - min_over_time(loki_distributor_bytes_received_total[%s]))`,
		duration, duration,
	)

	return c.executeScalarQuery(query)
}

func (c *client) DistributorGiPDReceivedTotal(job string, duration model.Duration) (float64, error) {
	query := fmt.Sprintf(
		`sum(rate(loki_distributor_bytes_received_total{job=~".*%s.*"}[%s])) / %s * %d`,
		job, duration, BytesToGigabytesMultiplier, SecondsPerDay,
	)

	return c.executeScalarQuery(query)
}
