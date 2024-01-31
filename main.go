package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// ARMInstruction represents an ARM instruction
// as a struct.
type ARMInstruction struct {
	Opcode               string
	DestReg              int
	SrcReg1              int
	SrcReg2              int
	Offset               int
	Cond                 int
	Imm                  int
	Addr                 int
	Op2                  int
	Shamt                int
	ShiftCode            int
	Field                int
	FormattedMachineCode string // New field to store the machine code with spaces
}

var breakEncountered = false // flag to indicate BREAK
var breakEncounteredSim1 = false
var breakEncounteredSim2 = false
var isData = false //If theres data

func main() {
	// Flag i/o retrieval
	inputFileName := flag.String("i", "", "Retrieves the input file name")
	outputFileName := flag.String("o", "", "Designates the output file name from the flag")
	var pc = 96     // Assembly Program Counter
	var pcPipe = 96 // For pipeline
	flag.Parse()

	if flag.NArg() != 0 {
		os.Exit(200)
	}

	// Opening input & output file
	inFile, inErr := os.Open(*inputFileName)
	outFileAssem, outErrAssem := os.Create(getOutputFileNameAssem(*outputFileName))

	if inErr != nil {
		log.Fatalf("Unable to open file: %s", inErr)
	}
	if outErrAssem != nil {
		log.Fatalf("Declared outfile either already exists, or could not be created in the local file")
	}
	defer inFile.Close()
	defer outFileAssem.Close()

	// Create .sim output file
	outFileSim, outErrSim := os.Create(getOutputFileNameSim(*outputFileName))
	if outErrSim != nil {
		log.Fatalf("Unable to create simulation output file: %s", outErrSim)
	}
	defer outFileSim.Close()

	// Initialize the registers array with 32 integers
	registers := make([]int, 32)

	// Reading input file
	scanner := bufio.NewScanner(inFile)
	scanner.Split(bufio.ScanLines)
	var lnData []string
	var cycle int = 1
	for scanner.Scan() {
		lnData = append(lnData, scanner.Text())
	}

	// Main Loop
	var dataValue int64
	var breakEncountered bool
	var dataMap, initDataLine = loadData(*inputFileName)

	// Writing to outFileAssem and creating instruction map
	pcToInstructionMap := make(map[int]int) // Maps PC to line index
	for index, line := range lnData {
		instruction := parseMachineCode(line)

		if breakEncountered {
			if instruction.Opcode == "BREAK" {
				fmt.Fprintf(outFileAssem, "%s     %d  BREAK\n", line, pc)
				breakEncountered = false
			} else {
				dataValue, _ = strconv.ParseInt(line, 2, 33)
				if dataValue&0x80000000 != 0 {
					dataValue |= ^0x7FFFFFFF
				}
				fmt.Fprintf(outFileAssem, "%s     %d  %d\n", line, pc, dataValue)
			}
		} else {
			disassembled := disassembleInstruction(instruction, pc)
			fmt.Fprintf(outFileAssem, "%s %s\n", instruction.FormattedMachineCode, disassembled)

			if instruction.Opcode == "BREAK" {
				breakEncountered = true
			}
		}
		pcToInstructionMap[pc] = index
		// Increment the program counter
		pc += 4
	}
	outFileAssem.Close() //Close the file after writing

	if len(dataMap) != 0 {
		isData = true
	}

	for i := 0; i < len(lnData); i++ {
		line := lnData[i]
		instruction := parseMachineCode(line)

		disassembledPipe := disassembleInstruction(instruction, pcPipe)
		// Handle different instructions for simulation
		switch instruction.Opcode {
		case "BREAK":
			if breakEncounteredSim1 == false {
				breakEncounteredSim1 = true
			} else {
				breakEncounteredSim1 = false
				breakEncounteredSim2 = false
			}
		case "ADDI":
			registers[instruction.DestReg] = registers[instruction.SrcReg1] + instruction.Imm
		case "SUBI":
			registers[instruction.DestReg] = registers[instruction.SrcReg1] - instruction.Imm
		case "STUR":
			memoryAddress := registers[instruction.SrcReg1] + (instruction.Imm * 4)
			delete(dataMap, memoryAddress)
			dataMap[memoryAddress] = registers[instruction.DestReg]
		case "LDUR":
			// Calculate the memory address
			memoryAddress := registers[instruction.SrcReg1] + (instruction.Imm * 4)
			if val, exists := dataMap[memoryAddress]; exists {
				registers[instruction.DestReg] = val
			}
		case "ADD":
			registers[instruction.DestReg] = registers[instruction.SrcReg1] + registers[instruction.SrcReg2]
		case "SUB":
			registers[instruction.DestReg] = registers[instruction.SrcReg1] - registers[instruction.SrcReg2]
		case "AND":
			registers[instruction.DestReg] = registers[instruction.SrcReg1] & registers[instruction.SrcReg2]
		case "ORR":
			registers[instruction.DestReg] = registers[instruction.SrcReg1] | registers[instruction.SrcReg2]
		case "EOR":
			registers[instruction.DestReg] = registers[instruction.SrcReg1] ^ registers[instruction.SrcReg2]
		case "LSL":
			registers[instruction.DestReg] = int(uint32(registers[instruction.SrcReg2]) << uint32(instruction.Shamt))
		case "LSR":
			registers[instruction.DestReg] = int(uint32(registers[instruction.SrcReg2]) >> uint32(instruction.Shamt))
		case "ASR":
			registers[instruction.DestReg] = int(int32(registers[instruction.SrcReg2]) >> uint32(instruction.Shamt))
		case "CBZ":
			if registers[instruction.Cond] == 0 {
				newPC := pcPipe + instruction.Offset*4 - 4
				if newIndex, ok := pcToInstructionMap[newPC]; ok {
					i = newIndex
					pcPipe = newPC
				}
			}
		case "CBNZ":
			if registers[instruction.Cond] != 0 {
				newPC := pcPipe + instruction.Offset*4 - 4
				if newIndex, ok := pcToInstructionMap[newPC]; ok {
					i = newIndex
					pcPipe = newPC
				}
			}
		case "B":
			newPC := pcPipe + instruction.Offset*4 - 4
			if newIndex, ok := pcToInstructionMap[newPC]; ok {
				i = newIndex
				pcPipe = newPC
			}
		}
		if breakEncounteredSim2 == true {
			break
		}

		fmt.Fprintf(outFileSim, "====================\n")
		fmt.Fprintf(outFileSim, "cycle:%d\t%s\n", cycle, disassembledPipe)
		fmt.Fprintf(outFileSim, "registers:\n")
		for i := 0; i < 32; i += 8 {
			fmt.Fprintf(outFileSim, "r%02d:\t", i)
			for j := 0; j < 8; j++ {
				fmt.Fprintf(outFileSim, "%d\t", registers[i+j])
			}
			fmt.Fprintf(outFileSim, "\n")
		}
		const maxDataPerLine = 8
		fmt.Fprintf(outFileSim, "data:\n")
		if isData {
			for i := 0; i < len(dataMap); i += maxDataPerLine {
				pcValue := initDataLine + i*4 // Calculate the PC value for the chunk
				fmt.Fprintf(outFileSim, "%d:\t", pcValue)

				// Iterate through a full line of maxDataPerLine entries
				for j := 0; j < maxDataPerLine; j++ {
					dataIndex := initDataLine + (i+j)*4

					// Check if the data exists in the map, otherwise print zero
					if dataValue, exists := dataMap[dataIndex]; exists {
						fmt.Fprintf(outFileSim, "%d\t", dataValue)
					} else {
						fmt.Fprintf(outFileSim, "0\t") // Print zero if data does not exist
					}
				}
				fmt.Fprintf(outFileSim, "\n")
			}
		}
		pcPipe += 4
		cycle++
		fmt.Fprintf(outFileSim, "\n")

		if breakEncounteredSim1 == true {
			breakEncounteredSim2 = true
		}
	}

}

