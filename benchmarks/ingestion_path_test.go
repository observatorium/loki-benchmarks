package benchmarks_test

import (
	"context"
	"fmt"
	"time"

	"github.com/observatorium/loki-benchmarks/internal/config"
	"github.com/observatorium/loki-benchmarks/internal/loadclient"
	"github.com/observatorium/loki-benchmarks/internal/metrics"
	"github.com/observatorium/loki-benchmarks/internal/utils"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gmeasure"
	"github.com/prometheus/common/model"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Ingestion Path", func() {
	var (
		ingestionTest *config.IngestionPath
		generatorDpl  client.Object
		samplingCfg   gmeasure.SamplingConfig
		samplingRange model.Duration
	)

	BeforeEach(func() {
		if !benchCfg.Scenarios.IsWriteTestEnabled() {
			Skip("Ingestion Path Benchmarks not enabled")
		}
		ingestionTest = benchCfg.Scenarios.IngestionPath
	})

	Describe("Forwarding logs to Loki service", func() {
		BeforeEach(func() {
			generatorDpl = loadclient.CreateGenerator(ingestionTest.Writers, benchCfg.Generator)

			err := k8sClient.Create(context.TODO(), generatorDpl, &client.CreateOptions{})
			Expect(err).Should(Succeed(), "Failed to deploy logger")

			err = utils.WaitForReadyDeployment(k8sClient, generatorDpl, defaultRetry, defaultTimeout)
			Expect(err).Should(Succeed(), "Failed to wait for ready logger deployment")

			DeferCleanup(func() {
				err := k8sClient.Delete(context.TODO(), generatorDpl, &client.DeleteOptions{})
				Expect(err).Should(Succeed(), "Failed to delete logger deployment")
			})
		})

		It("samples metric data from ingestion path related components", func() {
			samplingCfg, samplingRange = ingestionTest.SamplingConfiguration()

			// Sleeping for the first interval so that the data is accurate for the new workload.
			time.Sleep(samplingCfg.MinSamplingInterval)

			e := gmeasure.NewExperiment(ingestionTest.Description)
			AddReportEntry(e.Name, e)

			e.Sample(func(idx int) {
				// Load Generation
				err := metricsClient.MeasureIngestionVerificationMetrics(e, generatorDpl.GetName(), samplingRange)
				Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))

				// Distributors
				job := benchCfg.Metrics.Jobs.Distributor
				annotation := metrics.DistributorAnnotation

				err = metricsClient.MeasureHTTPRequestMetrics(e, metrics.WriteRequestPath, job, samplingRange, annotation)
				Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))

				// Ingesters
				job = benchCfg.Metrics.Jobs.Ingester
				annotation = metrics.IngesterAnnotation

				err = metricsClient.MeasureResourceUsageMetrics(e, job, samplingRange, annotation)
				Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
				err = metricsClient.MeasureVolumeUsageMetrics(e, job, samplingRange, annotation)
				Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
				err = metricsClient.MeasureGRPCRequestMetrics(e, metrics.WriteRequestPath, job, samplingRange, annotation)
				Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
				err = metricsClient.MeasureBoltDBShipperRequestMetrics(e, metrics.WriteRequestPath, job, samplingRange)
				Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
			}, samplingCfg)
		})
	})
})
