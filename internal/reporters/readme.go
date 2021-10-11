package reporter

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/kennygrant/sanitize"
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

	contents := map[string][]string{}
	header := ""
	contentKeys := []string{}
	readmePath := filepath.Join(cr.ReportDir, "README.md")
	resultPath := filepath.Join(cr.ReportDir, "result.md")

	for key, value := range specSummary.Measurements {
		if header == "" {
			nameComponents := strings.Split(value.Name, " - ")
			header = sanitize.BaseName(nameComponents[len(nameComponents)-1])
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

	resultsSection := "\n\n---\n\n## " + header + "\n\n"
	tableOfContents := "- " + header + "\n"

	sort.Strings(contentKeys)

	for _, key := range contentKeys {
		values := contents[key]
		sort.Strings(values)

		displayKey := strings.Title(key)
		markdownKey := strings.Join(strings.Split(key, " "), "-")

		tableOfContents += fmt.Sprintf("\t- [%s](#component-%s)\n", displayKey, strings.ToLower(markdownKey))
		resultsSection += fmt.Sprintf("### Component: %s\n\n", displayKey)

		for _, value := range values {
			displayValue := strings.Title(value)
			markdownValue := strings.ToLower(strings.Join(strings.Split(value, " "), "-"))

			tableOfContents += fmt.Sprintf("\t\t- [%s](%s)\n", displayValue, markdownValue)

			imageName := fmt.Sprintf("%s-%s-%s.gnuplot.png", markdownKey, markdownValue, strings.ToLower(header))
			resultsSection += fmt.Sprintf("#### %s\n\n", displayValue)
			resultsSection += fmt.Sprintf("![./%s](./%s)\n\n", imageName, imageName)
		}
	}

	resultFile, err := os.OpenFile(resultPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)

	if err != nil {
		return
	}

	defer resultFile.Close()

	readmeFile, err := os.OpenFile(readmePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return
	}

	defer readmeFile.Close()

	_, _ = readmeFile.WriteString(tableOfContents)
	_, _ = resultFile.WriteString(resultsSection)
}

func (cr *readmeReporter) AfterSuiteDidRun(setupSummary *types.SetupSummary) {}

func (cr *readmeReporter) SpecSuiteDidEnd(summary *types.SuiteSummary) {}
