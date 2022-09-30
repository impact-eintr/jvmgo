package lang

import (
	"fmt"
	"jvm/native"
	"jvm/rtda"
	"jvm/rtda/heap"
)

const jlThrowable = "java/lang/Throwable"

// StackTraceElement结构体用来记录Java虚拟机栈帧信息:
type StackTraceElement struct {
	fileName string // fileName字段给出类所在的文件名;
	className string //className字段给出声明方法的类名;
	methodName string // methodName字段给出方法名;
	lineNumber int // lineNumber字段给出帧正在执行哪行代码;
}

func (self *StackTraceElement) String() string {
	return fmt.Sprintf("%s.%s(%s:%d)",
		self.className, self.methodName, self.fileName, self.lineNumber)
}

func init() {
	native.Register(jlThrowable, "fillInStackTrace",
		"(I)Ljava/lang/Throwable;", fillInStackTrace)
}

// (I)Ljava/lang/Throwable;
func fillInStackTrace(frame *rtda.Frame) {
	this := frame.LocalVars().GetThis()
	frame.OperandStack().PushRef(this) // Throwable Class Ref

	stes := createStackTraceElements(this, frame.Thread())
	this.SetExtra(stes)
}

func createStackTraceElements(tObj *heap.Object, thread *rtda.Thread) []*StackTraceElement{
	// 由于栈顶两帧正在执行fillInStackTrace(int)和fillInStackTrace()方法,
	// 所以需要跳过这两帧
	skip := distanceToObject(tObj.Class()) + 2
	frames := thread.GetFrames()[skip:]
	stes := make([]*StackTraceElement, len(frames))
	for i, frame := range frames {
		stes[i] = createStackTraceElement(frame)
	}
	return stes
}

func distanceToObject(class *heap.Class) int {
	distance := 0
	for c := class.SuperClass(); c != nil; c = c.SuperClass() {
		distance++
	}
	return distance
}

func createStackTraceElement(frame *rtda.Frame)  *StackTraceElement {
	method := frame.Method()
	class := method.Class()
	return &StackTraceElement{
		fileName: class.SourceFile(),
		className: class.JavaName(),
		methodName: method.Name(),
		lineNumber: method.GetLineNumber(frame.NextPC() - 1),
	}
}
