package heap

import (
	"jvm/classfile"
	"strings"
)

type Class struct {
	accessFlags       uint16 // 访问级别
	name              string // thisClassName
	superClassName    string // 父类名
	interfaceNames    []string // 实现的接口名
	constantPool      *ConstantPool // 运行时常量池指针
	fields            []*Field // 类字段
	methods           []*Method // 类方法
	sourceFile        string // 源文件
	loader            *ClassLoader // 类加载器
	superClass        *Class // 父类指针
	interfaces        []*Class // 实现的接口表
	instanceSlotCount uint // 运行时数据占用的槽量
	staticSlotCount   uint // 静态数据占用的槽量
	staticVars        Slots // 静态数据
	initStarted       bool // 表示类的<clinit>方法是否已经开始执行
	jClass            *Object // java.lang.Class的变量引用
}

func newClass(cf *classfile.ClassFile) *Class {
	class := &Class{}
	class.accessFlags = cf.AccessFlags()
	class.name = cf.ClassName()
	class.superClassName = cf.SuperClassName()
	class.interfaceNames = cf.InterfaceNames()
	class.constantPool = newConstantPool(class, cf.ConstantPool()) // 加载运行时常量池
	class.fields = newFileds(class, cf.Fields())                   // 加载运行时字段
	class.methods = newMethods(class, cf.Methods())                // 加载运行时方法
	class.sourceFile = getSourceFile(cf)
	return class
}

func getSourceFile(cf *classfile.ClassFile) string {
	if sfAttr := cf.SourceFileAttribute(); sfAttr != nil {
		return sfAttr.FileName()
	}
	return "Unknown" // TODO
}

func (self *Class) IsPublic() bool {
	return 0 != self.accessFlags&ACC_PUBLIC
}

func (self *Class) IsFinal() bool {
	return 0 != self.accessFlags&ACC_FINAL
}

func (self *Class) IsSuper() bool {
	return 0 != self.accessFlags&ACC_SUPER
}

func (self *Class) IsInterface() bool {
	return 0 != self.accessFlags&ACC_INTERFACE
}

func (self *Class) IsAbstract() bool {
	return 0 != self.accessFlags&ACC_ABSTRACT
}

func (self *Class) IsSynthetic() bool {
	return 0 != self.accessFlags&ACC_SYNTHETIC
}

func (self *Class) IsAnnotation() bool {
	return 0 != self.accessFlags&ACC_ANNOTATION
}

func (self *Class) IsEnum() bool {
	return 0 != self.accessFlags&ACC_ENUM
}

// getters
func (self *Class) AccessFlags() uint16 {
	return self.accessFlags
}

func (self *Class) Name() string {
	return self.name
}

func (self *Class) ConstantPool() *ConstantPool {
	return self.constantPool
}

func (self *Class) Fields() []*Field {
	return self.fields
}

func (self *Class) Methods() []*Method {
	return self.methods
}

func (self *Class) SourceFile() string {
	return self.sourceFile
}

func (self *Class) Loader() *ClassLoader {
	return self.loader
}

func (self *Class) SuperClass() *Class {
	return self.superClass
}

func (self *Class) Interfaces() []*Class {
	return self.interfaces
}

func (self *Class) StaticVars() Slots {
	return self.staticVars
}

func (self *Class) InitStarted() bool {
	return self.initStarted
}

func (self *Class) JClass() *Object {
	return self.jClass
}

func (self *Class) StartInit() {
	self.initStarted = true
}

// JVM 5.4.4
// 检测是否可以被某个类访问:
func (self *Class) isAccessibleTo(other *Class) bool {
	return self.IsPublic() ||
		self.GetPackageName() == other.GetPackageName()
}

func (self *Class) GetPackageName() string {
	if i := strings.LastIndex(self.name, "/"); i >= 0 {
		return self.name[:i]
	}
	return ""
}

func (self *Class) GetMainMethod() *Method {
	return self.getMethod("main", "([Ljava/lang/String;)V", true)
}

func (self *Class) GetClinitMethod() *Method {
	return self.getMethod("<clinit>", "()V", true)
}

func (self *Class) getMethod(name, descriptor string, isStatic bool) *Method {
	for c := self; c != nil; c = c.superClass {
		for _, method := range c.methods {
			if method.IsStatic() == isStatic &&
				method.name == name && method.descriptor == descriptor {
				return method
			}
		}
	}
	return nil
}

func (self *Class) getField(name, descriptor string, isStatic bool) *Field {
	for c := self; c != nil; c = c.superClass {
		for _, field := range c.fields {
			if field.IsStatic() == isStatic &&
				field.name == name && field.descriptor == descriptor {
				return field
			}
		}
	}
	return nil
}

func (self *Class) isJlObject() bool {
	return self.name == "java/lang/Object"
}

func (self *Class) isJlCloneable() bool {
	return self.name == "java/lang/Cloneable"
}

func (self *Class) isJioSerializable() bool {
	return self.name == "java/io/Serializable"
}

func (self *Class) NewObject() *Object {
	return newObject(self)
}

func (self *Class) ArrayClass() *Class {
	arrayClassName := getArrayClassName(self.name)
	return self.loader.LoadClass(arrayClassName)
}

func (self *Class) JavaName() string {
	return strings.Replace(self.name, "/", ".", -1)
}

func (self *Class) IsPrimitive() bool {
	_, ok := primitiveTypes[self.name]
	return ok
}

func (self *Class) GetInstanceMethod(name, descriptor string) *Method {
	return self.getMethod(name, descriptor, false)
}

func (self *Class) GetStaticMethod(name, descriptor string) *Method {
	return self.getMethod(name, descriptor, true)
}

// reflection
func (self *Class) GetRefVar(fieldName, fieldDescriptor string) *Object {
	field := self.getField(fieldName, fieldDescriptor, true)
	return self.staticVars.GetRef(field.slotId)
}

// 设置静态字段值
func (self *Class) SetRefVar(fieldName, fieldDescriptor string, ref *Object) {
	field := self.getField(fieldName, fieldDescriptor, true)
	self.staticVars.SetRef(field.slotId, ref)
}

// 此函数用于反射时获取类的字段信息，不包含继承属性
func (self *Class) GetFields(publicOnly bool) []*Field {
	if publicOnly {
		publibFields := make([]*Field, 0, len(self.fields))
		for _, field := range self.fields {
			if field.IsPublic() {
				publibFields = append(publibFields, field)
			}
		}
		return publibFields
	} else {
		return self.fields
	}
}

func (self *Class) GetConstructor(descriptor string) *Method {
	return self.GetInstanceMethod("<init>", descriptor)
}


func (self *Class) GetConstructors(publicOnly bool) []*Method {
	constructors := make([]*Method, 0, len(self.methods))
	for _, method := range self.methods {
    if method.isConstructor() {
      if !publicOnly || method.IsPublic() {
        constructors = append(constructors, method)
      }
    }
  }
  return constructors
}

func (self *Class) GetMethods(publicOnly bool) []*Method {
	methods := make([]*Method, 0, len(self.methods))
	for _, method := range self.methods {
		if !method.isClinit() && !method.isConstructor() {
			if !publicOnly || method.IsPublic() {
				methods = append(methods, method)
			}
		}
	}
	return methods
}
