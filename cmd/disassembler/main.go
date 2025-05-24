package main

import (
	"chipgo/interpreter"
	"chipgo/utils"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "unrecognized command arguments, usage: disassembler [file]\n")
		os.Exit(1)
	}

	fileName := os.Args[1]
	data, err := utils.ReadBinary(fileName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error while reading file %s\n", fileName)
		os.Exit(1)
	}

	var high byte
	for i, b := range data {
		if i%2 == 0 {
			high = b
		} else {
			instruction := interpreter.BuildInstruction(high, b, uint16(i-1+0x200))
			fmt.Printf("0x%04x:\t0x%04x\t%s\n", instruction.Counter, instruction.Hex, instruction.Decompiled)
		}

	}
}
