package constants

import (
	"jvm/instructions/base"
	"jvm/rtda"
)

type NOP struct {
	base.NoOperandsInstruction
}

func (self *NOP) Execute(frame *rtda.Frame) {
	// really do nothing
}
