package references

import (
	"jvm/instructions/base"
	"jvm/rtda"
)

// Enter monitor for object
type MONITOR_ENTER struct{ base.NoOperandsInstruction }

// todo
func (self *MONITOR_ENTER) Execute(frame *rtda.Frame) {
  ref := frame.OperandStack().PopRef()
  if ref == nil {
    panic("java.lang.NullPointerException")
  }
}

// Exit monitor for object
type MONITOR_EXIT struct{ base.NoOperandsInstruction }

// todo
func (self *MONITOR_EXIT) Execute(frame *rtda.Frame) {
  ref := frame.OperandStack().PopRef()
  if ref == nil {
    panic("java.lang.NullPointerException")
  }
}
