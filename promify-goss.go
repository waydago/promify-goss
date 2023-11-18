package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strings"
)

var localPath string
var fileName string

// Results is the struct for the JSON output from goss
type Results struct {
	Tested  *[]Tested `json:"results,omitempty"`
	Summary *Summary  `json:"summary,omitempty"`
}

// Tested is the struct for the JSON output from goss
type Tested struct {
	Duration     int64    `json:"duration,omitempty"`
	Expected     []string `json:"expected,omitempty"`
	Found        []string `json:"found,omitempty"`
	Property     string   `json:"property,omitempty"`
	ResourceID   string   `json:"resource-id,omitempty"`
	ResourceType string   `json:"resource-type,omitempty"`
	Result       int64    `json:"result,omitempty"`
	Skipped      bool     `json:"skipped,omitempty"`
	Successful   bool     `json:"successful,omitempty"`
	TestType     int64    `json:"test-type,omitempty"`
}

// Summary is the struct for the JSON output from goss
type Summary struct {
	FailedCount   int64 `json:"failed-count,omitempty"`
	TestCount     int64 `json:"test-count,omitempty"`
	TotalDuration int64 `json:"total-duration,omitempty"`
}

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func checkRequiredArgs(piped bool, fileName string) {
	if fileName == "nill" || len(fileName) == 0 {
		fmt.Printf("Error: expected a filename to write the .prom file as\n")
		os.Exit(1)
	}
}

func checkIfPiped() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		log.Println(err)
	}
	if fi.Mode()&os.ModeNamedPipe != 0 {
		return true
	}
	return false
}

func loadPipedData() []byte {
	var Buf bytes.Buffer
	reader := bufio.NewReader(os.Stdin)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				Buf.WriteString(line)
				break
			} else {
				fmt.Println(err.Error())
				os.Exit(1)
			}
		}
		Buf.WriteString(line)
	}
	return Buf.Bytes()
}

func unmarshalResultsJSON(data []byte) (Results, error) {
	var r Results
	err := json.Unmarshal(data, &r)
	return r, err
}

func formatPromFriendly(r *Results, f *os.File, t string) error {
	for _, result := range *r.Tested {
		var resourceID string

		switch result.ResourceType {
		case "HTTP":
			re := regexp.MustCompile(`^([a-zA-Z0-9_]+): .*//.*$`)
			match := re.FindStringSubmatch(result.ResourceID)
			if len(match) > 1 {
				resourceID = strings.Split(result.ResourceID, ":")[0]
			} else {
				resourceID = result.ResourceID
			}
		case "Port":
			re := regexp.MustCompile(`^([a-zA-Z0-9_]+): `)
			match := re.FindStringSubmatch(result.ResourceID)
			if len(match) > 1 {
				resourceID = strings.ReplaceAll(result.ResourceID, ": ", "\", port=\"")
			} else {
				resourceID = result.ResourceID
			}
		case "Command":
			commandID := strings.Split(result.ResourceID, "|")
			commandID = strings.Split(commandID[0], " ")
			resourceID = strings.TrimRight(strings.Replace(commandID[0], " -", "", -1), " ")
		case "Process":
			resourceID = strings.ReplaceAll(result.ResourceID, "/", "_")
		default:
			resourceID = result.ResourceID
		}

		_, err := f.WriteString(fmt.Sprintf("goss_result_%v{property=\"%v\",resource=\"%v\",skipped=\"%t\"} %v\n",
			strings.ToLower(result.ResourceType), resourceID, result.Property, result.Skipped, result.Result))
		checkError(err)
		_, err = f.WriteString(fmt.Sprintf("goss_result_%v_duration{property=\"%v\",resource=\"%v\",skipped=\"%t\"} %v\n",
			strings.ToLower(result.ResourceType), resourceID, result.Property, result.Skipped, result.Duration))
		checkError(err)
	}

	_, err := f.WriteString(fmt.Sprintf("goss_results_summary{textfile=\"%v\",name=\"tested\"} %v\n", t, r.Summary.TestCount))
	checkError(err)
	_, err = f.WriteString(fmt.Sprintf("goss_results_summary{textfile=\"%v\",name=\"failed\"} %v\n", t, r.Summary.FailedCount))
	checkError(err)
	_, err = f.WriteString(fmt.Sprintf("goss_results_summary{textfile=\"%v\",name=\"duration\"} %v\n", t, r.Summary.TotalDuration))
	checkError(err)

	return nil
}

func writePromFileFriendly(r *Results, dotprom string, t string) error {
	f, err := os.Create(dotprom)
	if err != nil {
		return err
	}

	err = formatPromFriendly(r, f, t)
	if err != nil {
		log.Fatal(err)
	}

	f.Close()

	return nil
}

func main() {

	piped := checkIfPiped()

	flag.StringVar(&localPath, "path", "/var/lib/node_exporter/textfile_collector", "Where to store the .prom file")
	flag.StringVar(&fileName, "name", "", "Name your .prom")

	flag.Parse()

	checkRequiredArgs(piped, fileName)

	File := fmt.Sprintf("%v/%v", localPath, fileName)

	DataPiped := loadPipedData()

	results, err := unmarshalResultsJSON(DataPiped)
	checkError(err)

	err = writePromFileFriendly(&results, File, fileName)
	checkError(err)

}
