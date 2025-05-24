package interpreter

import (
	"chipgo/entities"
	"chipgo/hw"
	"chipgo/ui"
	"fmt"
	"math/rand"
)

func NextCPUCycle(hw *hw.HW, ui *ui.UI) {
	instruction := BuildInstruction(hw.Memory[hw.PC], hw.Memory[hw.PC+1], hw.PC)

	// get instruction parts

	hex := instruction.Hex
	addr := instruction.Addr
	x := instruction.X
	y := instruction.Y
	kk := instruction.Bytekk
	nibble := instruction.Nibble

	// increase now, so JM can overwrite it
	hw.PC += 2 // 2 bytes

	switch hex {
	case 0x00E0:
		ui.ClearDisplay()
	case 0x00EE:
		hw.SC--
		hw.PC = hw.Stack[hw.SC]
	case hex & 0x0FFF:
		ui.Log(fmt.Sprintf("Skip instruction %s", decompileInstruction(instruction)))
	case hex & 0x1FFF:
		hw.PC = addr
	case hex & 0x2FFF:
		hw.Stack[hw.SC] = hw.PC
		hw.SC++
		hw.PC = addr
	case hex & 0x3FFF:
		if hw.V[x] == kk {
			// increase program counter by 2 (2 bytes now and 2 at the end)
			hw.PC += 2
		}
	case hex & 0x4FFF:
		if hw.V[x] != kk {
			// increase program counter by 2 (2 bytes now and 2 at the end)
			hw.PC += 2
		}
	case hex & 0x5FF0:
		if hw.V[x] == hw.V[y] {
			hw.PC += 2
		}
	case hex & 0x6FFF:
		hw.V[x] = kk
	case hex & 0x7FFF:
		hw.V[x] += kk
	case hex & 0x8FF0:
		hw.V[x] = hw.V[y]
	case hex & 0x8FF1:
		hw.V[x] |= hw.V[y]
	case hex & 0x8FF2:
		hw.V[x] &= hw.V[y]
	case hex & 0x8FF3:
		hw.V[x] ^= hw.V[y]
	case hex & 0x8FF4:
		// if carry > sum, then set flag
		if (hw.V[x] ^ hw.V[y]) > (hw.V[x] + hw.V[y]) {
			hw.V[15] = 1
		} else {
			hw.V[15] = 0
		}
		hw.V[x] += hw.V[y]
	case hex & 0x8FF5:
		if hw.V[x] > hw.V[y] {
			hw.V[15] = 1
		} else {
			hw.V[15] = 0
		}
		hw.V[x] -= hw.V[y]
	case hex & 0x8FF6:
		hw.V[15] = 0x1 ^ hw.V[x]
		hw.V[x] >>= 1
	case hex & 0x8FF7:
		if hw.V[y] > hw.V[x] {
			hw.V[15] = 1
		} else {
			hw.V[15] = 0
		}
		hw.V[y] -= hw.V[x]
	case hex & 0x8FFE:
		if 0x10 >= hw.V[x] {
			hw.V[15] = 1
		} else {
			hw.V[15] = 0
		}
		hw.V[x] <<= 1
	case hex & 0x9FF0:
		if hw.V[x] != hw.V[y] {
			hw.PC += 2
		}
	case hex & 0xAFFF:
		hw.I = addr
	case hex & 0xBFFF:
		hw.PC = addr + uint16(hw.V[0])
	case hex & 0xCFFF:
		hw.V[x] = byte(rand.Intn(256)) & kk
	case hex & 0xDFFF:
		col := ui.DrawSprite(hw.Memory[hw.I:hw.I+uint16(nibble)], int(hw.V[x]), int(hw.V[y]))
		if col {
			hw.V[15] = 1
		} else {
			hw.V[15] = 0
		}
	case hex & 0xEF9E:
	// TODO: input
	case hex & 0xEFA1:
	// TODO: input
	case hex & 0xFF07:
		hw.V[x] = hw.DT
	case hex & 0xFF0A:
	// TODO: input
	case hex & 0xFF15:
		hw.DT = hw.V[x]
	case hex & 0xFF18:
		_ = hw.V[x]
	case hex & 0xFF1E:
		hw.I += uint16(hw.V[x])
	case hex & 0xFF29:
		hw.I = uint16(hw.V[x]) * 5 // position of sprite x
	case hex & 0xFF33:
		hw.Memory[hw.I] = hw.V[x] / 100
		hw.Memory[hw.I+1] = (hw.V[x] % 100) / 10
		hw.Memory[hw.I+2] = (hw.V[x] % 100) % 10
	case hex & 0xFF55:
		for idx := range x {
			hw.Memory[hw.I+uint16(idx)] = hw.V[idx]
		}
	case hex & 0xFF65:
		for idx := range x {
			hw.V[idx] = hw.Memory[hw.I+uint16(idx)]
		}
	default:
		ui.Log(fmt.Sprintf("unrecognized instruction 0x%x at 0x%x", hex, hw.PC))
	}

	// increment program counter
	hw.DecreaseTimers()
}

