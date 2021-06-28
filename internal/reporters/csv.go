package reporter

import (
	"encoding/csv"
	"fmt"
	"os"
	"time"

	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/reporters"
	"github.com/onsi/ginkgo/types"
)

type csvReporter struct {
	ReportDir string
}

func NewCsvReporter(reportDir string) reporters.Reporter {
	return &csvReporter{ReportDir: reportDir}
}

func (cr *csvReporter) SpecSuiteWillBegin(config config.GinkgoConfigType, summary *types.SuiteSummary) {
}

func (cr *csvReporter) BeforeSuiteDidRun(setupSummary *types.SetupSummary) {}

func (cr *csvReporter) SpecWillRun(specSummary *types.SpecSummary) {}

func (cr *csvReporter) SpecDidComplete(specSummary *types.SpecSummary) {
	for key, value := range specSummary.Measurements {
		dirName := getSubDirectory(value.Name, cr.ReportDir)
		filepath := createFilePath(key, dirName, "csv")

		file, err := os.Create(filepath)
		if err != nil {
			return
		}
		defer file.Close()

		writer := csv.NewWriter(file)
		defer writer.Flush()

		ts := time.Now().Unix()

		var records [][]string

		for _, res := range value.Results {
			values := []string{fmt.Sprintf("%d", ts), fmt.Sprintf("%f", res)}
			records = append(records, values)
			ts++
		}

		for _, record := range records {
			err = writer.Write(record)
			if err != nil {
				continue
			}
		}
	}
}

func (cr *csvReporter) AfterSuiteDidRun(setupSummary *types.SetupSummary) {}

func (cr *csvReporter) SpecSuiteDidEnd(summary *types.SuiteSummary) {}
