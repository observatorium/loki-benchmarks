package benchmarks_test

import (
    "sync"
    "time"

    . "github.com/onsi/ginkgo"
    . "github.com/onsi/gomega"

    "github.com/observatorium/loki-benchmarks/internal/config"
    "github.com/observatorium/loki-benchmarks/internal/k8s"
    "github.com/observatorium/loki-benchmarks/internal/latch"
    "github.com/observatorium/loki-benchmarks/internal/logger"
    "github.com/observatorium/loki-benchmarks/internal/metrics"
    "github.com/observatorium/loki-benchmarks/internal/querier"
)

var _ = Describe("Scenario: High Volume Reads", func() {

    var (
        beforeOnce  sync.Once
        afterOnce   sync.Once
        scenarioCfg config.HighVolumeReads
    )

    BeforeEach(func() {
        scenarioCfg = benchCfg.Scenarios.HighVolumeReads

        beforeOnce.Do(func() {
            writerCfg := scenarioCfg.Writers
            readerCfg := scenarioCfg.Readers

            // Deploy the logger to ingest logs
            err := logger.Deploy(k8sClient, benchCfg.Logger, writerCfg, benchCfg.Loki.PushURL())
            Expect(err).Should(Succeed(), "Failed to deploy logger")

            err = k8s.WaitForReadyDeployment(k8sClient, benchCfg.Logger.Namespace, benchCfg.Logger.Name, writerCfg.Replicas, defaultRetry, defaulTimeout)
            Expect(err).Should(Succeed(), "Failed to wait for ready logger deployment")

            // Wait until we ingested enough logs based on startThreshold
            err = latch.WaitUntilGreaterOrEqual(metricsClient, metrics.DistributorBytesReceivedTotal, readerCfg.StartThreshold, defaltLatchTimeout)
            Expect(err).Should(Succeed(), "Failed to wait until latch activated")

            // Undeploy logger to assert only read traffic
            err = logger.Undeploy(k8sClient, benchCfg.Logger)
            Expect(err).Should(Succeed(), "Failed to delete logger deployment")

            // Deploy the query clients
            err = querier.Deploy(k8sClient, benchCfg.Querier, readerCfg, benchCfg.Loki.QueryURL(), readerCfg.Query)
            Expect(err).Should(Succeed(), "Failed to deploy querier")

            err = k8s.WaitForReadyDeployment(k8sClient, benchCfg.Querier.Namespace, benchCfg.Querier.Name, readerCfg.Replicas, defaultRetry, defaulTimeout)
            Expect(err).Should(Succeed(), "Failed to wait for ready querier deployment")
        })

        time.Sleep(scenarioCfg.Samples.Interval)
    })

    AfterEach(func() {
        afterOnce.Do(func() {
            Expect(querier.Undeploy(k8sClient, benchCfg.Querier)).Should(Succeed(), "Failed to delete querier deployment")
        })
    })

    Measure("should result in measurements of p99, p50 and avg for all successful read requests to the query frontend", func(b Benchmarker) {
        job := benchCfg.Metrics.QueryFrontendJob()

        // Record p99 loki_request_duration_seconds_bucket
        p99, err := metricsClient.RequestDurationOkReadsP99(job, "1m")

        Expect(err).Should(Succeed(), "Failed to read p50 for all query frontend reads with status code 2xx")
        Expect(p99).Should(BeNumerically("<", scenarioCfg.P99), "p99 should not exceed expectation")

        b.RecordValue("All query frontend 2xx reads p99", p99)

        // Record p50 loki_request_duration_seconds_bucket
        p50, err := metricsClient.RequestDurationOkReadsP50(job, "1m")

        Expect(err).Should(Succeed(), "Failed to read p50 for all query frontend reads with status code 2xx")
        Expect(p50).Should(BeNumerically("<", scenarioCfg.P50), "p50 should not exceed expectation")

        b.RecordValue("All query frontend 2xx reads p50", p50)

        // Record avg from loki_request_duration_seconds_sum / loki_request_duration_seconds_count
        avg, err := metricsClient.RequestDurationOkReadsAvg(job, "1m")

        Expect(err).Should(Succeed(), "Failed to read average for all query frontend reads with status code 2xx")
        Expect(avg).Should(BeNumerically("<", scenarioCfg.AVG), "avg should not exceed expectation")

        b.RecordValue("All query frontend 2xx reads avg", avg)
    }, 10)
})
