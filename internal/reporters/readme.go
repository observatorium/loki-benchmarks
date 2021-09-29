package reporter

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kennygrant/sanitize"
	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/reporters"
	"github.com/onsi/ginkgo/types"
)

type readmeReporter struct {
	ReportDir       string
	tableOfContents string
	resultsSection  string
	isDone          bool
}

func NewReadmeReporter(reportDir string) reporters.Reporter {
	return &readmeReporter{ReportDir: reportDir, tableOfContents: "", resultsSection: "", isDone: false}
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
	header := ""
	for key, value := range specSummary.Measurements {
		if path == "" {
			nameComponents := strings.Split(value.Name, " - ")
			header = sanitize.BaseName(nameComponents[len(nameComponents)-1])
			path = filepath.Join(cr.ReportDir, "README.md")
		}

		components := strings.Split(key, " - ")
		lokiComponent := strings.Join(strings.Split(components[0], "-"), " ")

		if scenarios := contents[lokiComponent]; scenarios != nil {
			contents[lokiComponent] = append(scenarios, components[1])
		} else {
			contents[lokiComponent] = []string{components[1]}
		}
	}

	cr.resultsSection += "\n\n---\n\n## " + header + "\n\n"
	cr.tableOfContents += "- " + header + "\n"

	for key, values := range contents {
		displayKey := strings.Title(key)
		markdownKey := strings.Join(strings.Split(key, " "), "-")

		cr.tableOfContents += fmt.Sprintf("\t- [%s](#component-%s)\n", displayKey, markdownKey)
		cr.resultsSection += fmt.Sprintf("### Component: %s\n\n", displayKey)

		for _, value := range values {
			displayValue := strings.Title(value)
			markdownValue := strings.Join(strings.Split(value, " "), "-")

			cr.tableOfContents += fmt.Sprintf("\t\t- [%s](%s)\n", displayValue, markdownValue)

			imageName := fmt.Sprintf("%s-%s.gnuplot.png", markdownKey, markdownValue)
			cr.resultsSection += fmt.Sprintf("#### %s\n\n", displayValue)
			cr.resultsSection += fmt.Sprintf("![./%s](./%s)\n\n", imageName, imageName)
		}
	}

	if !cr.isDone {
		cr.isDone = true
		return
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)

	if err != nil {
		return
	}

	defer file.Close()

	_, _ = file.WriteString(cr.tableOfContents)
	_, _ = file.WriteString(cr.resultsSection)
}

func (cr *readmeReporter) AfterSuiteDidRun(setupSummary *types.SetupSummary) {}

func (cr *readmeReporter) SpecSuiteDidEnd(summary *types.SuiteSummary) {}
