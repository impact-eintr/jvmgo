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

// 如果类D通过符号引用类C的话
// 要解析N 先用D的类加载器加载C 然后检查D是否有权限访问C
func (self *SymRef) resolveClassRef() {
	d := self.cp.class 
	c := d.loader.LoadClass(self.className)
	if !c.isAccessibleTo(d) {
		panic("java.lang.IllegalAccessError")
	}
	self.class = c
}
