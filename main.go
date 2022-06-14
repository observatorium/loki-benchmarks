package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

const startingGB = 4250
const startingLogs = 986
const endingGB = 6000
const gbJumps = 250
const numOfRunsPerTest = 3

const defaultStyle = "\x1b[0m"
const cyanColor = "\x1b[36m"
const configFile string = "config/ocp-observatorium-test.yaml"

const filesFormat = "csv"
const fileCpuName = "CpuGraph"
const fileMemName = "MemoryGraph"
const filePathCpu string = "test_results/"
const filePathMem string = "test_results/"

func removeColorsExtentions(num string) string {
	tmp := strings.Split(num, cyanColor)
	tmp = strings.Split(tmp[1], defaultStyle)
	return tmp[0]
}

func changeWordInFile(path string, currWord string, newWord string) {
	input, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalln(err)
	}
	lines := strings.Split(string(input), "\n")

	for i, line := range lines {
		words := strings.Split(string(line), " ")

		for j, word := range words {
			if word == currWord {
				words[j] = newWord
				lines[i] = strings.Join(words, " ")

				newContent := strings.Join(lines, "\n")
				err = ioutil.WriteFile(path, []byte(newContent), 0644)
				if err != nil {
					log.Fatalln(err)
				}
				return
			}
		}
	}
}

func runShellCommand(strCmd string, args ...string) string {
	fmt.Println(cyanColor + strCmd + " " + strings.Join(args, " ") + defaultStyle)
	cmd := exec.Command(strCmd, args...)
	output, _ := cmd.CombinedOutput()
	return string(output)
}

func getGraphVal(str *string, header string) string {
	lines := strings.Split(*str, "\n")
	cpuVal := ""
	for i, line := range lines {
		if strings.Contains(line, header) {
			tmp := strings.Split(lines[i+3], " ")
			cpuVal = removeColorsExtentions(tmp[6])
		}
	}

	return cpuVal
}

func createResultFiles(extentionName string) (*os.File, *os.File) {
	cpuFullPath := filePathCpu + fileCpuName + extentionName + "." + filesFormat
	memFullPath := filePathMem + fileMemName + extentionName + "." + filesFormat

	osCpuFile, err := os.OpenFile(cpuFullPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		fmt.Println("Unable to create file")
		return nil, nil
	}

	osMemFile, err := os.OpenFile(memFullPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		fmt.Println("Unable to create file")
		return nil, nil
	}

	return osCpuFile, osMemFile
}

func main() {
	gbPerDay := startingGB
	logsLineRate := startingLogs
	logsJumps := int(math.Round((float64(startingLogs) / float64(startingGB)) * float64(gbJumps)))

	var cpuFiles [numOfRunsPerTest](*os.File)
	var memFiles [numOfRunsPerTest](*os.File)
	avgCpuFile, avgMemFile := createResultFiles("Avg")

	for i := 0; i < numOfRunsPerTest; i++ {
		cpuFiles[i], memFiles[i] = createResultFiles(strconv.Itoa(i + 1))
	}

	for {
		cpuSum, memSum, giPDSum := float64(0), float64(0), float64(0)

		for i := 0; i < numOfRunsPerTest; i++ {
			fmt.Println(runShellCommand("make", "s3-bucket-cleanup"))
			output := runShellCommand("make", "ocp-loki-bench")
			fmt.Println(output)

			giPD := getGraphVal(&output, "Load Total")
			cpu := getGraphVal(&output, "Processes CPU")
			mem := getGraphVal(&output, "Containers WorkingSet")

			val, err := strconv.ParseFloat(cpu, 64)
			if err != nil {
				fmt.Println("Failed to convert cpu string to float64")
				return
			}
			cpuSum = cpuSum + val

			val, err = strconv.ParseFloat(mem, 64)
			if err != nil {
				fmt.Println("Failed to convert memory string to float64")
				return
			}
			memSum = memSum + val

			val, err = strconv.ParseFloat(giPD, 64)
			if err != nil {
				fmt.Println("Failed to convert cpu string to float64")
				return
			}
			giPDSum = giPDSum + val

			cpuFiles[i].WriteString(giPD + "," + cpu + "\n")
			memFiles[i].WriteString(giPD + "," + mem + "\n")
		}

		avgCpu := float64(cpuSum / (float64(numOfRunsPerTest)))
		avgMem := float64(memSum / (float64(numOfRunsPerTest)))
		avgGiPD := float64(giPDSum / (float64(numOfRunsPerTest)))

		avgCpuFile.WriteString(fmt.Sprint(avgGiPD) + "," + fmt.Sprint(avgCpu) + "\n")
		avgMemFile.WriteString(fmt.Sprint(avgGiPD) + "," + fmt.Sprint(avgMem) + "\n")

		if gbPerDay >= endingGB {
			break
		}

		changeWordInFile(configFile, strconv.Itoa(gbPerDay), strconv.Itoa(gbPerDay+gbJumps))
		changeWordInFile(configFile, strconv.Itoa(logsLineRate), strconv.Itoa(logsLineRate+logsJumps))
		gbPerDay = gbPerDay + gbJumps
		logsLineRate = logsLineRate + logsJumps
	}

	changeWordInFile(configFile, strconv.Itoa(gbPerDay), strconv.Itoa(startingGB))
	changeWordInFile(configFile, strconv.Itoa(logsLineRate), strconv.Itoa(startingLogs))

	for i := 0; i < numOfRunsPerTest; i++ {
		cpuFiles[i].Close()
		memFiles[i].Close()
	}

	avgCpuFile.Close()
	avgMemFile.Close()
}
