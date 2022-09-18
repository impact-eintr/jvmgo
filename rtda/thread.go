package rtda

import "jvm/rtda/heap"

/*
JVM
   Thread
      pc
      Stack
        Frame
           LocalVars
           OperandStack
  **/


type Thread struct {
	pc int
	stack *Stack
}

func NewThread() *Thread {
	return &Thread{
		stack: newStack(1024), // 最多存放1024个栈帧
	}
}

func (self *Thread) PC() int {
	return self.pc
}

func (self *Thread) SetPC(pc int) {
	self.pc = pc
}

func (self *Thread) PushFrame(frame *Frame) {
	self.stack.push(frame)
}

func (self *Thread) PopFrame() *Frame {
	return self.stack.pop()
}

func (self *Thread) CurrentFrame() *Frame {
	return self.stack.top()
}

func (self *Thread) TopFrame() *Frame {
	return self.stack.top()
}

func (self *Thread) IsStackEmpty() bool {
	return self.stack.isEmpty()
}

func (self *Thread) NewFrame(method *heap.Method) *Frame {
	return newFrame(self, method)
}
