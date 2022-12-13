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

var _ = Describe("High Volume Writes", func() {
	var highVolumeWriteTest *config.HighVolumeWrites
	var generatorDpl client.Object
	var samplingCfg gmeasure.SamplingConfig
	var samplingRange model.Duration
	var tenant string

	BeforeEach(func() {
		if !benchCfg.Scenarios.IsWriteTestRunnable() {
			Skip("High Volumes Writes Benchmark not enabled")
		}
		highVolumeWriteTest = benchCfg.Scenarios.HighVolumeWrites
	})

	Describe("Forwarding logs to Loki service", func() {
		BeforeEach(func() {
			tenant = benchCfg.Generator.Tenant
			generatorDpl = loadclient.CreateGenerator(highVolumeWriteTest.Writers, benchCfg.Generator)

			err := k8sClient.Create(context.TODO(), generatorDpl, &client.CreateOptions{})
			Expect(err).Should(Succeed(), "Failed to deploy logger")

			err = utils.WaitForReadyDeployment(k8sClient, generatorDpl, defaultRetry, defaultTimeout)
			Expect(err).Should(Succeed(), "Failed to wait for ready logger deployment")

			DeferCleanup(func() {
				err := k8sClient.Delete(context.TODO(), generatorDpl, &client.DeleteOptions{})
				Expect(err).Should(Succeed(), "Failed to delete logger deployment")
			})
		})

		It("can measure performance", func() {
			samplingCfg, samplingRange = highVolumeWriteTest.SamplingConfiguration()

			// Sleeping for the first interval so that the data is accurate for the new workload.
			time.Sleep(samplingCfg.MinSamplingInterval)

			e := gmeasure.NewExperiment(highVolumeWriteTest.Description)
			AddReportEntry(e.Name, e)

			e.Sample(func(idx int) {
				// Distributors
				job := benchCfg.Metrics.Jobs.Distributor
				annotation := metrics.DistributorAnnotation

				// These are confirmation metrics to ensure that the workload matches expectations
				err := metricsClient.MeasureIngestionVerificationMetrics(e, generatorDpl.GetName(), job, tenant, samplingRange)
				Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))

				err = metricsClient.MeasureHTTPRequestMetrics(e, metrics.WriteRequestPath, job, samplingRange, annotation)
				Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))

				// Ingesters
				job = benchCfg.Metrics.Jobs.Ingester
				annotation = metrics.IngesterAnnotation

				err = metricsClient.MeasureGRPCRequestMetrics(e, metrics.WriteRequestPath, job, samplingRange, annotation)
				Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
				err = metricsClient.MeasureResourceUsageMetrics(e, job, samplingRange, annotation)
				Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))

				err = metricsClient.Measure(e, metrics.RequestBoltDBShipperWritesQPS(job, samplingRange))
				Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
			}, samplingCfg)
		})
	})
})
