package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
)

func main() {

	// Flag i/o retrieval
	var inputFileName *string
	var outputFileName *string

	inputFileName = flag.String("i", "", "Retrieves the input file name")
	outputFileName = flag.String("o", "", "Designates the output file name from the flag")

	flag.Parse()

	fmt.Println("Declared input: ", *inputFileName)
	fmt.Println("Declared output: ", *outputFileName)
	fmt.Println(flag.NArg())

	if flag.NArg() != 0 {
		os.Exit(200)
	}

	// Opening input & output file
	inFile, inErr := os.Open(*inputFileName)
	outFile, outErr := os.Create(*outputFileName)

	if inErr != nil {
		log.Fatalf("Unable to open file: %s", inErr)
	}

	if outErr != nil {
		log.Fatalf("Declared outfile either already exists, or could not be created in the local file")
	}

	scanner := bufio.NewScanner(inFile)
	scanner.Split(bufio.ScanLines)
	var txtlines []string
	// adding scanned lines to the slice
	for scanner.Scan() {
		txtlines = append(txtlines, scanner.Text())
	}

	inFile.Close()
	outFile.Close()
	// printing each line to the command line
	for _, eachLine := range txtlines {
		fmt.Println(eachLine)
	}
}
