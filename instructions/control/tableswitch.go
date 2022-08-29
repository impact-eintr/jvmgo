package control

import (
	"jvm/instructions/base"
	"jvm/rtda"
)

// 连续 case
type TABLE_SWITCH struct {
	defalutOffset int32
	low int32
	high int32
	jumpOffsets []int32
}

func (self *TABLE_SWITCH) FetchOperands(reader *base.BytecodeReader) {
	reader.SkipPadding()
	self.defalutOffset = reader.ReadInt32()
	self.low = reader.ReadInt32()
	self.high = reader.ReadInt32()
	jumpOffsetCount := self.high - self.low + 1
	self.jumpOffsets = reader.ReadInt32s(jumpOffsetCount)
}

func (self *TABLE_SWITCH) Execute(frame *rtda.Frame) {
	index := frame.OperandStack().PopInt()

	var offset int
	if index >= self.low && index <= self.high {
		offset = int(self.jumpOffsets[index-self.low])
	} else {
		offset = int(self.defalutOffset)
	}

	base.Branch(frame, offset)
}
