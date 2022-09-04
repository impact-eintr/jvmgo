package heap

import (
	"jvm/classfile"
)

type ClassRef struct {
	SymRef
}

func newClassRef(cp *ConstantPool, calssInfo *classfile.ConstantClassInfo) *ClassRef {
}
