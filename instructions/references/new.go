package references

import (
	"fmt"
	"jvm/instructions/base"
	"jvm/rtda"
	"jvm/rtda/heap"
)

// Create new object
type NEW struct {
	base.Index16Instruction
}


func (self *NEW) Execute(frame *rtda.Frame) {
	cp := frame.Method().Class().ConstantPool()
	classRef := cp.GetConstant(self.Index).(*heap.ClassRef)
	class := classRef.ResolvedClass() // 用符号引用加载整个类信息
	// new指令触发构建类实例 但类还没有初始化 终止指令执行
	if !class.InitStarted() {
		fmt.Println("未初始化", class.Name())
		frame.RevertNextPC() // 回置PC
		base.InitClass(frame.Thread(), class)
		return
	}
	fmt.Println("重新执行一次")

	// interface and abstract class can be instantced
	if class.IsInterface() || class.IsAbstract() {
		panic("java.lang.InstantiationError")
	}
	ref := class.NewObject() // 实例化
	frame.OperandStack().PushRef(ref)
}
