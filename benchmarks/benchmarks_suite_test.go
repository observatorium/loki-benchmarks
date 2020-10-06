package benchmarks_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v2"

	"github.com/observatorium/loki-benchmarks/internal/config"
	"github.com/observatorium/loki-benchmarks/internal/metrics"
	internalreporters "github.com/observatorium/loki-benchmarks/internal/reporters"
)

var (
	benchCfg      *config.Benchmark
	client        config.Client
	metricsClient metrics.Client

	reportDir string

	defaultRetry        = 5 * time.Second
	defaulTimeout       = 30 * time.Second
	defaultLatchTimeout = 5 * time.Minute
)

func init() {
	// Read target environment
	env := os.Getenv("TARGET_ENV")
	if env == "" {
		panic("Missing TARGET_ENV env variable")
	}

	reportDir = os.Getenv("REPORT_DIR")
	if reportDir == "" {
		panic("Missing REPORT_DIR env variable")
	}

	// Read config for benchmark tests
	filename := fmt.Sprintf("../config/%s.yaml", env)
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(fmt.Sprintf("Failed reading benchmark configuration file: %s", filename))
	}

	benchCfg = &config.Benchmark{}
	err = yaml.Unmarshal(yamlFile, benchCfg)
	if err != nil {
		panic("Failed to marshal benchmark configuration file")
	}

	// Create a client to collect metrics
	metricsClient, err = metrics.NewClient(benchCfg.Metrics.URL, 10*time.Second)
	if err != nil {
		panic("Failed to create metrics client")
	}

	// Read target environment to setup the deployer client
	switch cType := os.Getenv("CLIENT_TYPE"); cType {
	case "k8s":
		client, err = config.NewClient("k8s")
		if err != nil {
			panic(err)
		}
	case "docker":
		client, err = config.NewClient("docker")
		if err != nil {
			panic(err)
		}
	default:
		panic("Unknown client type")
	}
}

func TestBenchmarks(t *testing.T) {
	RegisterFailHandler(Fail)

	jr := reporters.NewJUnitReporter(fmt.Sprintf("%s/junit.xml", reportDir))
	csv := internalreporters.NewCsvReporter(reportDir)
	gp := internalreporters.NewGnuplotReporter(reportDir)

	RunSpecsWithDefaultAndCustomReporters(t, "Benchmarks Suite", []Reporter{jr, csv, gp})
}
