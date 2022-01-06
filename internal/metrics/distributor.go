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

func (c *client) DistributorGiPDReceivedTotal(label, job string, duration model.Duration) (float64, error) {
	query := fmt.Sprintf(
		`sum(rate(loki_distributor_bytes_received_total{%s=~".*%s.*"}[%s])) / %s * 86400`, // in Gi per day
		label, job, duration, BytesToGigabytesMultiplier,
	)

	return c.executeScalarQuery(query)
}

func (c *client) DistributorGiPDDiscardedTotal(label, job string, duration model.Duration) (float64, error) {
	query := fmt.Sprintf(
		`sum(rate(loki_distributor_discarded_bytes_total{%s=~".*%s.*"}[%s])) / %s * 86400`, // in Gi per day
		label, job, duration, BytesToGigabytesMultiplier,
	)

	return c.executeScalarQuery(query)
}
