package control

import (
	"jvm/instructions/base"
	"jvm/rtda"
)

// GOTO 实现无条件跳转
type GOTO struct {
	base.BranchInstruction
}

func (self *GOTO) Execute(frame *rtda.Frame) {
	base.Branch(frame,  self.Offset)
}
