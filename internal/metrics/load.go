package metrics

import (
	"fmt"

	"github.com/prometheus/common/model"
)

func (c *client) LoadNetworkTotal(job string, duration model.Duration) (float64, error) {
	query := fmt.Sprintf(
		`sum(rate(container_network_transmit_bytes_total{pod=~"%s-.*"}[%s])) / %s`,
		job, duration, BytesToMegabytesMultiplier,
	)

	return c.executeScalarQuery(query)
}

func (c *client) LoadNetworkGiPDTotal(job string, duration model.Duration) (float64, error) {
	query := fmt.Sprintf(
		`sum(rate(container_network_transmit_bytes_total{pod=~"%s-.*"}[%s])) / %s * %d`,
		job, duration, BytesToGigabytesMultiplier, SecondsPerDay,
	)

	return c.executeScalarQuery(query)
}
