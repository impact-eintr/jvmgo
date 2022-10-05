package base

import (
	"jvm/rtda"
	"jvm/rtda/heap"
)

// 初始化类
// 标志已经初始化
// 构造函数压栈
// 递归构造父类
func InitClass(thread *rtda.Thread, class *heap.Class) {
	class.StartInit()
	scheduleClinit(thread, class)
	initSuperClass(thread, class)
}

func scheduleClinit(thread *rtda.Thread, class *heap.Class) {
	clinit := class.GetClinitMethod()
	if clinit != nil {
		// exec <clinit>
		newFrame := thread.NewFrame(clinit)
		thread.PushFrame(newFrame)
	}
}

func initSuperClass(thread *rtda.Thread, class *heap.Class) {
	if !class.IsInterface() { // not a interface
		superClass := class.SuperClass()
		if superClass != nil && !superClass.InitStarted() {
			InitClass(thread, superClass)
		}
	}
}
