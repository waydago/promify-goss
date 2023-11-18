package main

import (
	"flag"
	"os"
	"strings"
	"testing"
)

// Fake results struct for testing
var sampleResults = Results{
	Tested: &[]Tested{
		{
			Duration:     123,
			Expected:     []string{"true"},
			Found:        []string{"true"},
			Property:     "exists",
			ResourceID:   "/test",
			ResourceType: "File",
			Result:       0,
			Skipped:      false,
			Successful:   true,
			TestType:     0,
		},
	},
	Summary: &Summary{
		FailedCount:   0,
		TestCount:     1,
		TotalDuration: 123,
	},
}

// TestCheckIfPiped tests the checkIfPiped function
func TestCheckIfPiped(t *testing.T) {
	oldStdin := os.Stdin
	defer func() {
		os.Stdin = oldStdin
	}()

	t.Run("With named pipe", func(t *testing.T) {
		r, w, _ := os.Pipe()
		os.Stdin = r
		go func() {
			defer w.Close()
		}()

		if !checkIfPiped() {
			t.Errorf("Expected true, but got false")
		}
	})

	t.Run("Without named pipe", func(t *testing.T) {
		os.Stdin = oldStdin
		if checkIfPiped() {
			t.Errorf("Expected false, but got true")
		}
	})
}

func parseFlags() {
	flag.StringVar(&localPath, "path", "", "Path to the text file")
	flag.StringVar(&fileName, "name", "", "Name of the prom file")
	flag.Parse()
}

// TestMainFlags tests the main function with flags
func TestMainFlags(t *testing.T) {
	origArgs := os.Args
	defer func() { os.Args = origArgs }()

	os.Args = []string{"cmd", "-path", "./test", "-name", "test_name"}

	parseFlags()

	if localPath != "./test" {
		t.Errorf("Expected path './test', but got '%s'", localPath)
	}
	if fileName != "test_name" {
		t.Errorf("Expected name 'test_name', but got '%s'", fileName)
	}
}

// TestUnmarshalResultsJSON tests the unmarshalResultsJSON function
func TestUnmarshalResultsJSON(t *testing.T) {
	// Provide a sample JSON input
	sampleJSON := `{"results":[{"duration":123,"expected":["true"],"found":["true"],"property":"exists","resource-id":"/test","resource-type":"File","result":0,"skipped":false,"successful":true,"test-type":0}],"summary":{"failed-count":0,"test-count":1,"total-duration":123}}`
	_, err := unmarshalResultsJSON([]byte(sampleJSON))
	if err != nil {
		t.Errorf("unmarshalResultsJSON failed: %v", err)
	}
}

// TestWritePromFileFriendly tests the writePromFileFriendly function
func TestWritePromFileFriendly(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	err = writePromFileFriendly(&sampleResults, tmpfile.Name(), "test_output")
	if err != nil {
		t.Errorf("writePromFileFriendly failed: %v", err)
	}

	content, err := os.ReadFile(tmpfile.Name())
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(content), "goss_result_file") {
		t.Errorf("Expected string 'goss_result_file' not found in output")
	}
}

func TestFormatPromFriendly(t *testing.T) {
	type args struct {
		r *Results
		f *os.File
		t string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := formatPromFriendly(tt.args.r, tt.args.f, tt.args.t); (err != nil) != tt.wantErr {
				t.Errorf("formatPromFriendly() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
