package entities

type Instruction struct {
	Hex        uint16
	Addr       uint16
	X          byte
	Y          byte
	Bytekk     byte
	Nibble     byte
	Counter    uint16
	Decompiled string
}
