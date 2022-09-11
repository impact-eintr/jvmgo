package heap

import "jvm/classfile"

type FieldRef struct {
	MemberRef
	field *Field
}

func newFieldRef(cp *ConstantPool, refInfo *classfile.ConstantFieldrefInfo) *FieldRef{
	ref := &FieldRef{}
	ref.cp = cp
	ref.copyMemberRefInfo(&refInfo.ConstantMemberrefInfo)
	return ref
}

// 对字段引用进行解引用
func (self *FieldRef) ResolvedField() *Field {
	if self.field == nil {
		self.resolveFieldRef()
	}
	return self.field
}

// 如果类D想通过字段符号引用访问类C的某个字段,首先要解析符号引用得到类C,
// 然后根据字段名和描述符查找字段。
// 如果字段查找失败,则虚拟机抛出NoSuchFieldError异常。
// 如果查找成功,但D没有足够的权限访问该字段,
// 则虚拟机抛出IllegalAccessError异常。
func (self *FieldRef) resolveFieldRef() {
	d := self.cp.class
	c := self.ResolvedClass()
	field := lookupField(c, self.name, self.descriptor)

	if field == nil {
		panic("java.lang.NoSuchFieldError")
	}
	if !field.isAccessibleTo(d) {
		panic("java.lang.IllegalAccessError")
	}
	self.field = field
}

func lookupField(c *Class, name, descriptor string) *Field {
	// check fields
	for _, field := range c.fields {
		if field.name == name && field.descriptor == descriptor {
			return field
		}
	}
	// check interfaces
	for _, iface := range c.interfaces {
		if field := lookupField(iface, name, descriptor); field != nil {
			return field
		}
	}
	// check super class
	if c.superClass != nil {
		return lookupField(c.superClass, name, descriptor)
	}
	return nil
}
