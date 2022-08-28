package stack

import (
	"jvm/instructions/base"
	"jvm/rtda"
)

type POP struct {
	base.NoOperandsInstruction
}

// pop int float ...
func (self *POP) Execute(frame *rtda.Frame) {
	stack := frame.OperandStack()
	stack.PopSlot()
}

type POP2 struct {
	base.NoOperandsInstruction
}

// pop dpuble long
func (self *POP2) Execute(frame *rtda.Frame) {
	stack := frame.OperandStack()
	stack.PopSlot()
	stack.PopSlot()
}
