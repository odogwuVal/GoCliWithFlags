package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	name := flag.String("name", "Valentine", "The name of the passed in user")
	flag.Parse()
	// NArg is the number of arguments passed after the flag
	if flag.NArg() == 0 {
		fmt.Printf("Hello, %s!\n", *name)
	} else if flag.Arg(0) == "list" {
		files, _ := os.Open(".")
		defer files.Close()

		fileInfo, _ := files.Readdir(-1)
		for _, file := range fileInfo {
			fmt.Println(file.Name())
		}
	} else {
		fmt.Println("Check documentation")
	}
}
