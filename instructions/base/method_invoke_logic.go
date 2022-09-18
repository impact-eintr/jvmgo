package base

import (
	"fmt"
	"jvm/rtda"
	"jvm/rtda/heap"
)

func InvokeMethod(invokerFrame *rtda.Frame, method *heap.Method) {
	thread := invokerFrame.Thread()
	newFrame := thread.NewFrame(method) // 分配合适的栈帧空间
	thread.PushFrame(newFrame)

	argSlotCount := int(method.ArgSlotCount())
	if argSlotCount > 0 {
		for i := argSlotCount - 1;i >= 0; i-- {
			slot := invokerFrame.OperandStack().PopSlot() // 参数位于栈顶
			newFrame.LocalVars().SetSlot(uint(i), slot) // 参数拷贝
		}
	}

	// hack!
	if method.IsNative() {
		if method.Name() == "registerNatives" {
			thread.PopFrame()
		} else {
			panic(fmt.Sprintf("native method: %v.%v%v\n",
				method.Class().Name(), method.Name(), method.Descriptor()))
		}
	}
}
