package hw

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
)

type HW struct {
	// memory
	Memory [4096]byte

	// registers

	// data
	V [16]byte
	// address register
	I uint16
	// delay
	DT byte
	// sound
	ST byte

	// pseudo-registers

	// program counter, starts at 0x200
	PC uint16
	// stack counter
	SC uint8

	// stack (starting with virtual stack)
	Stack [16]uint16

	Input byte
}

func InitHW() (HW, error) {

	// first, load HEX digit sprites
	mem, err := loadSprites()
	if err != nil {
		return HW{}, errors.New(fmt.Sprintf("error while loading sprites: %v\n", err))
	}

	hw := HW{
		Memory: mem,
		// initialize at 0x200
		PC: 0x200,
		V:  [16]byte{},
	}

	return hw, nil
}

func (hw *HW) LoadROM(data []byte) {
	// use contiguous mem locations from 0x200
	offset := 0x200
	for i, b := range data {
		hw.Memory[i+offset] = b
	}

}

func (hw *HW) DecreaseTimers() {
	if hw.DT > 0 {
		hw.DT--
	}

	if hw.ST > 0 {
		hw.ST--
	}
}

func loadSprites() ([4096]byte, error) {
	// load digit sprites into memory
	f, err := os.Open("sprites.data")
	if err != nil {
		return [4096]byte{}, err
	}

	scanner := bufio.NewScanner(f)
	if err := scanner.Err(); err != nil {
		return [4096]byte{}, err
	}

	// read sprite data for hex digits and put them in memory
	// use contiguous mem locations from 0x000 to 0x1FF
	mem := [4096]byte{}
	for i := 0; scanner.Scan(); i++ {
		line := scanner.Text()
		b, err := strconv.ParseUint(line, 0, 8)
		if err != nil {
			return [4096]byte{}, err
		}

		mem[i] = byte(b)
	}

	return mem, nil
}
