package main

import (
	"flag"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

func Test_getFileData(t *testing.T) {
	tests := []struct {
		name    string    // name of test
		want    inputFile // the input file we want the function to return
		wantErr bool      // whether or not we want an error
		osArgs  []string  // the command arguments used for the test
	}{
		// Here we're declaring each unit test input and output data as defined before
		{"Default parameters", inputFile{"test.csv", "comma", false}, false, []string{"cmd", "test.csv"}},
		{"No parameters", inputFile{}, true, []string{"cmd"}},
		{"Semicolon enabled", inputFile{"test.csv", "semicolon", false}, false, []string{"cmd", "--separator=semicolon", "test.csv"}},
		{"Pretty enabled", inputFile{"test.csv", "comma", true}, false, []string{"cmd", "--pretty", "test.csv"}},
		{"Pretty and semicolon enabled", inputFile{"test.csv", "semicolon", true}, false, []string{"cmd", "--pretty", "--separator=semicolon", "test.csv"}},
		{"Separator not identified", inputFile{}, true, []string{"cmd", "--separator=pipe", "test.csv"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualOsArgs := os.Args
			// defer function will run after the function runs
			defer func() {
				os.Args = actualOsArgs                                           // Restoring the original os.Args reference
				flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError) // Reseting the flag command line. So that we can parse flags again
			}()

			os.Args = tt.osArgs
			got, err := getFileData()
			if (err != nil) != tt.wantErr {
				t.Errorf("getFileData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getFileData() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_checkIfValidFile(t *testing.T) {
	// create a temporary and empty csv
	tmpfile, err := ioutil.TempFile("", "*test*.csv")
	if err != nil {
		panic(err)
	}
	defer os.Remove(tmpfile.Name())

	tests := []struct {
		name     string
		filename string
		want     bool
		wantErr  bool
	}{
		{"File does exist", tmpfile.Name(), true, false},
		{"File does not exist", "nowhere/test.csv", false, true},
		{"File is not csv", "test.txt", false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := checkIfValidFile(tt.filename)
			if (err != nil) != tt.wantErr {
				t.Errorf("checkIfValidFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("checkIfValidFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_processCsvFile(t *testing.T) {
	// Defining the maps we're expenting to get from our function
	wantMapSlice := []map[string]string{
		{"COL1": "1", "COL2": "2", "COL3": "3"},
		{"COL1": "4", "COL2": "5", "COL3": "6"},
	}

	tests := []struct {
		name      string
		csvString string // The content of our tested CSV file
		separator string // The separator used for each test case
	}{
		{"Comma separator", "COL1,COL2,COL3\n1,2,3\n4,5,6\n", "comma"},
		{"Semicolon separator", "COL1;COL2;COL3\n1;2;3\n4;5;6\n", "semicolon"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Creating a CSV temp file for testing
			tmpfile, err := ioutil.TempFile("", "test*.csv")
			check(err)

			defer os.Remove(tmpfile.Name())            // Removing the CSV test file before living
			_, err = tmpfile.WriteString(tt.csvString) // Writing the content of the CSV test file
			tmpfile.Sync()                             // Persisting data on disk
			// Defining the inputFile struct that we're going to use as one parameter of our function
			testFileData := inputFile{
				filepath:  tmpfile.Name(),
				pretty:    false,
				separator: tt.separator,
			}
			// Defining the writerChanel
			writerChannel := make(chan map[string]string)
			// Calling the targeted function as a go routine
			go processCsvFile(testFileData, writerChannel)
			// Iterating over the slice containing the expected map values
			for _, wantMap := range wantMapSlice {
				record := <-writerChannel                // Waiting for the record that we want to compare
				if !reflect.DeepEqual(record, wantMap) { // Making the corresponding test assertion
					t.Errorf("processCsvFile() = %v, want %v", record, wantMap)
				}
			}
		})
	}
}
