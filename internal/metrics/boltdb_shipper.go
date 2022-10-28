package metrics

import "github.com/prometheus/common/model"

func (c *client) RequestBoltDBShipperReadsQPS(job string, duration model.Duration) (float64, error) {
	return c.requestBoltDBShipperQPS(job, "Shipper.Query", "2.*", duration)
}

func (c *client) RequestBoltDBShipperWritesQPS(job string, duration model.Duration) (float64, error) {
	return c.requestBoltDBShipperQPS(job, "WRITE", "2.*", duration)
}