func loadData(inputFileName string) (dMap map[int]int, initDataLine int) {
	dataMap := make(map[int]int)
	var dataFlag = false

	line := 0
	lineOffset := 0
	file, _ := os.Open(inputFileName)
	sc := bufio.NewScanner(file)
	sc.Split(bufio.ScanLines)

	for sc.Scan() {
		line++

		if dataFlag == true && sc.Text() != "11111110110111101111111111100111" {
			lineOffset++
			lineData, _ := strconv.ParseInt(sc.Text(), 2, 33)

			if lineData&0x80000000 != 0 {
				lineData |= ^0x7FFFFFFF
			}

			data := int(lineData)
			dataMap[(line*4)+92] = data

		}

		if sc.Text() == "11111110110111101111111111100111" {
			dataFlag = !dataFlag
			initDataLine = (line * 4) + 96

		}

	}
	file.Close()
	if dataFlag == true {
		return dataMap, initDataLine
	} else {
		return dataMap, initDataLine - (lineOffset * 4) - 4
	}

}

// ////////////////////////////////////////
// Parameter: outputFileName string
// Function to generate the assembly output
// file destination. This file will hold all
// data related to the output of the machine
// code disassembler.
// Return: String of appended output file name
// //////////////////////////////////////////
func getOutputFileNameAssem(outputFileName string) string {
	base := filepath.Base(outputFileName)
	ext := filepath.Ext(base)
	name := base[:len(base)-len(ext)] + "_dis.txt"
	return filepath.Join(filepath.Dir(outputFileName), name)
}

