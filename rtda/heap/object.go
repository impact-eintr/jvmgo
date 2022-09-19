package heap

type Object struct {
	class *Class
	data interface{}
}

func newObject(class *Class) *Object {
	return &Object{
		class: class,
		data: newSlots(class.instanceSlotCount), // 分配类实例内存
	}
}

func (self *Object) Class() *Class {
	return self.class
}

func (self *Object) Fields() Slots {
	return self.data.(Slots)
}

func (self *Object) IsInstanceOf(class *Class) bool {
	return class.isAssignableFrom(self.class)
}

// reflection
