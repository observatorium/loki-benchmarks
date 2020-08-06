package benchmarks_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestBenchmarks(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Benchmarks Suite")
}
