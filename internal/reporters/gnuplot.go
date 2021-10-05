package reporter

import (
	"fmt"
	"os"
	"time"

	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/reporters"
	"github.com/onsi/ginkgo/types"
)

type gnuplotReporter struct {
	ReportDir string
}

func NewGnuplotReporter(reportDir string) reporters.Reporter {
	return &gnuplotReporter{ReportDir: reportDir}
}

func (cr *gnuplotReporter) SpecSuiteWillBegin(config config.GinkgoConfigType, summary *types.SuiteSummary) {
}

func (cr *gnuplotReporter) BeforeSuiteDidRun(setupSummary *types.SetupSummary) {}

func (cr *gnuplotReporter) SpecWillRun(specSummary *types.SpecSummary) {}

func (cr *gnuplotReporter) SpecDidComplete(specSummary *types.SpecSummary) {
	for key, value := range specSummary.Measurements {
		filepath := createFilePath(key, cr.ReportDir, "gnuplot")

		file, err := os.Create(filepath)
		if err != nil {
			return
		}
		defer file.Close()

		header := "set grid\n" +
			"set key left top\n" +
			"set xdata time\n" +
			"set timefmt '%s'\n" +
			"set datafile separator ','\n" +
			"offset = 0\n" +
			"t0(x)=(offset=($0==0) ? x : offset, x - offset)\n" +
			"set xtics format '%H:%M:%S' font ',6'\n" +
			"set datafile separator ','\n"

		_, _ = file.WriteString(header)
		_, _ = file.WriteString("$DATA << EOD\n")

		ts := time.Now().Unix()

		var records []string
		for _, res := range value.Results {
			records = append(records, fmt.Sprintf("%d,%f\n", ts, res))
			ts++
		}

		for _, record := range records {
			_, _ = file.WriteString(record)
		}

		_, _ = file.WriteString("EOD\n")

		plot := fmt.Sprintf("plot $DATA using (t0(timecolumn(1))*60):2 with lines lw 1 title '%s'\n", value.Name)
		_, _ = file.WriteString(plot)
	}
}

func (cr *gnuplotReporter) AfterSuiteDidRun(setupSummary *types.SetupSummary) {}

func (cr *gnuplotReporter) SpecSuiteDidEnd(summary *types.SuiteSummary) {}
