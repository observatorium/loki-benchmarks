package benchmarks_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v2"

	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	k8sconfig "sigs.k8s.io/controller-runtime/pkg/client/config"

	"github.com/observatorium/loki-benchmarks/internal/config"
	"github.com/observatorium/loki-benchmarks/internal/metrics"
)

var (
	benchCfg      *config.Benchmark
	k8sClient     client.Client
	metricsClient metrics.Client

	defaultLatchRange = "5m"

	defaultRetry        = 5 * time.Second
	defaultTimeout      = 60 * time.Second
	defaultLatchTimeout = 5 * time.Minute
)

func init() {
	// Read target environment
	env := os.Getenv("TARGET_ENV")
	if env == "" {
		panic("Missing TARGET_ENV env variable")
	}

	promURL := os.Getenv("PROMETHEUS_URL")
	promToken := os.Getenv("PROMETHEUS_TOKEN")

	// Read config for benchmark tests
	filename := fmt.Sprintf("../config/%s.yaml", env)

	yamlFile, err := os.ReadFile(filename)
	if err != nil {
		panic(fmt.Sprintf("Failed reading benchmark configuration file: %s", filename))
	}

	benchCfg = &config.Benchmark{}

	err = yaml.Unmarshal(yamlFile, benchCfg)
	if err != nil {
		panic(fmt.Sprintf("Failed to marshal benchmark configuration file %s with errors %v", filename, err))
	}

	fmt.Printf("\nUsing benchmark configuration:\n===============================\n%s\n", yamlFile)

	if promURL == "" {
		promURL = benchCfg.Metrics.URL
	}

	// Create a client to collect metrics
	metricsClient, err = metrics.NewClient(promURL, promToken, 10*time.Second)
	if err != nil {
		panic("Failed to create metrics client")
	}

	// Create kubernetes client for deployments
	cfg, err := k8sconfig.GetConfig()
	if err != nil {
		panic("Failed to read kubeconfig")
	}

	mapper, err := apiutil.NewDynamicRESTMapper(cfg)
	if err != nil {
		panic("Failed to create new dynamic REST mapper")
	}

	opts := client.Options{Scheme: scheme.Scheme, Mapper: mapper}

	k8sClient, err = client.New(cfg, opts)
	if err != nil {
		panic("Failed to create new k8s client")
	}
}

func TestBenchmarks(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Benchmarks Suite")
}
