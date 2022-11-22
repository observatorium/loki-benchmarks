package benchmarks_test

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gmeasure"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/observatorium/loki-benchmarks/internal/loadclient"
	"github.com/observatorium/loki-benchmarks/internal/querier"
	"github.com/observatorium/loki-benchmarks/internal/utils"
)

var _ = Describe("Scenario: High Volume Reads", func() {

	loggerCfg := benchCfg.Logger
	querierCfg := benchCfg.Querier

	scenarioCfgs := benchCfg.Scenarios.HighVolumeReads

	samplingRange := scenarioCfgs.Samples.Range
	samplingCfg := gmeasure.SamplingConfig{
		N:                   scenarioCfgs.Samples.Total,
		Duration:            scenarioCfgs.Samples.Interval * time.Duration(scenarioCfgs.Samples.Total+1),
		MinSamplingInterval: scenarioCfgs.Samples.Interval,
	}

	willQueryForBoltDBShipperMetrics := samplingCfg.Duration >= (15 * time.Minute)

	BeforeEach(func() {
		if !scenarioCfgs.Enabled {
			Skip("High Volumes Reads Benchmark not enabled!")
		}

		generatorCfg := loadclient.GeneratorConfig(scenarioCfgs.Generator, loggerCfg, benchCfg.Loki.PushURL())

		err := loadclient.CreateDeployment(k8sClient, generatorCfg)
		Expect(err).Should(Succeed(), "Failed to deploy logger")

		err = utils.WaitForReadyDeployment(k8sClient, loggerCfg.Namespace, loggerCfg.Name, defaultRetry, defaultTimeout)
		Expect(err).Should(Succeed(), "Failed to wait for ready logger deployment")

		err = utils.WaitUntilReceivedBytes(metricsClient, scenarioCfgs.StartThreshold, defaultLatchRange, defaultRetry, defaultLatchTimeout)
		Expect(err).Should(Succeed(), "Failed to wait until latch activated")

		err = loadclient.DeleteDeployment(k8sClient, generatorCfg.Name, generatorCfg.Namespace)
		Expect(err).Should(Succeed(), "Failed to delete logger deployment")
	})

	for _, scenarioCfg := range scenarioCfgs.Configurations {
		scenarioCfg := scenarioCfg
		querierDpls := querier.CreateQueriers(scenarioCfg.Readers, querierCfg, benchCfg.Loki.QueryFrontend, scenarioCfg.Readers.Queries)

		Describe(fmt.Sprintf("Configuration: %s", scenarioCfg.Description), func() {
			BeforeEach(func() {
				for _, dpl := range querierDpls {
					err := k8sClient.Create(context.TODO(), dpl, &client.CreateOptions{})
					Expect(err).Should(Succeed(), "Failed to deploy querier")

					err = utils.WaitForReadyDeployment(k8sClient, querierCfg.Namespace, dpl.GetName(), defaultRetry, defaultTimeout)
					Expect(err).Should(Succeed(), fmt.Sprintf("Failed to wait for ready querier deployment: %s", dpl.GetName()))
				}

				DeferCleanup(func() {
					for _, dpl := range querierDpls {
						err := k8sClient.Delete(context.TODO(), dpl, &client.DeleteOptions{})
						Expect(err).Should(Succeed(), "Failed to delete querier deployment")
					}
				})
			})

			It("should measure metrics", func() {
				e := gmeasure.NewExperiment(scenarioCfg.Description)
				AddReportEntry(e.Name, e)

				e.Sample(func(idx int) {
					// Query Frontend
					job := benchCfg.Metrics.QueryFrontendJob()

					err := metricsClient.Measure(e, metricsClient.RequestReadsQPS, "2xx reads QPS", job, samplingRange)
					Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
					err = metricsClient.Measure(e, metricsClient.RequestDurationOkQueryRangeP99, "2xx reads p99", job, samplingRange)
					Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
					err = metricsClient.Measure(e, metricsClient.RequestDurationOkQueryRangeP50, "2xx reads p50", job, samplingRange)
					Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
					err = metricsClient.Measure(e, metricsClient.RequestDurationOkQueryRangeAvg, "2xx reads avg", job, samplingRange)
					Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
					err = metricsClient.Measure(e, metricsClient.RequestQueryRangeThroughput, "2xx reads throughput", job, samplingRange)
					Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))

					// Querier
					job = benchCfg.Metrics.QuerierJob()

					err = metricsClient.Measure(e, metricsClient.ProcessCPU, "Processes CPU", job, samplingRange)
					Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
					err = metricsClient.Measure(e, metricsClient.RequestReadsQPS, "2xx reads QPS", job, samplingRange)
					Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
					err = metricsClient.Measure(e, metricsClient.RequestDurationOkQueryRangeP99, "2xx reads p99", job, samplingRange)
					Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
					err = metricsClient.Measure(e, metricsClient.RequestDurationOkQueryRangeP50, "2xx reads p50", job, samplingRange)
					Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
					err = metricsClient.Measure(e, metricsClient.RequestDurationOkQueryRangeAvg, "2xx reads avg", job, samplingRange)
					Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
					err = metricsClient.Measure(e, metricsClient.RequestQueryRangeThroughput, "2xx reads throughput", job, samplingRange)
					Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))

					if benchCfg.Metrics.EnableCadvisorMetrics {
						job = benchCfg.Metrics.CadvisorQuerierJob()

						err = metricsClient.Measure(e, metricsClient.ContainerUserCPU, "Containers User CPU (Mi/Core)", job, samplingRange)
						Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
						err = metricsClient.Measure(e, metricsClient.ContainerWorkingSetMEM, "Containers WorkingSet memory (MB)", job, samplingRange)
						Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
					}

					// Ingesters
					job = benchCfg.Metrics.IngesterJob()

					err = metricsClient.Measure(e, metricsClient.ProcessCPU, "Processes CPU", job, samplingRange)
					Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
					err = metricsClient.Measure(e, metricsClient.RequestReadsGrpcQPS, "successful GRPC reads QPS", job, samplingRange)
					Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
					err = metricsClient.Measure(e, metricsClient.RequestDurationOkGrpcQuerySampleP99, "successful GRPC query p99", job, samplingRange)
					Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
					err = metricsClient.Measure(e, metricsClient.RequestDurationOkGrpcQuerySampleP50, "successful GRPC query p50", job, samplingRange)
					Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
					err = metricsClient.Measure(e, metricsClient.RequestDurationOkGrpcQuerySampleAvg, "successful GRPC query avg", job, samplingRange)
					Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))

					if benchCfg.Metrics.EnableCadvisorMetrics {
						job = benchCfg.Metrics.CadvisorIngesterJob()

						err = metricsClient.Measure(e, metricsClient.ContainerUserCPU, "Containers User CPU (Mi/Core)", job, samplingRange)
						Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
						err = metricsClient.Measure(e, metricsClient.ContainerWorkingSetMEM, "Containers WorkingSet memory (MB)", job, samplingRange)
						Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
					}

					if willQueryForBoltDBShipperMetrics {
						err = metricsClient.Measure(e, metricsClient.RequestBoltDBShipperReadsQPS, "Boltdb shipper reads QPS", job, samplingRange)
						Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
					}
				}, samplingCfg)
			})
		})
	}
})
