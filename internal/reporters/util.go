package reporter

import (
  "fmt"
  "strings"
  "path/filepath"

  "github.com/kennygrant/sanitize"
)

func getSubDirectory(measurmentName, directory string) string {
  // The last element in the name is the description of the scenario
  nameComponents := strings.Split(measurmentName, " - ")
  dirName := sanitize.BaseName(nameComponents[len(nameComponents) - 1])

  return filepath.Join(directory, dirName)
}

func createFilePath(measurmentName, directory, extension string) string {
  // The last element in the name is the description of the scenario
  nameComponents := strings.Split(measurmentName, " - ")
  joinedName := strings.Join(nameComponents[:len(nameComponents) - 1], "-")
  fileName := fmt.Sprintf("%s.%s", sanitize.BaseName(joinedName), extension)

  return filepath.Join(directory, fileName)
}
