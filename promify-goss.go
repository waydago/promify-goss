package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

var GossURI string
var TextFilePath string
var PromFileName string

type TestResults struct {
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

func LoadArgs(piped bool, GossURI string, FileName string) {}

func IncludeUsageInError() {
	//fmt.Printf("Usage:\n")
	flag.PrintDefaults()
	os.Exit(1)
}

func CheckRequiredArgs(piped bool, GossURI string, FileName string) {
	if !piped && len(GossURI) == 0 {
		fmt.Printf("Error: expected a goss uri\n")
		IncludeUsageInError()
		os.Exit(1)
	}
	if FileName == "nill" || len(FileName) == 0  {
		fmt.Printf("Error: expected a filename to write the .prom file as\n")
		IncludeUsageInError()
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
		// return (CalledByPipe)
	} else {
		return false
		// return (CalledByPipe)
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


func GetResultsJSON(url string) ([]byte, error) {
	c := &http.Client{}
	r, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("accept", "application/json")

	resp, err := c.Do(r)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer resp.Body.Close()

	results, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return results, nil
}

func UnmarshalResultsJSON(data []byte) (TestResults, error) {
	var r TestResults
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *TestResults) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func (r *Tested) String() string {
	s, err := json.Marshal(r)
	if err != nil {
		log.Fatal(err)
	}
	return string(s)
}

func FormatPromFriendly(r *TestResults, f *os.File, t string) error {
	for _, result := range *r.Tested {
		var resourceId string

		switch result.ResourceType {
		case "Addr":
			resourceId = result.ResourceID
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

func WritePromFileFriendly(r *TestResults, dotprom string, t string) error {
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
	
	LoadArgs(piped, GossURI, PromFileName)

	if !piped {
		flag.StringVar(&GossURI, "uri", "", "Goss endpoint to fetch data from.")
	}

	flag.StringVar(&TextFilePath, "path", "/var/lib/node_exporter/textfile_collector", "Where to store the .prom file")
	flag.StringVar(&PromFileName, "name", "", "Name your .prom")

	flag.Parse()

	CheckRequiredArgs(piped, GossURI, PromFileName)

	File := fmt.Sprintf("%v/%v", TextFilePath, PromFileName)

	if piped {

		DataPiped := LoadPipedData()

		results, err := UnmarshalResultsJSON(DataPiped)
		if err != nil {
			log.Fatal(err)
		}

		WritePromFileFriendly(&results, File, PromFileName)

	} else {

		Response, err := GetResultsJSON(GossURI)
		if err != nil {
			log.Fatal(err)
		} else if Response == nil {
			log.Fatal("No response from the server!")
		}

		results, err := UnmarshalResultsJSON(Response)
		if err != nil {
			log.Fatal(err)
		}

		WritePromFileFriendly(&results, File, PromFileName)

	}

}