// ////////////////////////////////////////
// Function: getOutputFileNameSim
// Parameter: outputFileName
// Function to generate the simulation output
// file destination. This file will hold all
// data related to the output of the machine
// simulation.
// Return: String of appended output file name
// //////////////////////////////////////////
func getOutputFileNameSim(outputFileName string) string {
	base := filepath.Base(outputFileName)
	ext := filepath.Ext(base)
	name := base[:len(base)-len(ext)] + "_sim.txt"
	return filepath.Join(filepath.Dir(outputFileName), name)
}

// //////////////////////////////////////////
// Function: parseMachineCode
// Parameter: string of machine code binary
// Desc: This function serves to dissasemb-
// le and interpret the machine code binary
// provided to the function. This data is
// then stored into a struct for later use.
// The machine binary is disassembled based
// on a least-bit basis. I.e. the instruction
// with the least length opcode will be chec-
// ed first, and so on.
// The return is an ARMInstruction struct.
// //////////////////////////////////////////
func parseMachineCode(machineCode string) ARMInstruction {

	var opcode string
	var destReg, srcReg1, srcReg2, offset, cond, imm, addr, op2, shamt, shiftCode, field int64

	if "000101" == machineCode[:6] { //B Format | 6-bit
		opcode = "B"
		offset, _ = strconv.ParseInt(machineCode[6:], 2, 32)

		if offset&0x2000000 != 0 {
			offset |= ^0x1FFFFFF
		}

	} else if "10110101" == machineCode[:8] || "10110100" == machineCode[:8] { //CB Format | 8-bit

		if machineCode[:8] == "10110100" {
			opcode = "CBZ"
		} else {
			opcode = "CBNZ"
		}

		offset, _ = strconv.ParseInt(machineCode[8:27], 2, 32)

		if offset&0x40000 != 0 {
			offset |= ^0x1FFFF
		}
		cond, _ = strconv.ParseInt(machineCode[27:], 2, 32)

	} else if "110100101" == machineCode[:9] || "111100101" == machineCode[:9] { //IM Format | 9-bit

		if machineCode[:9] == "110100101" {
			opcode = "MOVZ"
		} else {
			opcode = "MOVK"
		}

		shiftCode, _ = strconv.ParseInt(machineCode[9:11], 2, 32)
		field, _ = strconv.ParseInt(machineCode[11:27], 2, 32)
		destReg, _ = strconv.ParseInt(machineCode[27:], 2, 32)

	} else if "1101000100" == machineCode[:10] || "1001000100" == machineCode[:10] { // I Format | 10-bit

		if machineCode[:10] == "1001000100" {
			opcode = "ADDI"
		} else {
			opcode = "SUBI"
		}
		imm, _ = strconv.ParseInt(machineCode[10:22], 2, 32)
		srcReg1, _ = strconv.ParseInt(machineCode[22:27], 2, 32)
		destReg, _ = strconv.ParseInt(machineCode[27:], 2, 32)

	} else if "11111000010" == machineCode[:11] || "11111000000" == machineCode[:11] { // D Format | 11-bit

		if machineCode[:11] == "11111000000" {
			opcode = "STUR"
		} else {
			opcode = "LDUR"
		}

		imm, _ = strconv.ParseInt(machineCode[11:20], 2, 32)
		op2, _ = strconv.ParseInt(machineCode[20:22], 2, 32)
		srcReg1, _ = strconv.ParseInt(machineCode[22:27], 2, 32)
		destReg, _ = strconv.ParseInt(machineCode[27:], 2, 32)

	} else if "11001011000" == machineCode[:11] || //SUB
		"10001011000" == machineCode[:11] || //ADD
		"10001010000" == machineCode[:11] || //ANDS
		"10101010000" == machineCode[:11] || //AND
		"11001010000" == machineCode[:11] || //ORR
		"11010011011" == machineCode[:11] || //EOR
		"11101010000" == machineCode[:11] || //LSL
		"11010011100" == machineCode[:11] || //ASR
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
		} else if machineCode[:11] == "11010011100" {
			opcode = "ASR"
		}

		srcReg1, _ = strconv.ParseInt(machineCode[11:16], 2, 32)
		shamt, _ = strconv.ParseInt(machineCode[16:22], 2, 32)
		srcReg2, _ = strconv.ParseInt(machineCode[22:27], 2, 32)
		destReg, _ = strconv.ParseInt(machineCode[27:], 2, 32)

	} else if "11111110110111101111111111100111" == machineCode[0:] {
		opcode = "BREAK"
	}
	if "00000000000000000000000000000000" == machineCode[0:] {
		opcode = "NOP"
	}

	// Initialize the variable to hold the formatted machine code
	var formattedMachineCode string
	var builder strings.Builder
	switch {
	case opcode == "B":
		// For B instructions, add a space after the opcode
		builder.WriteString(machineCode[:6] + " " + machineCode[6:])

	case opcode == "CBZ" || opcode == "CBNZ":
		// For CB instructions, add a space after the opcode and condition
		builder.WriteString(machineCode[:8] + " " + machineCode[8:27] + " " + machineCode[27:])

	case opcode == "MOVZ" || opcode == "MOVK":
		// For IM instructions, add a space after the opcode, shiftCode, and destination register
		builder.WriteString(machineCode[:9] + " " + machineCode[9:11] + " " + machineCode[11:27] + " " + machineCode[27:])

	case opcode == "ADDI" || opcode == "SUBI":
		// For I instructions, add a space after the opcode, immediate value, and source register
		builder.WriteString(machineCode[:10] + " " + machineCode[10:22] + " " + machineCode[22:27] + " " + machineCode[27:])

	case opcode == "STUR" || opcode == "LDUR":
		// For D instructions, add a space after the opcode, immediate value, op2, and source register
		builder.WriteString(machineCode[:11] + " " + machineCode[11:20] + " " + machineCode[20:22] + " " + machineCode[22:27] + " " + machineCode[27:])

	case opcode == "SUB" || opcode == "ADD" || opcode == "AND" || opcode == "ORR" || opcode == "EOR" || opcode == "LSL" || opcode == "LSR" || opcode == "ASR":
		// For R instructions, add a space after the opcode, source registers, and destination register
		builder.WriteString(machineCode[:11] + " " + machineCode[11:16] + " " + machineCode[16:22] + " " + machineCode[22:27] + " " + machineCode[27:])

	default:
		// If the opcode is not recognized, don't format the machine code
		builder.WriteString(machineCode)
	}

	formattedMachineCode = builder.String()

	// Return the ARMInstruction with the formatted machine code
	return ARMInstruction{
		// ... [existing fields] ...
		Opcode:               opcode,
		DestReg:              int(destReg),
		SrcReg1:              int(srcReg1),
		SrcReg2:              int(srcReg2),
		Offset:               int(offset),
		Cond:                 int(cond),
		Imm:                  int(imm),
		Addr:                 int(addr),
		Op2:                  int(op2),
		Shamt:                int(shamt),
		ShiftCode:            int(shiftCode),
		Field:                int(field),
		FormattedMachineCode: formattedMachineCode,
	}
}

