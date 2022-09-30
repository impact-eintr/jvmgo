package references

import (
	"fmt"
	"jvm/instructions/base"
	"jvm/rtda"
	"jvm/rtda/heap"
	"reflect"
)

type ATHROW struct {
	base.NoOperandsInstruction
}

func (self *ATHROW) Execute(frame *rtda.Frame) {
	ex := frame.OperandStack().PopRef()
	if ex == nil {
		panic("java.lang.NullPointerException")
	}
	thread := frame.Thread()
	if !findAndGotoExceptionHandler(thread, ex) {
		handleUncaughtException(thread, ex)
	}
}

// Find and goto the Exception handler
func findAndGotoExceptionHandler(thread *rtda.Thread, ex *heap.Object) bool {
	for {
		frame := thread.CurrentFrame()
		pc := frame.NextPC() - 1

		handlerPC := frame.Method().FindExceptionHandler(ex.Class(), pc)
		if handlerPC > 0 {
			stack := frame.OperandStack()
			stack.Clear()
			stack.PushRef(ex)
			frame.SetNextPC(handlerPC)
			return true
		}

		thread.PopFrame()
		if thread.IsStackEmpty() {
			break
		}
	}
	return false
}

// handle Uncaught Exception by printing its message
func handleUncaughtException(thread *rtda.Thread, ex *heap.Object) {
	thread.ClearStack() // 清空虚拟机栈 解释器终止执行

	jMsg := ex.GetRefVar("detailMessage", "Ljava/lang/String;")
	goMsg := heap.GoString(jMsg)
	fmt.Printf("%s: %s\n", ex.Class().JavaName(), goMsg)

	stes := reflect.ValueOf(ex.Extra()) // 异常对象的extra存放jvm栈信息
	for i := 0;i < stes.Len();i++ {
		ste := stes.Index(i).Interface().(interface {String() string})
		fmt.Println("\tat", ste.String())
	}
}
