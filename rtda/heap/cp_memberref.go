package heap

import "jvm/classfile"

type MemberRef struct {
	SymRef
	name string
	descriptor string // 这个描述符是必要的 jvm并不限制类中有不同类型的同名字段
}

func (self *MemberRef) copyMemberRefInfo(refInfo *classfile.ConstantMemberrefInfo) {
	self.className = refInfo.ClassName()
	self.name, self.descriptor = refInfo.NameAndDescriptor()
}

func (self *MemberRef) Name() string {
	return self.name
}

func (self *MemberRef) Descriptor() string {
	return self.descriptor
}
