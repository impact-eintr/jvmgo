package reverved

import (
	"jvm/instructions/base"
	"jvm/native"
	"jvm/rtda"
	_ "jvm/native/java/lang"
	_ "jvm/native/java/io"
	_ "jvm/native/sun/misc"
)

// Invoke native method
type INVOKE_NATIVE struct {
	base.NoOperandsInstruction
}

func (self *INVOKE_NATIVE) Execute(frame *rtda.Frame) {
	method := frame.Method()
	className := method.Class().Name()
	methodName := method.Name()
	methodDescriptor := method.Descriptor()

	nativeMethod := native.FindNativeMethod(className, methodName, methodDescriptor)
	if nativeMethod == nil {
		methodInfo := className + "." + methodName + methodDescriptor
		panic("java.lang.UnsatisfiedLinkError:" + methodInfo)
	}
	nativeMethod(frame)
}
