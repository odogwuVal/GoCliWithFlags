package main

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	// Showing useful information when the user enters the --help option
	flag.Usage = func() {
		fmt.Printf("Usage: %s [options] <csvFile>\nOptions:\n", os.Args[0])
		flag.PrintDefaults()
	}
	// Getting the file data that was entered by the user
	fileData, err := getFileData()

	if err != nil {
		exitGracefully(err)
	}
	// Validating the file entered
	if _, err := checkIfValidFile(fileData.filepath); err != nil {
		exitGracefully(err)
	}
	// Declaring the channels that our go-routines are going to use
	writerChannel := make(chan map[string]string)
	done := make(chan bool)
	// Running both of our go-routines, the first one responsible for reading and the second one for writing
	go processCsvFile(fileData, writerChannel)
	go writeJSONFile(fileData.filepath, writerChannel, done, fileData.pretty)
	// Waiting for the done channel to receive a value, so that we can terminate the programn execution
	<-done
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

func writeJSONFile(csvPath string, writerChannel <-chan map[string]string, done chan<- bool, pretty bool) {
	writeString := createStringWriter(csvPath) // Instanciating a JSON writer function
	jsonFunc, breakLine := getJSONFunc(pretty) // Instanciating the JSON parse function and the breakline character
	// Log for informing
	fmt.Println("Writing JSON file...")
	// Writing the first character of our JSON file. We always start with a "[" since we always generate array of record
	writeString("["+breakLine, false)
	first := true
	for {
		// Waiting for pushed records into our writerChannel
		record, more := <-writerChannel
		if more {
			if !first { // If it's not the first record, we break the line
				writeString(","+breakLine, false)
			} else {
				first = false // If it's the first one, we don't break the line
			}

			jsonData := jsonFunc(record) // Parsing the record into JSON
			writeString(jsonData, false) // Writing the JSON string with our writer function
		} else { // If we get here, it means there aren't more record to parse. So we need to close the file
			writeString(breakLine+"]", true) // Writing the final character and closing the file
			fmt.Println("Completed!")        // Logging that we're done
			done <- true                     // Sending the signal to the main function so it can correctly exit out.
			break                            // Stoping the for-loop
		}
	}
}

func createStringWriter(csvPath string) func(string, bool) {
	jsonDir := filepath.Dir(csvPath)                                                       // Getting the directory where the CSV file is
	jsonName := fmt.Sprintf("%s.json", strings.TrimSuffix(filepath.Base(csvPath), ".csv")) // Declaring the JSON filename, using the CSV file name as base
	finalLocation := filepath.Join(jsonDir, jsonName)                                      // Declaring the JSON file location, using the previous variables as base
	// Opening the JSON file that we want to start writing
	f, err := os.Create(finalLocation)
	check(err)
	// This is the function we want to return, we're going to use it to write the JSON file
	return func(data string, close bool) { // 2 arguments: The piece of text we want to write, and whether or not we should close the file
		_, err := f.WriteString(data) // Writing the data string into the file
		check(err)
		// If close is "true", it means there are no more data left to be written, so we close the file
		if close {
			f.Close()
		}
	}
}

func getJSONFunc(pretty bool) (func(map[string]string) string, string) {
	// Declaring the variables we're going to return at the end
	var jsonFunc func(map[string]string) string
	var breakLine string
	if pretty { //Pretty is enabled, so we should return a well-formatted JSON file (multi-line)
		breakLine = "\n"
		jsonFunc = func(record map[string]string) string {
			jsonData, _ := json.MarshalIndent(record, "   ", "   ") // By doing this we're ensuring the JSON generated is indented and multi-line
			return "   " + string(jsonData)                         // Transforming from binary data to string and adding the indent characets to the front
		}
	} else { // Now pretty is disabled so we should return a compact JSON file (one single line)
		breakLine = "" // It's an empty string because we never break lines when adding a new JSON object
		jsonFunc = func(record map[string]string) string {
			jsonData, _ := json.Marshal(record) // Now we're using the standard Marshal function, which generates JSON without formating
			return string(jsonData)             // Transforming from binary data to string
		}
	}

	return jsonFunc, breakLine // Returning everythinbg
}
