package lang

import (
	"jvm/native"
	"jvm/rtda"
	"jvm/rtda/heap"
)

const jlSystem = "java/lang/System"

func init() {
	native.Register(jlSystem, "arraycopy",
		"(Ljava/lang/Object;ILjava/lang/Object;II)V", arraycopy)
}

func arraycopy(frame *rtda.Frame) {
	vars := frame.LocalVars()
	src := vars.GetRef(0)
	srcPos := vars.GetInt(1)
	dst := vars.GetRef(2)
	dstPos := vars.GetInt(3)
	length := vars.GetInt(4)

	if src == nil || dst == nil {
		panic("java.lang.NullPointerException")
	}
	if !checkArrayCopy(src, dst) {
		panic("java.lang.ArrayStoreException")
	}
	if srcPos < 0 || dstPos < 0 || length < 0 ||
		srcPos+length > src.ArrayLength() ||
		dstPos+length > dst.ArrayLength() {
		panic("java.lang.IndexOutOfBoundsException")
	}
	heap.ArrayCopy(src, dst, srcPos, dstPos, length)
}

func checkArrayCopy(src, dst *heap.Object) bool {
	srcClass := src.Class()
	dstClass := dst.Class()

	if !srcClass.IsArray() || !dstClass.IsArray() {
		return false
	}
	// 如果是基本类型数组 两种数组应该同类型
	if srcClass.ComponentClass().IsPrimitive() ||
		dstClass.ComponentClass().IsPrimitive() {
		return srcClass == dstClass
	}
	return true
}
