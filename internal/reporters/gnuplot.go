package reporter

import (
    "fmt"
    "os"
    "time"

    "github.com/onsi/ginkgo/config"
    "github.com/onsi/ginkgo/reporters"
    "github.com/onsi/ginkgo/types"

    "github.com/kennygrant/sanitize"
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
        filename := sanitize.BaseName(key)
        filepath := fmt.Sprintf("%s/%s.gnuplot", cr.ReportDir, filename)

        file, err := os.Create(filepath)
        if err != nil {
            return
        }
        defer file.Close()

        header := "set grid\n" +
            "set key left top\n" +
            "set xdata time\n" +
            "set timefmt '%s'\n" +
            "set datafile separator ','\n"

        file.WriteString(header)
        file.WriteString("$DATA << EOD\n")

        ts := time.Now().Unix()

        var records []string
        for _, res := range value.Results {
            records = append(records, fmt.Sprintf("%d,%f\n", ts, res))
            ts = ts + 1
        }

        for _, record := range records {
            file.WriteString(record)
        }
        file.WriteString("EOD\n")

        plot := fmt.Sprintf("plot $DATA using 1:2 with lines lw 1 title '%s'\n", value.Name)
        file.WriteString(plot)
    }
}

func (cr *gnuplotReporter) AfterSuiteDidRun(setupSummary *types.SetupSummary) {}

func (cr *gnuplotReporter) SpecSuiteDidEnd(summary *types.SuiteSummary) {}