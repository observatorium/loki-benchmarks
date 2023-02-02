package benchmarks_test

import (
	"context"
	"fmt"
	"time"

	"github.com/observatorium/loki-benchmarks/internal/config"
	"github.com/observatorium/loki-benchmarks/internal/loadclient"
	"github.com/observatorium/loki-benchmarks/internal/metrics"
	"github.com/observatorium/loki-benchmarks/internal/querier"
	"github.com/observatorium/loki-benchmarks/internal/utils"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gmeasure"
	"github.com/prometheus/common/model"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Query Path", func() {
	var (
		queryTest     *config.QueryPath
		generatorDpl  client.Object
		querierDpls   []client.Object
		samplingCfg   gmeasure.SamplingConfig
		samplingRange model.Duration
	)

	BeforeEach(func() {
		if !benchCfg.Scenarios.IsReadTestEnabled() {
			Skip("Query Path Benchmarks not enabled")
		}
		queryTest = benchCfg.Scenarios.QueryPath
	})

	Describe("Querying logs from Loki service", func() {
		BeforeEach(func() {
			querierDpls = querier.CreateQueriers(queryTest.Readers, benchCfg.Querier)
			generatorDpl = loadclient.CreateGenerator(queryTest.LogGenerator(), benchCfg.Generator)

			err := k8sClient.Create(context.TODO(), generatorDpl, &client.CreateOptions{})
			Expect(err).Should(Succeed(), "Failed to deploy logger")

			err = utils.WaitForReadyDeployment(k8sClient, generatorDpl, defaultRetry, defaultTimeout)
			Expect(err).Should(Succeed(), "Failed to wait for ready logger deployment")

			// Begin loading data into the Loki service so there is something to query for.
			time.Sleep(time.Minute * 5)

			for _, dpl := range querierDpls {
				err := k8sClient.Create(context.TODO(), dpl, &client.CreateOptions{})
				Expect(err).Should(Succeed(), "Failed to deploy querier")

				err = utils.WaitForReadyDeployment(k8sClient, dpl, defaultRetry, defaultTimeout)
				Expect(err).Should(Succeed(), fmt.Sprintf("Failed to wait for ready querier deployment: %s", dpl.GetName()))
			}

			DeferCleanup(func() {
				err = k8sClient.Delete(context.TODO(), generatorDpl, &client.DeleteOptions{})
				Expect(err).Should(Succeed(), "Failed to delete logger deployment")

				for _, dpl := range querierDpls {
					err := k8sClient.Delete(context.TODO(), dpl, &client.DeleteOptions{})
					Expect(err).Should(Succeed(), "Failed to delete querier deployment")
				}
			})
		})

		It("samples metric data from query path related components", func() {
			samplingCfg, samplingRange = queryTest.SamplingConfiguration()

			// Sleeping for the first interval so that the data is accurate for the new workload.
			time.Sleep(samplingCfg.MinSamplingInterval)

			e := gmeasure.NewExperiment(queryTest.Description)
			AddReportEntry(e.Name, e)

			e.Sample(func(idx int) {
				// Load Generation
				err := metricsClient.MeasureLoadQuerierMetrics(e, samplingRange)
				Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
				err = metricsClient.MeasureIngestionVerificationMetrics(e, generatorDpl.GetName(), samplingRange)
				Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))

				// Query Frontend
				job := benchCfg.Metrics.Jobs.QueryFrontend
				annotation := metrics.QueryFrontendAnnotation

				err = metricsClient.MeasureResourceUsageMetrics(e, job, samplingRange, annotation)
				Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
				err = metricsClient.MeasureHTTPRequestMetrics(e, metrics.ReadRequestPath, job, samplingRange, annotation)
				Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
				err = metricsClient.MeasureQueryMetrics(e, job, samplingRange, annotation)
				Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))

				// Querier
				job = benchCfg.Metrics.Jobs.Querier
				annotation = metrics.QuerierAnnotation

				err = metricsClient.MeasureResourceUsageMetrics(e, job, samplingRange, annotation)
				Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
				err = metricsClient.MeasureHTTPRequestMetrics(e, metrics.ReadRequestPath, job, samplingRange, annotation)
				Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
				err = metricsClient.MeasureQueryMetrics(e, job, samplingRange, annotation)
				Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))

				// Index Gateway
				job = benchCfg.Metrics.Jobs.IndexGateway
				annotation = metrics.IndexGatewayAnnotation

				err = metricsClient.MeasureResourceUsageMetrics(e, job, samplingRange, annotation)
				Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
				err = metricsClient.MeasureVolumeUsageMetrics(e, job, samplingRange, annotation)
				Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))

				// Ingesters
				job = benchCfg.Metrics.Jobs.Ingester
				annotation = metrics.IngesterAnnotation

				err = metricsClient.MeasureResourceUsageMetrics(e, job, samplingRange, annotation)
				Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
				err = metricsClient.MeasureGRPCRequestMetrics(e, metrics.ReadRequestPath, job, samplingRange, annotation)
				Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
				err = metricsClient.MeasureBoltDBShipperRequestMetrics(e, metrics.ReadRequestPath, job, samplingRange)
				Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
			}, samplingCfg)
		})
	})
})