// //////////////////////////////////////////
// Function: dissassembleInstruction
// Parameter(s): ARMInstruction and an inte-
// ger representation of the PC starting
// counter value.
// Return: String of translated machine code
// binary into LegV8 (ARM) assembly instruc-
// tions.
// //////////////////////////////////////////
func disassembleInstruction(instr ARMInstruction, pc int) string {
	switch instr.Opcode {

	case "ADD", "SUB":
		return fmt.Sprintf("%03d\t%s\t\tR%d, R%d, R%d", pc, instr.Opcode, instr.DestReg, instr.SrcReg2, instr.SrcReg1)

	case "AND", "ORR", "EOR":
		return fmt.Sprintf("%03d\t%s\t\tR%d, R%d, R%d", pc, instr.Opcode, instr.DestReg, instr.SrcReg2, instr.SrcReg1)

	case "B":
		return fmt.Sprintf("   %03d\t%s\t\t#%d", pc, instr.Opcode, instr.Offset)

	case "ADDI", "SUBI":
		return fmt.Sprintf(" %03d\t%s\tR%d, R%d, #%d", pc, instr.Opcode, instr.DestReg, instr.SrcReg1, instr.Imm)

	case "LDUR", "STUR":
		return fmt.Sprintf("%03d\t%s\tR%d, [R%d, #%d]", pc, instr.Opcode, instr.DestReg, instr.SrcReg1, instr.Imm)

	case "CBZ", "CBNZ":
		return fmt.Sprintf("  %03d\t%s\t\tR%d, #%d", pc, instr.Opcode, instr.Cond, instr.Offset)

	case "MOVZ", "MOVK":
		return fmt.Sprintf("%03d\t%s\tR%d, %d, LSL %d", pc, instr.Opcode, instr.DestReg, instr.Field, instr.ShiftCode*16)

	case "LSR", "LSL", "ASR":
		return fmt.Sprintf("%03d\t%s\t\tR%d, R%d, #%d", pc, instr.Opcode, instr.DestReg, instr.SrcReg2, instr.Shamt)

	case "BREAK":
		return fmt.Sprintf("    %03d\t%s\t", pc, instr.Opcode)

	case "NOP":
		return fmt.Sprintf("\t%s", instr.Opcode)

	default:
		return "UNKNOWN INSTRUCTION"
	}
}
