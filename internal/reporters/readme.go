package reporter

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/reporters"
	"github.com/onsi/ginkgo/types"
)

type readmeReporter struct {
	ReportDir string
}

func NewReadmeReporter(reportDir string) reporters.Reporter {
	return &readmeReporter{ReportDir: reportDir}
}

func (cr *readmeReporter) SpecSuiteWillBegin(config config.GinkgoConfigType, summary *types.SuiteSummary) {
}

func (cr *readmeReporter) BeforeSuiteDidRun(setupSummary *types.SetupSummary) {}

func (cr *readmeReporter) SpecWillRun(specSummary *types.SpecSummary) {}

func (cr *readmeReporter) SpecDidComplete(specSummary *types.SpecSummary) {
	if len(specSummary.Measurements) == 0 {
		return
	}

	path := ""
	contents := map[string][]string{}
	contentKeys := []string{}

	for key, value := range specSummary.Measurements {
		if path == "" {
			dirName := getSubDirectory(value.Name, cr.ReportDir)
			path = filepath.Join(dirName, "README.md")
		}

		sanitizedKey := strings.ReplaceAll(key, "_", "-")
		components := strings.Split(sanitizedKey, " - ")
		lokiComponent := strings.Join(strings.Split(components[0], "-"), " ")

		if scenarios := contents[lokiComponent]; scenarios != nil {
			contents[lokiComponent] = append(scenarios, components[1])
		} else {
			contentKeys = append(contentKeys, lokiComponent)
			contents[lokiComponent] = []string{components[1]}
		}
	}

	file, err := os.Create(path)

	if err != nil {
		return
	}
	defer file.Close()

	title := "# Benchmark Report\n\n" +
		"This document contains baseline benchmark results for Loki under synthetic load.\n\n"
	tableOfContents := "## Table of Contents\n\n"
	resultsSection := "---\n\n## Benchmark Results\n\n"

	sort.Strings(contentKeys)

	for _, key := range contentKeys {
		values := contents[key]
		sort.Strings(values)

		displayKey := strings.Title(key)
		markdownKey := strings.Join(strings.Split(key, " "), "-")

		tableOfContents += fmt.Sprintf("- [%s](#component-%s)\n", displayKey, strings.ToLower(markdownKey))
		resultsSection += fmt.Sprintf("### Component: %s\n\n", displayKey)

		for _, value := range values {
			displayValue := strings.Title(value)
			markdownValue := strings.Join(strings.Split(value, " "), "-")

			tableOfContents += fmt.Sprintf("\t- [%s](#%s)\n", displayValue, strings.ToLower(markdownValue))

			imageName := fmt.Sprintf("%s-%s.gnuplot.png", markdownKey, markdownValue)
			resultsSection += fmt.Sprintf("#### %s\n\n", displayValue)
			resultsSection += fmt.Sprintf("![./%s](./%s)\n\n", imageName, imageName)
		}
	}

	_, _ = file.WriteString(title)
	_, _ = file.WriteString(tableOfContents + "\n")
	_, _ = file.WriteString(resultsSection)
}

func (cr *readmeReporter) AfterSuiteDidRun(setupSummary *types.SetupSummary) {}

func (cr *readmeReporter) SpecSuiteDidEnd(summary *types.SuiteSummary) {}
