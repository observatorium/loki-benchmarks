package benchmarks_test

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gmeasure"

	"github.com/observatorium/loki-benchmarks/internal/k8s"
	"github.com/observatorium/loki-benchmarks/internal/logger"
)

var _ = Describe("Scenario: High Volume Writes", func() {

	loggerCfg := benchCfg.Logger
	scenarioCfgs := benchCfg.Scenarios.HighVolumeWrites

	BeforeEach(func() {
		if !scenarioCfgs.Enabled {
			Skip("High Volumes Writes Benchmark not enabled!")
		}
	})

	for _, scenarioCfg := range scenarioCfgs.Configurations {
		scenarioCfg := scenarioCfg

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

				err = k8s.WaitForReadyDeployment(k8sClient, loggerCfg.Namespace, loggerCfg.Name, writerCfg.Replicas, defaultRetry, defaultTimeout)
				Expect(err).Should(Succeed(), "Failed to wait for ready logger deployment")

				DeferCleanup(func() {
					err := logger.Undeploy(k8sClient, benchCfg.Logger)
					Expect(err).Should(Succeed(), "Failed to delete logger deployment")
				})
			})

			It("should measure metrics", func() {
				e := gmeasure.NewExperiment(scenarioCfg.Description)
				AddReportEntry(e.Name, e)

				e.Sample(func(idx int) {
					// Distributor
					job := benchCfg.Metrics.DistributorJob()

					err := metricsClient.Measure(e, metricsClient.RequestWritesQPS, "2xx push (Req/s)", job.QueryLabel, job.Job, defaultRange)
					Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
					err = metricsClient.Measure(e, metricsClient.RequestDurationOkPushP99, "2xx push p99", job.QueryLabel, job.Job, defaultRange)
					Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
					err = metricsClient.Measure(e, metricsClient.RequestDurationOkPushP50, "2xx push p50", job.QueryLabel, job.Job, defaultRange)
					Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
					err = metricsClient.Measure(e, metricsClient.RequestDurationOkPushAvg, "2xx push avg", job.QueryLabel, job.Job, defaultRange)
					Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
					err = metricsClient.Measure(e, metricsClient.DistributorGiPDReceivedTotal, "Received Total (Gi/Day)", job.QueryLabel, job.Job, defaultRange)
					Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
					err = metricsClient.Measure(e, metricsClient.DistributorGiPDDiscardedTotal, "Discarded Total (Gi/Day)", job.QueryLabel, job.Job, defaultRange)
					Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))

					// Network Load
					err = metricsClient.Measure(e, metricsClient.LoadNetworkTotal, "Load Total (MB/s)", job.QueryLabel, job.Job, defaultRange)
					Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
					err = metricsClient.Measure(e, metricsClient.LoadNetworkGiPDTotal, "Load Total (Gi/Day)", job.QueryLabel, job.Job, defaultRange)
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

					err = metricsClient.Measure(e, metricsClient.ProcessCPU, "Processes CPU (Mi/Core)", job.QueryLabel, job.Job, defaultRange)
					Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))

					if sampleCfg.Interval > 15*time.Minute {
						err = metricsClient.Measure(e, metricsClient.RequestBoltDBShipperWritesQPS, "Boltdb shipper successful writes QPS", job.QueryLabel, job.Job, defaultRange)
						Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
					}

					err = metricsClient.Measure(e, metricsClient.RequestWritesGrpcQPS, "successful GRPC push QPS", job.QueryLabel, job.Job, defaultRange)
					Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
					err = metricsClient.Measure(e, metricsClient.RequestDurationOkGrpcPushP99, "successful GRPC push p99", job.QueryLabel, job.Job, defaultRange)
					Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
					err = metricsClient.Measure(e, metricsClient.RequestDurationOkGrpcPushP50, "successful GRPC push p50", job.QueryLabel, job.Job, defaultRange)
					Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
					err = metricsClient.Measure(e, metricsClient.RequestDurationOkGrpcPushAvg, "successful GRPC push avg", job.QueryLabel, job.Job, defaultRange)
					Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
				}, samplingCfg)
			})
		})
	}
})
