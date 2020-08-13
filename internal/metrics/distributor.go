package metrics

func (c *client) DistributorBytesReceivedTotal() (float64, error) {
    return c.executeScalarQuery(`loki_distributor_bytes_received_total`)
}
