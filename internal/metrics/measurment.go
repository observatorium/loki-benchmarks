package metrics

import (
	"github.com/onsi/gomega/gmeasure"
)

const (
	CoresToMillicores               int = 1000
	SecondsToMillisecondsMultiplier int = 1000
	BytesToMegabytesMultiplier      int = 1000 * 1000
	BytesToGigabytesMultiplier      int = BytesToMegabytesMultiplier * 1000

	GigabytesUnit    = gmeasure.Units("GB")
	MillicoresUnit   = gmeasure.Units("m")
	MillisecondsUnit = gmeasure.Units("ms")

	RequestsPerSecondUnit = gmeasure.Units("requests per second")

	DistributorAnnotation   = gmeasure.Annotation("distributor")
	IngesterAnnotation      = gmeasure.Annotation("ingester")
	QuerierAnnotation       = gmeasure.Annotation("querier")
	QueryFrontendAnnotation = gmeasure.Annotation("query-frontend")
)

type Measurement struct {
	Name       string
	Query      string
	Unit       gmeasure.Units
	Annotation gmeasure.Annotation
}
