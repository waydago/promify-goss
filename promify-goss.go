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

var LocalPath string
var FileName string

type Results struct {
	Tested  *[]Tested `json:"results,omitempty"`
	Summary *Summary  `json:"summary,omitempty"`
}

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

type Summary struct {
	FailedCount   int64 `json:"failed-count,omitempty"`
	TestCount     int64 `json:"test-count,omitempty"`
	TotalDuration int64 `json:"total-duration,omitempty"`
}

func CheckRequiredArgs(piped bool, FileName string) {
	if FileName == "nill" || len(FileName) == 0 {
		fmt.Printf("Error: expected a filename to write the .prom file as\n")
		os.Exit(1)
	}
}

func CheckIfPiped() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		log.Println(err)
	}
	if fi.Mode()&os.ModeNamedPipe != 0 {
		return true
	} else {
		return false
	}
}

func LoadPipedData() []byte {
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

func UnmarshalResultsJSON(data []byte) (Results, error) {
	var r Results
	err := json.Unmarshal(data, &r)
	return r, err
}

func FormatPromFriendly(r *Results, f *os.File, t string) error {
	for _, result := range *r.Tested {
		var resourceId string

		fmt.Printf("Resource Type: %v\n", result.ResourceType)
		fmt.Printf("Resource ID: %v\n", result.ResourceID)
		fmt.Printf("Resource Property: %v\n", result.Property)

		switch result.ResourceType {
		case "HTTP":
			re := regexp.MustCompile(`^([a-zA-Z0-9_]+): .*//.*$`)
			match := re.FindStringSubmatch(result.ResourceID)
			if len(match) > 1 {
				resourceId = strings.Split(result.ResourceID, ":")[0]
			} else {
				resourceId = result.ResourceID
			}
		case "Port":
			re := regexp.MustCompile(`^([a-zA-Z0-9_]+): `)
			match := re.FindStringSubmatch(result.ResourceID)
			if len(match) > 1 {
				resourceId = strings.ReplaceAll(result.ResourceID, ": ", "\", port=\"")
			} else {
				resourceId = result.ResourceID
			}

		case "Command":
			commandId := strings.Split(result.ResourceID, "|")
			resourceId = strings.TrimRight(strings.Replace(commandId[0], " -", "", -1), " ")
		case "Process":
			resourceId = strings.ReplaceAll(result.ResourceID, "/", "_")
		default:
			resourceId = result.ResourceID
		}

		f.WriteString(fmt.Sprintf("goss_result_%v{property=\"%v\",resource=\"%v\",skipped=\"%t\"} %v\n",
			strings.ToLower(result.ResourceType), resourceId, result.Property, result.Skipped, result.Result))
		f.WriteString(fmt.Sprintf("goss_result_%v_duration{property=\"%v\",resource=\"%v\",skipped=\"%t\"} %v\n",
			strings.ToLower(result.ResourceType), resourceId, result.Property, result.Skipped, result.Duration))
	}

	f.WriteString(fmt.Sprintf("goss_results_summary{textfile=\"%v\",name=\"tested\"} %v\n", t, r.Summary.TestCount))
	f.WriteString(fmt.Sprintf("goss_results_summary{textfile=\"%v\",name=\"failed\"} %v\n", t, r.Summary.FailedCount))
	f.WriteString(fmt.Sprintf("goss_results_summary{textfile=\"%v\",name=\"duration\"} %v\n", t, r.Summary.TotalDuration))

	return nil
}

func WritePromFileFriendly(r *Results, dotprom string, t string) error {
	f, err := os.Create(dotprom)
	if err != nil {
		return err
	}

	err = FormatPromFriendly(r, f, t)
	if err != nil {
		log.Fatal(err)
	}

	f.Close()

	return nil
}

func main() {

	piped := CheckIfPiped()

	flag.StringVar(&LocalPath, "path", "/var/lib/node_exporter/textfile_collector", "Where to store the .prom file")
	flag.StringVar(&FileName, "name", "", "Name your .prom")

	flag.Parse()

	CheckRequiredArgs(piped, FileName)

	File := fmt.Sprintf("%v/%v", LocalPath, FileName)

	DataPiped := LoadPipedData()

	results, err := UnmarshalResultsJSON(DataPiped)
	if err != nil {
		log.Fatal(err)
	}

	WritePromFileFriendly(&results, File, FileName)

}
