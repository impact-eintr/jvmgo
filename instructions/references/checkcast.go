package references

import (
	"jvm/instructions/base"
	"jvm/rtda"
	"jvm/rtda/heap"
)

type CHECK_CAST struct {
	base.Index16Instruction
}

func (self *CHECK_CAST) Execute(frame *rtda.Frame) {
	stack := frame.OperandStack()
	// 先从操作数栈中弹出对象引用,再推回去,这样就不会改变操作数栈的状态
	ref := stack.PopRef()
	stack.PushRef(ref)
	if ref == nil {
		return
	}

	cp := frame.Method().Class().ConstantPool()
	classRef := cp.GetConstant(self.Index).(*heap.ClassRef)
	class := classRef.ResolvedClass()
	if !ref.IsInstanceOf(class) {
		panic("java.lang.ClasCastException")
	}
}
