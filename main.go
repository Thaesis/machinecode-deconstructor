package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
)

// ARMInstruction represents an ARM instruction
type ARMInstruction struct {
	Opcode  string
	DestReg int
	SrcReg1 int
	SrcReg2 int
}

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
	var lnData []string
	// adding scanned lines to the slice
	for scanner.Scan() {
		lnData = append(lnData, scanner.Text())
	}

	inFile.Close()
	outFile.Close()

}

func parseMachineCode(machineCode string) ARMInstruction {
	// Extract fields from the machine code
	opcode := machineCode[:6]
	destReg, _ := strconv.ParseInt(machineCode[6:8], 2, 32)
	srcReg1, _ := strconv.ParseInt(machineCode[8:10], 2, 32)
	srcReg2, _ := strconv.ParseInt(machineCode[10:12], 2, 32)

	// Create and return an ARMInstruction struct
	return ARMInstruction{
		Opcode:  opcode,
		DestReg: int(destReg),
		SrcReg1: int(srcReg1),
		SrcReg2: int(srcReg2),
	}
}
