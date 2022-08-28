package math

import (
	"jvm/instructions/base"
	"jvm/rtda"
)

// iinc 指令给局部变量表中的int变量增加常量值
//      局部变量表索引和常量值都由指令的操作数提供
type IINC struct {
	Index uint
	Const int32
}

func (self *IINC) FetchOperands(reader *base.BytecodeReader) {
	self.Index = uint(reader.ReadInt8())
	self.Const = int32(reader.ReadInt8())
}

func (self *IINC) Execute(frame *rtda.Frame) {
	localVars := frame.LocalVars()
	val := localVars.GetInt(self.Index)
	val += self.Const
	localVars.SetInt(self.Index, val)
}
