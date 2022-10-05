package math

import (
	"jvm/instructions/base"
	"jvm/rtda"
)

// Shitf Left Int
type ISHL struct {
	base.NoOperandsInstruction
}

func (self *ISHL) Execute(frame *rtda.Frame) {
	stack := frame.OperandStack()
	v2 := stack.PopInt()
	v1 := stack.PopInt()
	s := uint32(v2) & 0x1f // 取最后五位 2**5 = 32 足以表示位移
	result := v1 << s
	stack.PushInt(result)
}

// Int 算术右移 有符号右移
type ISHR struct {
	base.NoOperandsInstruction
}

func (self *ISHR) Execute(frame *rtda.Frame) {
	stack := frame.OperandStack()
	v2 := stack.PopInt()
	v1 := stack.PopInt()
	s := uint32(v2) & 0x1f
	result := v1 >> s
	stack.PushInt(result)
}

// 逻辑右移 无符号右移
type IUSHR struct {
	base.NoOperandsInstruction
}

func (self *IUSHR) Execute(frame *rtda.Frame) {
	stack := frame.OperandStack()
	v2 := stack.PopInt()
	v1 := stack.PopInt()
	s := uint32(v2) & 0x1f
	result := int32((uint32(v1) >> s))
	stack.PushInt(result)
}

// Shift Left Long
type LSHL struct {
	base.NoOperandsInstruction
}

func (self *LSHL) Execute(frame *rtda.Frame) {
	stack := frame.OperandStack()
	v2 := stack.PopInt()
	v1 := stack.PopInt()
	s := uint32(v2) & 0x3f
	result := v1 << s
	stack.PushInt(result)
}

// Shift Right Long
type LSHR struct {
	base.NoOperandsInstruction
}

func (self *LSHR) Execute(frame *rtda.Frame) {
	stack := frame.OperandStack()
	v2 := stack.PopInt()
	v1 := stack.PopInt()
	s := uint32(v2) & 0x3f
	result := v1 >> s
	stack.PushInt(result)
}

// Long 算术右移 有符号右移
type LUSHR struct {
	base.NoOperandsInstruction
}

func (self *LUSHR) Execute(frame *rtda.Frame) {
	stack := frame.OperandStack()
	v2 := stack.PopInt()
	v1 := stack.PopInt()
	s := uint32(v2) & 0x3f // 0x0011 1111 2**6=64
	result := int64(uint64(v1) >> s)
	stack.PushLong(result)
}
