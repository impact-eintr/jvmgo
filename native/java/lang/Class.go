package lang

import (
	"jvm/native"
	"jvm/rtda"
	"jvm/rtda/heap"
)

const jlClass = "java/lang/Class"

func init() {
  native.Register(jlClass, "getPrimitiveClass",
		"(Ljava/lang/String;)Ljava/lang/Class;  ", getPrimitiveClass)
	native.Register(jlClass, "getName0", "()Ljava/lang/String;", getName0)
	native.Register(jlClass, "desiredAssertionStatus0",
		"(Ljava/lang/Class;)Z", desiredAssertionStatus0)
  //native.Register(jlClass, "isInterface", "()Z", isInterface)
  //native.Register(jlClass, "isPrimitive", "()Z", isPrimitive)
}


// static native Class<?> getPrimitiveClass(String name);
// (Ljava/lang/String;)Ljava/lang/Class;
func getPrimitiveClass(frame *rtda.Frame) {
	nameObj := frame.LocalVars().GetRef(0) // 获取类名
	name := heap.GoString(nameObj)

	loader := frame.Method().Class().Loader()
	class := loader.LoadClass(name).JClass() // 获取类对象引用 java.lang.Class

	frame.OperandStack().PushRef(class)
}

// private native String getName0();
// ()Ljava/lang/String;
func getName0(frame *rtda.Frame)  {
	this := frame.LocalVars().GetThis()
	class := this.Extra().(*heap.Class)

	name := class.JavaName()
	nameObj := heap.JString(class.Loader(), name)

	frame.OperandStack().PushRef(nameObj)
}

// private static native boolean desiredAssertionStatus0(Class<?> clazz);
// (Ljava/lang/Class;)Z
func desiredAssertionStatus0(frame *rtda.Frame) {
	// TODO
	frame.OperandStack().PushBoolean(false)
}

// public native boolean isInterface();
// ()Z
func isInterface(frame *rtda.Frame)  {

}

// public native boolean isPrimitive();
// ()Z
func isPrimitive(frame *rtda.Frame) {

}