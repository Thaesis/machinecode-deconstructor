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
	Opcode    string
	DestReg   int
	SrcReg1   int
	SrcReg2   int
	Offset    int
	Cond      int
	Imm       int
	Addr      int
	Op2       int
	Shamt     int
	ShiftCode int
	Field     int
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
		pc += 4
	}

}

func parseMachineCode(machineCode string) ARMInstruction {
	/*
	 Parses the fields of the format into the struct. In order of least bit length
	*/
	var opcode string
	var destReg, srcReg1, srcReg2, offset, cond, imm, addr, op2, shamt, shiftCode, field int64

	if "000101" == machineCode[:6] { //B Format | 6-bit
		opcode = "B"
		offset, _ = strconv.ParseInt(machineCode[6:], 2, 32)
	}
	if "10110101" == machineCode[:9] || "10110100" == machineCode[:9] { //CB Format | 8-bit

		if machineCode[:9] == "10110100" {
			opcode = "CBZ"
		} else {
			opcode = "CBNZ"
		}

		offset, _ = strconv.ParseInt(machineCode[8:27], 2, 32)
		cond, _ = strconv.ParseInt(machineCode[27:], 2, 32)

	}
	if "110100101" == machineCode[:9] || "111100101" == machineCode[:9] { //IM Format | 9-bit

		if machineCode[:9] == "110100101" {
			opcode = "MOVZ"
		} else {
			opcode = "MOVK"
		}

		shiftCode, _ = strconv.ParseInt(machineCode[9:11], 2, 32)
		field, _ = strconv.ParseInt(machineCode[11:28], 2, 32)
		destReg, _ = strconv.ParseInt(machineCode[28:], 2, 32)

	}
	if "1101000100" == machineCode[:10] || "100100100" == machineCode[:10] { // I Format | 10-bit

		if machineCode[:10] == "1001000100" {
			opcode = "ADDI"
		} else {
			opcode = "SUBI"
		}

		imm, _ = strconv.ParseInt(machineCode[10:23], 2, 32)
		srcReg1, _ = strconv.ParseInt(machineCode[23:28], 2, 32)
		destReg, _ = strconv.ParseInt(machineCode[28:], 2, 32)

	}
	if "11111000010" == machineCode[:11] || "11111000000" == machineCode[:11] { // D Format | 11-bit

		if machineCode[:11] == "11111000000" {
			opcode = "STUR"
		} else {
			opcode = "LUDR"
		}

		addr, _ = strconv.ParseInt(machineCode[11:21], 2, 32)
		op2, _ = strconv.ParseInt(machineCode[21:23], 2, 32)
		srcReg1, _ = strconv.ParseInt(machineCode[23:28], 2, 32)
		srcReg2, _ = strconv.ParseInt(machineCode[28:], 2, 32)

	}
	if "11001011000" == machineCode[:11] || //SUB
		"10001011000" == machineCode[:11] || //ADD
		"10001010000" == machineCode[:11] || //ANDS
		"10101010000" == machineCode[:11] || //AND
		"11001010000" == machineCode[:11] || //ORR
		"11010011011" == machineCode[:11] || //EOR
		"11101010000" == machineCode[:11] || //LSL
		"11010011010" == machineCode[:11] { //LSR

		if machineCode[:11] == "11001011000" {
			opcode = "SUB"
		} else if machineCode[:11] == "10001011000" {
			opcode = "ADD"
		} else if machineCode[:11] == "10001010000" {
			opcode = "AND"
		} else if machineCode[:11] == "10101010000" {
			opcode = "ORR"
		} else if machineCode[:11] == "11101010000" {
			opcode = "EOR"
		} else if machineCode[:11] == "11010011011" {
			opcode = "LSL"
		} else if machineCode[:11] == "11010011010" {
			opcode = "LSR"
		} else {
			opcode = "ASR"
		}

		srcReg1, _ = strconv.ParseInt(machineCode[11:16], 2, 32)
		shamt, _ = strconv.ParseInt(machineCode[16:22], 2, 32)
		srcReg2, _ = strconv.ParseInt(machineCode[22:27], 2, 32)
		destReg, _ = strconv.ParseInt(machineCode[27:], 2, 32)

	}
	if "11111110110111101111111111100111" == machineCode[0:] {
		opcode = "BREAK"
	}
	if "00000000000000000000000000000000" == machineCode[0:] {
		opcode = "NOP"
	}

	return ARMInstruction{
		Opcode:    opcode,
		DestReg:   int(destReg),
		SrcReg1:   int(srcReg1),
		SrcReg2:   int(srcReg2),
		Offset:    int(offset),
		Cond:      int(cond),
		Imm:       int(imm),
		Addr:      int(addr),
		Op2:       int(op2),
		Shamt:     int(shamt),
		ShiftCode: int(shiftCode),
		Field:     int(field),
	}
}

func disassembleInstruction(instr ARMInstruction, pc int) string {
	switch instr.Opcode {

	case "ADD", "SUB":
		return fmt.Sprintf("%03d %s R%d, R%d, R%d", pc, instr.Opcode, instr.DestReg, instr.SrcReg1, instr.SrcReg2)

	case "AND", "ORR", "EOR":
		return fmt.Sprintf("%03d %s R%d, R%d, R%d", pc, instr.Opcode, instr.DestReg, instr.SrcReg1, instr.SrcReg2)

	case "B":
		// Adjust for signed value
		instr.Offset <<= 2
		instr.DestReg = pc + int(instr.Offset*4)
		return fmt.Sprintf("%03d %s #%d", pc, instr.Opcode, instr.DestReg)

	case "ADDI", "SUBI":
		return fmt.Sprintf("%03d %s R%d, R%d, #%d", pc, instr.Opcode, instr.DestReg, instr.SrcReg1, instr.Imm)

	case "LDUR", "STUR":
		instr.DestReg = pc + int(instr.Offset*4)
		return fmt.Sprintf("%03d %s R%d, [R%d, #%d]", pc, instr.Opcode, instr.DestReg, instr.SrcReg1, instr.DestReg)

	case "CBZ", "CBNZ":
		// Adjust for signed value
		instr.Offset <<= 2
		instr.DestReg = pc + int(instr.Offset*4)
		return fmt.Sprintf("%03d %s R%d, #%d", pc, instr.Opcode, instr.DestReg, instr.DestReg)

	case "MOVZ", "MOVK":
		return fmt.Sprintf("%03d %s R%d, #%d, LSL %d", pc, instr.Opcode, instr.DestReg, instr.Imm, instr.ShiftCode)

	case "LSR", "LSL", "ASR":
		return fmt.Sprintf("%03d %s R%d, R%d, #%d", pc, instr.Opcode, instr.DestReg, instr.SrcReg1, instr.Shamt)

	case "BREAK":
		return fmt.Sprintf("%03d %s", pc, instr.Opcode)
	case "NOP":
		return fmt.Sprintf("%s", instr.Opcode)
	default:
		return "UNKNOWN INSTRUCTION"
	}
}
