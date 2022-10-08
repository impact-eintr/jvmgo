package rtda

import "jvm/rtda/heap"

func NewShimFrame(thread *Thread, ops *OperandStack) *Frame {
	return &Frame{
		thread: thread,
		method: heap.ShimReturnMethod(),
		operandStack: ops,
	}
}
