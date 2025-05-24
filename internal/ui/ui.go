package ui

import (
	"chipgo/internal/entities"
	"chipgo/internal/hw"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gdamore/tcell"
	"github.com/mattn/go-runewidth"
)

const (
	DISPLAY_WIDTH  = 64
	DISPLAY_HEIGHT = 32

	CONSOLE_WIDTH  = 64
	CONSOLE_HEIGHT = 16

	PROGRAM_PAGE_SIZE = 40
	KEY_RUNES         = "0123456789abcdef"
)

type UI struct {
	screen       tcell.Screen
	lock         sync.Locker
	width        int
	height       int
	display      panel
	console      panel
	program      panel
	hardware     panel
	logs         []string
	Instructions []entities.Instruction
	isDebug      bool
	Quit         chan bool
	Tick         chan time.Time
	Input        chan int
}

type panel struct {
	name         string
	posX         int
	posY         int
	padding      int
	width        int
	height       int
	screenStyle  tcell.Style
	textStyle    tcell.Style
	paddingStyle tcell.Style
}

func (p *panel) getStartX(x int) int {
	return x + p.posX
}

func (p *panel) getStartY(y int) int {
	return y + p.posY
}

func (p *panel) getNormX(x int) int {
	return (x % p.width) + p.posX + p.padding
}

func (p *panel) getNormY(y int) int {
	return (y % p.height) + p.posY + p.padding
}

func Init(isDebug bool) (*UI, error) {
	s, err := tcell.NewScreen()
	if err != nil {
		return nil, err
	}

	if err := s.Init(); err != nil {
		return nil, err
	}

	w, h := s.Size()
	if w < DISPLAY_WIDTH || h < DISPLAY_HEIGHT {
		return nil, fmt.Errorf("insufficient resolution to initialize the dislay")
	}

	screenStyle := tcell.StyleDefault.
		Background(tcell.ColorGreenYellow).
		Foreground(tcell.ColorWhite)

	textStyle := tcell.StyleDefault.Foreground(tcell.ColorCadetBlue).
		Background(tcell.ColorWhite)

	paddingStyle := tcell.StyleDefault.Background(tcell.ColorMintCream)

	display := panel{
		name:         "DISPLAY",
		width:        DISPLAY_WIDTH,
		height:       DISPLAY_HEIGHT,
		posX:         0,
		posY:         0,
		padding:      1,
		screenStyle:  screenStyle,
		textStyle:    textStyle,
		paddingStyle: paddingStyle,
	}

	console := panel{
		name:         "CONSOLE",
		width:        CONSOLE_WIDTH,
		height:       h - display.height - display.padding*2 - 2,
		posX:         0,
		posY:         display.height + display.padding*2,
		padding:      1,
		screenStyle:  tcell.StyleDefault.Background(tcell.ColorBlack),
		textStyle:    tcell.StyleDefault.Foreground(tcell.ColorWhite),
		paddingStyle: paddingStyle,
	}

	program := panel{
		name:         "PROGRAM",
		width:        (w - display.width - 2*display.padding - 2) / 2,
		height:       h - 2,
		posX:         display.width + display.padding*2,
		posY:         0,
		padding:      1,
		screenStyle:  tcell.StyleDefault.Background(tcell.ColorBlack),
		textStyle:    tcell.StyleDefault.Foreground(tcell.ColorWhite),
		paddingStyle: paddingStyle,
	}

	hardware := panel{
		name:         "HARDWARE",
		width:        w - (program.posX + program.width + 2*program.padding + 2),
		height:       h - 2,
		posX:         program.posX + program.width + 2*program.padding,
		posY:         0,
		padding:      1,
		screenStyle:  tcell.StyleDefault.Background(tcell.ColorBlack),
		textStyle:    tcell.StyleDefault.Foreground(tcell.ColorWhite),
		paddingStyle: paddingStyle,
	}

	quit := make(chan bool)
	tick := make(chan time.Time)
	input := make(chan int)
	ui := UI{
		screen:   s,
		lock:     &sync.Mutex{},
		display:  display,
		console:  console,
		program:  program,
		hardware: hardware,
		isDebug:  isDebug,
		Quit:     quit,
		Tick:     tick,
		Input:    input,
	}

	// poll escape sequence
	// TODO: is this thread safe?
	go ui.handleEvents(isDebug)
	return &ui, nil
}

func (ui *UI) handleEvents(isDebug bool) {
	if isDebug {
		ui.Log("DEBUG mode: press ENTER to cycle CPU instructions")
	}
	for {
		switch ev := ui.screen.PollEvent().(type) {
		case *tcell.EventKey:
			if ev.Key() == tcell.KeyEscape {
				ui.screen.Fini()
				ui.Quit <- true
			}
			if ev.Key() == tcell.KeyEnter {
				if isDebug {
					ui.Tick <- ev.When()
				}
			}
			if keyRune := strings.IndexRune(KEY_RUNES, ev.Rune()); keyRune > -1 {
				ui.Input <- keyRune
			}
		}
	}
}

