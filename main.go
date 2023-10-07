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

  // Open output file for writing
	outFile, outErr = os.Create(*outputFileName)
	if outErr != nil {
		log.Fatalf("Error creating output file: %s", outErr)
	}
	defer outFile.Close()

  pc := 96 //initial program counter value
  for _, line := range lnData {
		instruction := parseMachineCode(line)
		disassembled := disassembleInstruction(instruction, pc)
		fmt.Fprintf(outFile, "%s %s\n", line, disassembled)
		pc += 4 // Assuming each instruction is 4 bytes
	}

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

func disassembleInstruction(instr ARMInstruction, pc int) string {
	switch instr.Opcode {
	case "100010": // ADD
		return fmt.Sprintf("ADD R%d, R%d, R%d", instr.DestReg, instr.SrcReg1, instr.SrcReg2)
	case "110010": // SUB
		return fmt.Sprintf("SUB R%d, R%d, R%d", instr.DestReg, instr.SrcReg1, instr.SrcReg2)
	case "100000": // AND
		return fmt.Sprintf("AND R%d, R%d, R%d", instr.DestReg, instr.SrcReg1, instr.SrcReg2)
	case "101000": // ORR
		return fmt.Sprintf("ORR R%d, R%d, R%d", instr.DestReg, instr.SrcReg1, instr.SrcReg2)
	case "000101": // B
		offset, _ := strconv.ParseInt(instr.Opcode[6:], 2, 32)
		// Adjust for signed value
		offset <<= 2
		target := pc + int(offset)
		return fmt.Sprintf("B #%d", target)
	case "100100": // ADDI
		imm, _ := strconv.ParseInt(instr.Opcode[6:], 2, 32)
		return fmt.Sprintf("ADDI R%d, R%d, #%d", instr.DestReg, instr.SrcReg1, imm)
	case "110100": // SUBI
		imm, _ := strconv.ParseInt(instr.Opcode[6:], 2, 32)
		return fmt.Sprintf("SUBI R%d, R%d, #%d", instr.DestReg, instr.SrcReg1, imm)
	case "111110": // LDUR, STUR
		offset, _ := strconv.ParseInt(instr.Opcode[6:], 2, 32)
		return fmt.Sprintf("%s R%d, [R%d, #%d]", "LDUR", instr.DestReg, instr.SrcReg1, offset)
	case "101101": // CBZ, CBNZ
		offset, _ := strconv.ParseInt(instr.Opcode[6:], 2, 32)
		// Adjust for signed value
		offset <<= 2
		target := pc + int(offset)
		if instr.DestReg&1 == 0 {
			return fmt.Sprintf("CBZ R%d, #%d", instr.DestReg, target)
		} else {
			return fmt.Sprintf("CBNZ R%d, #%d", instr.DestReg, target)
		}
	default:
		return "UNKNOWN INSTRUCTION"
	}
}
