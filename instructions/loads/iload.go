package loads

import (
	"jvm/instructions/base"
	"jvm/rtda"
)

// iload 指令的索引来自于操作数
type ILOAD struct {
	base.Index8Instruction
}

func (self *ILOAD) Execute(frame *rtda.Frame) {
	_iload(frame, self.Index)
}

// 其余指令的索引隐含于操作码中
type ILOAD_0 struct {
	base.NoOperandsInstruction
}
func (self *ILOAD_0) Execute(frame *rtda.Frame) {
	_iload(frame, 0)
}

type ILOAD_1 struct {
	base.NoOperandsInstruction
}
func (self *ILOAD_1) Execute(frame *rtda.Frame) {
	_iload(frame, 1)
}

type ILOAD_2 struct {
	base.NoOperandsInstruction
}
func (self *ILOAD_2) Execute(frame *rtda.Frame) {
	_iload(frame, 2)
}

type ILOAD_3 struct {
	base.NoOperandsInstruction
}
func (self *ILOAD_3) Execute(frame *rtda.Frame) {
	_iload(frame, 3)
}

// 加载指令总局部变量表中获取变量 然后推入操作数栈顶
func _iload(frame *rtda.Frame, index uint) {
	val := frame.LocalVars().GetInt(index)
	frame.OperandStack().PushInt(val)
}
