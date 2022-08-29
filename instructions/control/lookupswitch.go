package control

import (
	"jvm/instructions/base"
	"jvm/rtda"
)

// 不连续 case
type LOOKUP_SWITCH struct {
	defalutOffset int32
	npairs int32
	matchOffsets []int32
}

func (self *LOOKUP_SWITCH) FetchOperands(reader *base.BytecodeReader) {
	reader.SkipPadding()
	self.defalutOffset = reader.ReadInt32()
	self.npairs = reader.ReadInt32()
	self.matchOffsets = reader.ReadInt32s(self.npairs * 2) // first:key second:value
}

func (self *LOOKUP_SWITCH) Execute(frame *rtda.Frame) {
	key := frame.OperandStack().PopInt()
	for i := int32(0);i < self.npairs * 2;i += 2{
		if self.matchOffsets[i] == key {
			offset := self.matchOffsets[i+1]
			base.Branch(frame, int(offset))
			return
		}
	}
	base.Branch(frame, int(self.defalutOffset))
}
