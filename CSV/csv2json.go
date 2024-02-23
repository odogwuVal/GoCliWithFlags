package main

import (
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func main() {
	getFileData()
}

type inputFile struct {
	filepath  string
	separator string
	pretty    bool
}

func check(e error) {
	if e != nil {
		exitGracefully(e)
	}
}

func exitGracefully(err error) {
	fmt.Fprintf(os.Stderr, "error: %v\n", err)
	os.Exit(1)
}

func getFileData() (inputFile, error) {
	// validate the correct number of arguments
	if len(os.Args) < 2 {
		return inputFile{}, errors.New("a file path argument is required")
	}

	// Define the option flags
	// this will contain the name of the flag, the default value and a description of the flag
	separator := flag.String("separator", "comma", "column separator")
	pretty := flag.Bool("pretty", false, "Prettify JSON or not")

	flag.Parse()

	fileLocation := flag.Arg(0) // this basically returns the first argument which is not a flag

	// validating the separator we have recieved
	if !(*separator == "comma" || *separator == "semicolon") {
		return inputFile{}, errors.New("separator has to be either comma or semicolon")
	}

	// If everything goes well and we get to this point,
	// we return the corresponding struct instance with all required data
	return inputFile{fileLocation, *separator, *pretty}, nil
}

func checkIfValidFile(filename string) (bool, error) {
	// checking if entered file is CSV by using the filepath package from the standard library
	if fileExtension := filepath.Ext(filename); fileExtension != ".csv" {
		return false, fmt.Errorf("file %s is not CSV", filename)
	}

	// checking if filepath entered belongs to an existing file. We use the stat method from the os package (standard library)
	if _, err := os.Stat(filename); err != nil && os.IsNotExist(err) {
		return false, fmt.Errorf("file %s does not exist", filename)
	}

	// if everything goes well and we get to this point
	return true, nil
}

func processCsvFile(fileData inputFile, writerChannel chan map[string]string) {
	file, err := os.Open(fileData.filepath)
	check(err)
	defer file.Close()

	// Define headers and line slice
	var headers, line []string

	// Initialize the csv reader
	reader := csv.NewReader(file)

	// if the separator supplied from the commandline is semicolon, we need to add it here
	if fileData.separator == "semicolon" {
		reader.Comma = ';'
	}

	// Reading the first line where we will find our headers
	headers, err = reader.Read()
	check(err)

	// Iterate over each line of the CSV file
	for {
		line, err = reader.Read()
		// close the channel if we get to the end of the file

		if err == io.EOF {
			close(writerChannel)
			break
		} else if err != nil {
			exitGracefully(err)
		}
		// Processing a CSV Line
		record, err := processLine(headers, line)
		if err != nil {
			fmt.Printf("Line: %sError: %s\n", line, err)
			continue
		}

		// send the processed record to the channel
		writerChannel <- record
	}
}

func processLine(headers []string, datalist []string) (map[string]string, error) {
	// validating if we are getting the same number of headers and columns, otherwise return an error
	if len(datalist) != len(headers) {
		return nil, errors.New("line does not match headers format. skipping")
	}

	// creating the map we are going to populate
	recordMap := make(map[string]string)
	// for each header, we are going to set a new map key with the corresponding column value
	for i, name := range headers {
		recordMap[name] = datalist[i]
	}

	return recordMap, nil
}
