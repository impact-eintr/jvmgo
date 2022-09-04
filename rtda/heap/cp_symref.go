package heap


// 1. 类符号引用
// 2. 字段符号引用
// 3. 方法符号引用
// 4. 接口方法符号引用
type SymRef struct {
	cp *ConstantPool // 符号引用所在的运行时常量池指针
	className string
	class *Class
}

func (self *SymRef) ResolvedClass() *Class {
	if self.class == nil {
		self.resolveClassRef()
	}
	return self.class
}

func (self *SymRef) resolveClassRef() {
}
