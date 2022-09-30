package benchmarks_test

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gmeasure"

	"github.com/observatorium/loki-benchmarks/internal/loadclient"
	"github.com/observatorium/loki-benchmarks/internal/utils"
)

var _ = Describe("Scenario: High Volume Reads", func() {

	loggerCfg := benchCfg.Logger
	querierCfg := benchCfg.Querier
	scenarioCfgs := benchCfg.Scenarios.HighVolumeReads

	BeforeEach(func() {
		if !scenarioCfgs.Enabled {
			Skip("High Volumes Reads Benchmark not enabled!")
		}

		generatorCfg := loadclient.GeneratorConfig(scenarioCfg.Generator, loggerCfg, benchCfg.Loki.PushURL())

		err := loadclient.CreateDeployment(k8sClient, generatorCfg)
		Expect(err).Should(Succeed(), "Failed to deploy logger")

		err = utils.WaitForReadyDeployment(k8sClient, loggerCfg.Namespace, loggerCfg.Name, generatorCfg.Replicas, defaultRetry, defaultTimeout)
		Expect(err).Should(Succeed(), "Failed to wait for ready logger deployment")

		err = utils.WaitUntilReceivedBytes(metricsClient, scenarioCfgs.StartThreshold, defaultLatchRange, defaultRetry, defaultLatchTimeout)
		Expect(err).Should(Succeed(), "Failed to wait until latch activated")

		err = loadclient.DeleteDeployment(k8sClient, generatorCfg.Name, generatorCfg.Namespace)
		Expect(err).Should(Succeed(), "Failed to delete logger deployment")
	})

	for _, scenarioCfg := range scenarioCfgs.Configurations {
		scenarioCfg := scenarioCfg
		sampleCfg := scenarioCfg.Samples

		defaultRange := sampleCfg.Range
		samplingCfg := gmeasure.SamplingConfig{
			N:                   sampleCfg.Total,
			Duration:            sampleCfg.Interval * time.Duration(sampleCfg.Total+1),
			MinSamplingInterval: sampleCfg.Interval,
		}

		var querierCfgs []loadclient.DeploymentConfig
		for id, query := range scenarioCfg.Readers.Queries {
			querierCfgs = append(querierCfgs, loadclient.QuerierConfig(scenarioCfg.Readers, querierCfg, benchCfg.Loki.QueryFrontend, query, id))
		}

		Describe(fmt.Sprintf("Configuration: %s", scenarioCfg.Description), func() {
			BeforeEach(func() {
				for _, cfg := range querierCfgs {
					err = loadclient.CreateDeployment(k8sClient, cfg)
					Expect(err).Should(Succeed(), "Failed to deploy querier")

					err = utils.WaitForReadyDeployment(k8sClient, cfg.Namespace, cfg.Name, cfg.Replicas, defaultRetry, defaultTimeout)
					Expect(err).Should(Succeed(), fmt.Sprintf("Failed to wait for ready querier deployment: %s", cfg.Name))
				}

				DeferCleanup(func() {
					for _, cfg := range querierCfgs {
						err := loadclient.DeleteDeployment(k8sClient, cfg.Name, cfg.Namespace)
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
