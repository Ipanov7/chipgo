package main

import (
	"chipgo/internal/hw"
	"chipgo/internal/interpreter"
	"chipgo/internal/ui"
	"chipgo/internal/utils"
	"fmt"
	"os"
	"time"
)

const (
	HZ  = 500
	FPS = 60
)

func main() {
	// load ROM
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "wrong command format. usage: chipgo [file] (--debug)")
		os.Exit(1)
	}

	fileName := os.Args[1]
	isDebug := false
	if len(os.Args) == 3 {
		isDebug = os.Args[2] == "--debug"
	}

	// init UI

	ui, err := ui.Init(isDebug)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	defer ui.Stop()

	// init hardware
	hw, err := hw.InitHW()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	data, err := utils.ReadBinary(fileName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	hw.LoadROM(data)

	// load prgram in ui debugger
	for i := 0x200; i < len(hw.Memory); i += 2 {
		instruction := interpreter.BuildInstruction(hw.Memory[i], hw.Memory[i+1], uint16(i))
		ui.Instructions = append(ui.Instructions, instruction)
	}

	// run emulation

	// select event channel

	// normal mode: use time tickers
	nextFrame := time.Tick(time.Duration(1000/FPS) * time.Millisecond)
	nextTick := time.Tick(time.Duration(1_000_000/HZ) * time.Microsecond)
	if isDebug {
		// debug mode: manual ticker
		nextTick = ui.Tick
	}

	var frameCnt uint16 = 0
	quitSign := false

	// to show the display initially
	ui.ClearDisplay()

	for !quitSign {
		// indepedent CPU clock rate and graphics refresh rate
		select {
		// new CPU tick: process next instruction
		case <-nextTick:
			interpreter.NextCPUCycle(&hw, ui)
		case <-nextFrame:
			ui.ClearDebugPanels()
			ui.DrawDebugPanels(hw)

			ui.Render()
			frameCnt++
		case keyInput := <-ui.Input:
			hw.Input = byte(keyInput)
		case <-ui.Quit:
			quitSign = true
		}
	}
}