func decompileInstruction(instruction entities.Instruction) string {
	hex := instruction.Hex
	addr := instruction.Addr
	x := instruction.X
	y := instruction.Y
	kk := instruction.Bytekk
	nibble := instruction.Nibble
	switch hex {
	case 0x00E0:
		return "CLS"
	case 0x00EE:
		return "RET"
	case hex & 0x0FFF:
		return fmt.Sprintf("SYS 0x%04x", addr)
	case hex & 0x1FFF:
		return fmt.Sprintf("JP 0x%04x", addr)
	case hex & 0x2FFF:
		return fmt.Sprintf("CALL 0x%04x", addr)
	case hex & 0x3FFF:
		return fmt.Sprintf("SE V[%02d], %x", x, kk)
	case hex & 0x4FFF:
		return fmt.Sprintf("SNE V[%02d], %x", x, kk)
	case hex & 0x5FF0:
	case hex & 0x6FFF:
		return fmt.Sprintf("LD V[%02d], %x", x, kk)
	case hex & 0x7FFF:
		return fmt.Sprintf("ADD V[%02d], %x", x, kk)
	case hex & 0x8FF0:
		return fmt.Sprintf("LD V[%02d], V[%02d]", x, y)
	case hex & 0x8FF1:
	case hex & 0x8FF2:
	case hex & 0x8FF3:
	case hex & 0x8FF4:
		return fmt.Sprintf("ADD V[%02d], V[%02d]", x, y)
	case hex & 0x8FF5:
	case hex & 0x8FF6:
	case hex & 0x8FF7:
	case hex & 0x8FFE:
	case hex & 0x9FF0:
	case hex & 0xAFFF:
		return fmt.Sprintf("LD I, 0x%x", addr)
	case hex & 0xBFFF:
		return fmt.Sprintf("JP V[00], 0x%x", addr)
	case hex & 0xCFFF:
		return fmt.Sprintf("RND V[%02d] 0x%x", x, kk)
	case hex & 0xDFFF:
		return fmt.Sprintf("DRW V[%02d], V[%02d], 0x%x", x, y, nibble)
	case hex & 0xEFA1:
		return fmt.Sprintf("SKNP V[%02d] (unimplemented)", x)
	case hex & 0x9F9E:
	case hex & 0x9FA1:
	case hex & 0xFF07:
	case hex & 0xFF0A:
		return fmt.Sprintf("LD V[%02d], K", x)
	case hex & 0xFF15:
	case hex & 0xFF18:
	case hex & 0xFF1E:
		return fmt.Sprintf("ADD I, V[%02d]", x)
	case hex & 0xFF29:
	case hex & 0xFF33:
	case hex & 0xFF55:
	case hex & 0xFF65:
	}

	return "UNKNOWN"
}

func BuildInstruction(high, low byte, pc uint16) entities.Instruction {
	// build instruction from bytes
	hex := (uint16(high) << 8) + uint16(low)
	instruction := entities.Instruction{
		Hex:     hex,
		Addr:    hex & 0x0FFF,
		X:       (high & 0x0F),
		Y:       (low & 0xF0) >> 4,
		Bytekk:  low & 0x00FF,
		Counter: pc,
		Nibble:  low & 0x000F,
	}
	instruction.Decompiled = decompileInstruction(instruction)
	return instruction
}