/*
accepts a sprite represented as a byte array (size ranging from 5 to 15)
the point (x,y) represents the top-left "corner" of the sprite
*/
func (ui *UI) DrawSprite(sprite []byte, x, y int) bool {
	ui.lock.Lock()
	defer ui.lock.Unlock()

	col := false
	for i := range 8 {
		for j, b := range sprite {
			pos := 8 - i - 1 // current bit pos

			// only draw if current bit is 1, aka bitmask operation is true
			normX, normY := ui.display.getNormX(int(x)+i), ui.display.getNormY(int(y)+j)
			if b&(1<<pos) > 0 {
				// XOR operation between existing bit and new bit. If there is an active pixel
				// clear position
				if mainc, _, _, _ := ui.screen.GetContent(normX, normY); mainc == '*' {
					col = true
					ui.screen.SetContent(normX, normY, ' ', nil, ui.display.screenStyle)
				} else {
					ui.screen.SetContent(normX, normY, '*', nil, ui.display.textStyle)
				}
			}
		}
	}
	return col
}

func (ui *UI) clear(p panel) {
	// panel screen
	ui.lock.Lock()
	for x := range p.width {
		for y := range p.height {
			ui.screen.SetContent(p.getNormX(x), p.getNormY(y), ' ', nil, p.screenStyle)
		}
	}
	// panel padding
	for x := range p.width + p.padding*2 {
		for y := range p.height + p.padding*2 {
			if isPadding(x, p.padding, p.width) || isPadding(y, p.padding, p.height) {
				ui.screen.SetContent(p.getStartX(x), p.getStartY(y), ' ', nil, p.paddingStyle)
			}
		}
	}
	ui.lock.Unlock()

	// title
	ui.emitString(p, p.name, (p.width-len(p.name))/2, -1)
}

func isPadding(coord, padding, length int) bool {
	return (coord >= 0 && coord < padding) || (coord >= length+padding && coord <= padding*2+length)
}

func (ui *UI) emitString(p panel, str string, x, y int) {
	for _, c := range str {
		var comb []rune
		w := runewidth.RuneWidth(c)
		if w == 0 {
			comb = []rune{c}
			c = ' '
			w = 1
		}
		ui.lock.Lock()
		ui.screen.SetContent(p.getNormX(x), p.getNormY(y), c, comb, p.textStyle)
		ui.lock.Unlock()
		x += w
	}
}

func (ui *UI) ClearDisplay() {
	ui.clear(ui.display)
}

func (ui *UI) ClearDebugPanels() {
	ui.clear(ui.console)
	ui.clear(ui.program)
	ui.clear(ui.hardware)
}

func (ui *UI) DrawDebugPanels(hw hw.HW) {
	// Console
	ui.emitString(ui.console, "> press ESC to quit", 1, 1)
	logStart := 0
	if len(ui.logs) > 6 {
		logStart = len(ui.logs) - 6
	}
	for i, log := range ui.logs[logStart:] {
		ui.emitString(ui.console, log, 1, i+3)
	}

	// Program
	currentInstructionPos := (hw.PC - 0x200) / 2

	// check current page
	pageStart := (currentInstructionPos / PROGRAM_PAGE_SIZE) * PROGRAM_PAGE_SIZE
	pageEnd := min(pageStart+PROGRAM_PAGE_SIZE, uint16(len(ui.Instructions)))

	for i, instruction := range ui.Instructions[pageStart:pageEnd] {
		marker := ""
		if hw.PC == instruction.Counter {
			marker = "  <-- PC"
		}
		ui.emitString(ui.program, fmt.Sprintf("0x%x: 0x%x %s%s", instruction.Counter, instruction.Hex, instruction.Decompiled, marker), 1, i+1)
	}

	// Hardware
	line := 0
	for i, vi := range hw.V {
		line = i + 1
		ui.emitString(ui.hardware, fmt.Sprintf("V[%02d] = 0x%x", i, vi), 1, line)
	}

	line += 2
	ui.emitString(ui.hardware, fmt.Sprintf("PC = 0x%x", hw.PC), 1, line)
	line++
	ui.emitString(ui.hardware, fmt.Sprintf("I = 0x%x", hw.I), 1, line)

	line++
	for i, stc := range hw.Stack {
		line++
		marker := ""
		if hw.SC == byte(i) {
			marker = "  <-- SC"
		}
		ui.emitString(ui.hardware, fmt.Sprintf("Stack[%02d] = 0x%x%s", i, stc, marker), 1, line)
	}

	line += 2
	ui.emitString(ui.hardware, fmt.Sprintf("DT = 0x%x", hw.DT), 1, line)
	line++
	ui.emitString(ui.hardware, fmt.Sprintf("ST = 0x%x", hw.ST), 1, line)
}

func (ui *UI) Render() {
	ui.screen.Show()
}

func (ui *UI) Stop() {
	ui.screen.Fini()
}

func (ui *UI) Log(msg string) {
	ui.logs = append(ui.logs, msg)
}
