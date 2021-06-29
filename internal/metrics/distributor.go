package metrics

import (
	"fmt"
	"github.com/prometheus/common/model"
)

func (c *client) DistributorBytesReceivedTotal() (float64, error) {
	query := fmt.Sprintf(
		`(max_over_time(loki_distributor_bytes_received_total[5m]) - min_over_time(loki_distributor_bytes_received_total[5m]))`,
	)
	return c.executeScalarQuery(query)
}

func (c *client) DistributorGiPDReceivedTotal(label, job string, duration model.Duration) (float64, error) {
	query := fmt.Sprintf(
		`(sum(rate(loki_distributor_bytes_received_total{%s=~".*%s.*"}[%s])) / 1000000000 * 86400)`, // in Gi per day
		label, job, duration,
	)

	return c.executeScalarQuery(query)
}
