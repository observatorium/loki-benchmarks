package benchmarks_test

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gmeasure"

	"github.com/observatorium/loki-benchmarks/internal/utils"
	"github.com/observatorium/loki-benchmarks/internal/logger"
	"github.com/observatorium/loki-benchmarks/internal/querier"
)

var _ = Describe("Scenario: High Volume Reads", func() {

	loggerCfg := benchCfg.Logger
	querierCfg := benchCfg.Querier
	scenarioCfgs := benchCfg.Scenarios.HighVolumeReads

	BeforeEach(func() {
		if !scenarioCfgs.Enabled {
			Skip("High Volumes Reads Benchmark not enabled!")
		}
	})

	for _, scenarioCfg := range scenarioCfgs.Configurations {
		scenarioCfg := scenarioCfg

		readerCfg := scenarioCfg.Readers
		writerCfg := scenarioCfg.Writers
		sampleCfg := scenarioCfg.Samples

		defaultRange := sampleCfg.Range

		samplingCfg := gmeasure.SamplingConfig{
			N:                   sampleCfg.Total,
			Duration:            sampleCfg.Interval * time.Duration(sampleCfg.Total+1),
			MinSamplingInterval: sampleCfg.Interval,
		}

		Describe("should measure metrics for configuration", func() {
			BeforeEach(func() {
				err := logger.Deploy(k8sClient, loggerCfg, writerCfg, benchCfg.Loki.PushURL())
				Expect(err).Should(Succeed(), "Failed to deploy logger")

				err = utils.WaitForReadyDeployment(k8sClient, loggerCfg.Namespace, loggerCfg.Name, writerCfg.Replicas, defaultRetry, defaultTimeout)
				Expect(err).Should(Succeed(), "Failed to wait for ready logger deployment")

				// Wait until we ingested enough logs based on startThreshold
				err = utils.WaitUntilReceivedBytes(metricsClient, readerCfg.StartThreshold, defaultLatchRange, defaultRetry, defaultLatchTimeout)
				Expect(err).Should(Succeed(), "Failed to wait until latch activated")

				// Undeploy logger to assert only read traffic
				err = logger.Undeploy(k8sClient, loggerCfg)
				Expect(err).Should(Succeed(), "Failed to delete logger deployment")

				// Deploy the query clients
				for id, query := range readerCfg.Queries {
					err = querier.Deploy(k8sClient, querierCfg, readerCfg, benchCfg.Loki.QueryFrontend, id, query, samplingCfg.Duration)
					Expect(err).Should(Succeed(), "Failed to deploy querier")

					name := querier.DeploymentName(querierCfg, id)

					err = utils.WaitForReadyDeployment(k8sClient, querierCfg.Namespace, name, readerCfg.Replicas, defaultRetry, defaultTimeout)
					Expect(err).Should(Succeed(), fmt.Sprintf("Failed to wait for ready querier deployment: %s", name))
				}

				DeferCleanup(func() {
					for id := range readerCfg.Queries {
						err := querier.Undeploy(k8sClient, querierCfg, id)
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

					err := metricsClient.Measure(e, metricsClient.RequestReadsQPS, "2xx reads QPS", job.QueryLabel, job.Job, defaultRange)
					Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
					err = metricsClient.Measure(e, metricsClient.RequestDurationOkQueryRangeP99, "2xx reads p99", job.QueryLabel, job.Job, defaultRange)
					Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
					err = metricsClient.Measure(e, metricsClient.RequestDurationOkQueryRangeP50, "2xx reads p50", job.QueryLabel, job.Job, defaultRange)
					Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
					err = metricsClient.Measure(e, metricsClient.RequestDurationOkQueryRangeAvg, "2xx reads avg", job.QueryLabel, job.Job, defaultRange)
					Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
					err = metricsClient.Measure(e, metricsClient.RequestQueryRangeThroughput, "2xx reads throughput", job.QueryLabel, job.Job, defaultRange)
					Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))

					// Querier
					if benchCfg.Metrics.EnableCadvisorMetrics {
						job = benchCfg.Metrics.CadvisorQuerierJob()

						err = metricsClient.Measure(e, metricsClient.ContainerUserCPU, "Containers User CPU (Mi/Core)", job.QueryLabel, job.Job, defaultRange)
						Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
						err = metricsClient.Measure(e, metricsClient.ContainerWorkingSetMEM, "Containers WorkingSet memory (MB)", job.QueryLabel, job.Job, defaultRange)
						Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
					}

					job = benchCfg.Metrics.QuerierJob()

					err = metricsClient.Measure(e, metricsClient.ProcessCPU, "Processes CPU", job.QueryLabel, job.Job, defaultRange)
					Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))

					err = metricsClient.Measure(e, metricsClient.RequestReadsQPS, "2xx reads QPS", job.QueryLabel, job.Job, defaultRange)
					Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
					err = metricsClient.Measure(e, metricsClient.RequestDurationOkQueryRangeP99, "2xx reads p99", job.QueryLabel, job.Job, defaultRange)
					Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
					err = metricsClient.Measure(e, metricsClient.RequestDurationOkQueryRangeP50, "2xx reads p50", job.QueryLabel, job.Job, defaultRange)
					Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
					err = metricsClient.Measure(e, metricsClient.RequestDurationOkQueryRangeAvg, "2xx reads avg", job.QueryLabel, job.Job, defaultRange)
					Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
					err = metricsClient.Measure(e, metricsClient.RequestQueryRangeThroughput, "2xx reads throughput", job.QueryLabel, job.Job, defaultRange)
					Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))

					// Ingesters
					if benchCfg.Metrics.EnableCadvisorMetrics {
						job = benchCfg.Metrics.CadvisorIngesterJob()

						err = metricsClient.Measure(e, metricsClient.ContainerUserCPU, "Containers User CPU (Mi/Core)", job.QueryLabel, job.Job, defaultRange)
						Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
						err = metricsClient.Measure(e, metricsClient.ContainerWorkingSetMEM, "Containers WorkingSet memory (MB)", job.QueryLabel, job.Job, defaultRange)
						Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
					}

					job = benchCfg.Metrics.IngesterJob()

					err = metricsClient.Measure(e, metricsClient.ProcessCPU, "Processes CPU", job.QueryLabel, job.Job, defaultRange)
					Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))

					if sampleCfg.Interval > 15*time.Minute {
						err = metricsClient.Measure(e, metricsClient.RequestBoltDBShipperReadsQPS, "Boltdb shipper reads QPS", job.QueryLabel, job.Job, defaultRange)
						Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
					}

					err = metricsClient.Measure(e, metricsClient.RequestReadsGrpcQPS, "successful GRPC reads QPS", job.QueryLabel, job.Job, defaultRange)
					Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
					err = metricsClient.Measure(e, metricsClient.RequestDurationOkGrpcQuerySampleP99, "successful GRPC query p99", job.QueryLabel, job.Job, defaultRange)
					Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
					err = metricsClient.Measure(e, metricsClient.RequestDurationOkGrpcQuerySampleP50, "successful GRPC query p50", job.QueryLabel, job.Job, defaultRange)
					Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
					err = metricsClient.Measure(e, metricsClient.RequestDurationOkGrpcQuerySampleAvg, "successful GRPC query avg", job.QueryLabel, job.Job, defaultRange)
					Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
				}, samplingCfg)
			})
		})
	}
})
