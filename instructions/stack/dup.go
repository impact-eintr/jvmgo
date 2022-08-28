package stack

import (
	"jvm/instructions/base"
	"jvm/rtda"
)

type DUP struct {
	base.NoOperandsInstruction
}

// dup 指令复制栈顶的单个变量
func (self *DUP) Execute(frame *rtda.Frame) {
	stack := frame.OperandStack()
	slot := stack.PopSlot()
	stack.PushSlot(slot)
	stack.PushSlot(slot)
}

// dup_x1 复制栈顶的变量并将之插入到栈的前两个变量后
type DUP_X1 struct {
	base.NoOperandsInstruction
}

func (self *DUP_X1) Execute(frame *rtda.Frame) {
	stack := frame.OperandStack()
	slot1 := stack.PopSlot()
	slot2 := stack.PopSlot()
	stack.PushSlot(slot1)
	stack.PushSlot(slot2)
	stack.PushSlot(slot1)
}

// dup_x1 复制栈顶的变量并将之插入到栈的前三个变量后
type DUP_X2 struct {
	base.NoOperandsInstruction
}

func (self *DUP_X2) Execute(frame *rtda.Frame) {
	stack := frame.OperandStack()
	slot1 := stack.PopSlot()
	slot2 := stack.PopSlot()
	slot3 := stack.PopSlot()
	stack.PushSlot(slot1)
	stack.PushSlot(slot3)
	stack.PushSlot(slot2)
	stack.PushSlot(slot1)
}

// dup 指令复制栈顶的两个变量
type DUP2 struct {
	base.NoOperandsInstruction
}

func (self *DUP2) Execute(frame *rtda.Frame) {
	stack := frame.OperandStack()
	slot1 := stack.PopSlot()
	slot2 := stack.PopSlot()
	stack.PushSlot(slot2)
	stack.PushSlot(slot1)
	stack.PushSlot(slot2)
	stack.PushSlot(slot1)
}

// dup_x1 复制栈顶的两个变量并将之插入到栈的前两/三个变量后
type DUP2_X1 struct {
	base.NoOperandsInstruction
}

func (self *DUP2_X1) Execute(frame *rtda.Frame) {
stack := frame.OperandStack()
	slot1 := stack.PopSlot()
	slot2 := stack.PopSlot()
	slot3 := stack.PopSlot()
	stack.PushSlot(slot2)
	stack.PushSlot(slot1)
	stack.PushSlot(slot3)
	stack.PushSlot(slot2)
	stack.PushSlot(slot1)
}

// dup_x1 复制栈顶的两个变量并将之插入到栈的前二/三/四个变量后
type DUP2_X2 struct {
	base.NoOperandsInstruction
}

func (self *DUP2_X2) Execute(frame *rtda.Frame) {
stack := frame.OperandStack()
	slot1 := stack.PopSlot()
	slot2 := stack.PopSlot()
	slot3 := stack.PopSlot()
	slot4 := stack.PopSlot()
	stack.PushSlot(slot2)
	stack.PushSlot(slot1)
	stack.PushSlot(slot4)
	stack.PushSlot(slot3)
	stack.PushSlot(slot2)
	stack.PushSlot(slot1)
}
