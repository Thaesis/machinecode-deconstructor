package main

import (
	"bufio"
	"debug/macho"
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
  Offset int
  Cond int
  Imm int
  Addr int
  Op2 int
  Shamt int
  ShiftCode int
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
  switch {
    
  case "000101" == machineCode[:6]: //B Format | 6-bit
    opcode := "B"
    offset, _ := strconv.ParseInt(machineCode[6:], 2, 32)
    
  case "10110100", "10110101" == machineCode[:9]: //CB Format | 8-bit
    
    if (machineCode[:9] == "10110100") {
      opcode := "CBZ"
    } else {
      opcode := "CBNZ"
    }
    offset, _ := strconv.ParseInt(machineCode[8:27], 2, 32)
    cond, _ := strconv.ParseInt(machineCode[27:], 2, 32)
    
  case "110100101" == machineCode[:9]: //IM Format | 9-bit
    
    opcode := "MOVZ"
    shiftCode, _ := strconv.ParseInt(machineCode[9:11], 2, 32)
    field, _ := strconv.ParseInt(machineCode[11:28], 2, 32)
    destReg, _ := strconv.ParseInt(machineCode[28:], 2, 32)
    
  case "1001000100", "1101000100" == machineCode[:10]: // I Format | 10-bit
    
    if (machineCode[:10] == "1001000100") {
      opcode := "ADDI"
    } else {
      opcode := "SUBI"
    }
    imm, _ := strconv.ParseInt(machineCode[10:23], 2, 32)
    srcReg1, _ := strconv.ParseInt(machineCode[23:28], 2, 32)
    destReg, _ := strconv.ParseInt(machineCode[28:], 2, 32)
    
  case "11111000000", "11111000010" == machineCode[:11]: // D Format | 11-bit
    
    if (machineCode[:11] == "11111000000") q{
      opcode := "STUR"
    } else {
      opcode := "LUDR"
    }
    addr, _ := strconv.ParseInt(machineCode[11:21], 2, 32)
    op2, _ := strconv.ParseInt(machineCode[21:23], 2, 32)
    srcReg1, _ := strconv.ParseInt(machineCode[23:28], 2, 32)
    srcReg2, _ := strconv.ParseInt(machineCode[28:], 2, 32)

  case "11001011000",  //SUB
        "10001011000", //ADD
        "10001010000", //AND
        "10101010000", //ORR
        "11001010000", //EOR
        "11010011011", //LSL
        "11010011010" == machineCode[:12]:

    if (machineCode[:12] == "1100101100") {opcode := "SUB"}
    if (machineCode[:12] == "10001011000") {opcode := "ADD"}
    if (machineCode[:12] == "10001010000") {opcoded := "AND"}
    if (machineCode[:12] == "10101010000") {opcode := "ORR"}
    if (machineCode[:12] == "11001010000") {opcode := "EOR"}
    if (machineCode[:12] == "11010011011") {opcode := "LSL"}
    if (machineCode[:12] == "11010011010") {
      opcode := "LSR"
    } else {opcode := "ASR"}

    srcReg1, _ := strconv.ParseInt(machineCode[13:17], 2, 32)
    shamt, _ := strconv.ParseInt(machineCode[17:23], 2, 32)
    srcReg2, _ := strconv.ParseInt(machineCode[23:28], 2, 32)
    destReg, _ := strconv.ParseInt(machineCode[28:], 2, 32)
    }
  
  return ARMInstruction{
    Opcode: opcode,
    DestReg: int(destReg),
    SrcReg1: int(srcReg1),
    SrcReg2: int(srcReg2),
    Offset: int(offset),
    Cond: int(cond),
    Imm: int(imm),
    Addr: int(addr),
    Op2: int(op2),
    Shamt: int(shamt),
    ShiftCode: int(shiftCode),
    }
}

func disassembleInstruction(instr ARMInstruction, pc int) string {
	switch instr.Opcode {
  
  case "ADD", "SUB", "AND", "ORR", "EOR" :
    retrun fmt.Sprintf("%03d %d R%d, R%d, R%d", pc, instr.Opcode, instr.SrcReg1, instr.SrcReg2)

	case "B": 
		// Adjust for signed value
		instr.Offset <<= 2
		instr.DestReg := pc + int(instr.Offset * 4)
		return fmt.Sprintf("%03d B #%d", pc, instr.DestReg)
    
	case "ADDI", "SUBI": 
		return fmt.Sprintf("%03d %d R%d, R%d, #%d", pc, instr.Opcode, instr.DestReg, instr.SrcReg1, isntr.Imm)
    
	case "LDUR", "STUR": 
    instr.DestReg := pc + int(instr.Offset * 4)
		return fmt.Sprintf("%03d %d R%d, [R%d, #%d]", pc, instr.Opcode, instr.DestReg, instr.SrcReg1, instr.DestReg)
    
	case "CBZ", "CBNZ": 
		// Adjust for signed value
		instr.Offset <<= 2
		instr.DestReg := pc + int(instr.Offset * 4)
		return fmt.Sprintf("%03d %d R%d, #%d", pc, instr.Opcode, instr.DestReg, instr.DestReg)
		
  case "MOVZ": 
    return fmt.Sprintf("%03d %d R%d, #%d, LSL %d", pc, instr.Opcode, instr.DestReg, instr.Imm, instr.ShiftCode)
    
  case "LSR", "LSL":
    return fmt.Sprintf("%03d %d R%d, R%d, #%d", pc, instr.Opcode, instr.DestReg, instr.SrcReg1, instr.Shamt)
    
  case "1111001": //BREAK
    return "BREAK"
	default:
		return "UNKNOWN INSTRUCTION"
	}
}