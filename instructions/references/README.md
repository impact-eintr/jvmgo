# 类与对象

``` java
public class MyObject {
public static int staticVar;
public int instanceVar;

public static void main(String[] args) {
    int x = 32768;                       // ldc
    MyObject myObj = new MyObject();     // new
    MyObject.staticVar = x;              // putstatic
    x = MyObject.staticVar;              // getstatic
    myObj.instanceVar = x;               // putfield
    x = myObj.instanceVar;               // getfield
    Object obj = myObj;
    if (obj instanceof MyObject) {       // instanceof
        myObj = (MyObject) obj;          // checkcast
        System.out.println(myObj.instanceVar);
    }
}
```

以上面这段为例，我们来说一下JVM对类与对象的实现机制

首先是`LDC`指令

## LDC

``` go
    int x = 32768;                       // ldc
```

ldc系列指令从运行时常量池中加载常量值,并把它推入操作数栈

ldc系列指令属于常量类指令,共3条。

1. 其中ldc和ldc_w指令用于加载int、float和字符串常量,java.lang.Class实例或者MethodType和MethodHandle实例。
2. ldc2_w指令用于加载long和double常量。
3. ldc 和ldc_w指令的区别仅在于操作数的宽度。


``` go
// Push item from run-time constant pool
type LDC struct{ base.Index8Instruction }

func (self *LDC) Execute(frame *rtda.Frame) {
	_ldc(frame, self.Index)
}

func _ldc(frame *rtda.Frame, index uint) {
	stack := frame.OperandStack()
	class := frame.Method().Class()
	c := class.ConstantPool().GetConstant(index)

	switch c.(type) {
	case int32:
		stack.PushInt(c.(int32))
	case float32:
		stack.PushFloat(c.(float32))
	case string:
		internedStr := heap.JString(class.Loader(), c.(string))
		stack.PushRef(internedStr)
	case *heap.ClassRef:
		classRef := c.(*heap.ClassRef) // 常量池的常量是类引用
		classObj := classRef.ResolvedClass().JClass() // 解析类引用
		stack.PushRef(classObj) // 将类对象入栈
	// case MethodType, MethodHandle
	default:
		panic("todo: ldc!")
	}
}
```

首先来回忆一下我们的JVM 运行时的结构：

``` go
/*
JVM
   Thread
      pc
      Stack
        Frame
           LocalVars
           OperandStack
  **/

```

其中，stack是一个抽象的栈，栈的元素是Frame，也就是函数栈帧

``` go
type Frame struct {
	lower *Frame
	localVars LocalVars
	operandStack *OperandStack
	thread *Thread
	method *heap.Method
	nextPC int // the next instruction after the call
}
```

函数执行的必要资源都在frame中。

### Method

函数栈帧绑定的函数对象

``` go
type Method struct {
	ClassMember
	maxStack uint // 最大栈深度
	maxLocals uint // 最大局部变量数
	code []byte // 代码源码
	exceptionTable ExceptionTable // 异常处理表
	lineNumberTable *classfile.LineNumberTableAttribute // 行号表
    // ...
	argSlotCount uint // 函数参数槽量
}
```

其中ClassMember是为了与Field的部分字段复用

``` go
type ClassMember struct {
	accessFlags    uint16 // 访问级别
	name           string // 类名
	descriptor     string // 描述符
    // ...
	class          *Class // 类指针
}
```

这时，再来看LDC指令的实现，获取了常量池的数据，然后压入操作数栈。

``` go
    // 从函数的类指针中获取与运行时常量池指针，并获取常量池中保存的数据: 数值或引用
	class := frame.Method().Class()
	c := class.ConstantPool().GetConstant(index)
```


### OperandStack

``` go
    // 获取操作数栈
	stack := frame.OperandStack()
    // ...
		stack.PushInt(c.(int32))
```

``` go
// 一个用数组实现的栈
type OperandStack struct {
	size uint
	slots []Slot
}
```

## NEW

``` java
    MyObject myObj = new MyObject();     // new
```

`NEW`的实现：

``` go
// Create new object
type NEW struct {
	base.Index16Instruction
}

func (self *NEW) Execute(frame *rtda.Frame) {
	cp := frame.Method().Class().ConstantPool()
	classRef := cp.GetConstant(self.Index).(*heap.ClassRef)
	class := classRef.ResolvedClass() // 用符号引用加载整个类信息
	// new指令触发构建类实例 但类还没有初始化 终止指令执行
	if !class.InitStarted() {
		frame.RevertNextPC() // 回置PC
		base.InitClass(frame.Thread(), class)
		return
	}

	// interface and abstract class can be instantced
	if class.IsInterface() || class.IsAbstract() {
		panic("java.lang.InstantiationError")
	}
	ref := class.NewObject() // 实例化
	frame.OperandStack().PushRef(ref)
}

```

``` go
	cp := frame.Method().Class().ConstantPool()
	classRef := cp.GetConstant(self.Index).(*heap.ClassRef)
	class := classRef.ResolvedClass() // 用符号引用加载整个类信息

```

首先，上面这段代码从方法所属的类的运行时常量池中获取`myObj`引用，接下来进行解引用。那么什么是引用呢？

``` go
type ClassRef struct {
	SymRef
}

// 1. 类符号引用
// 2. 字段符号引用
// 3. 方法符号引用
// 4. 接口方法符号引用
type SymRef struct {
	cp *ConstantPool // 符号引用所在的运行时常量池指针
	className string
	class *Class
}

```

类引用包含了一些可以索引到对应类的必要信息，通过这些必要信息我们可以获取该类的具体信息，这个过程就是解引用。

``` go
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
```

可以发现，我们在解引用的时候进行了对类的加载。简单叙述一下就是从classpath中找到对应的class文件，读取并解析其中的内容，**但类的初始化并不是这里进行的，而是在NEW指令中**。

还有一点，就是`d := self.cp.class 	c := d.loader.LoadClass(self.className)` 这两句中出现了两个类，他们有什么区别呢？

是这样的，为了支持多态，java允许`Super obj = new Sub()` 这样的实现，也就是说，classRef.className为`Sub`但`classRef.cp.class.`为`Super`，也就是说JVM使用了一个`Super`的常量池指针保存了一个`Sub`的类引进。

**c 为 Sub 实际的类，d为Super 保存该引用的类，可能是引用类本身 ,也可能是引用类的父类。这就解释了为什么后面我们需要检查d是否可以访问到c，只有Super的引用才能访问Sub的实例**

``` go
	// new指令触发构建类实例 但类还没有初始化 终止指令执行
	if !class.InitStarted() {
		frame.RevertNextPC() // 回置PC
		base.InitClass(frame.Thread(), class)
		return
	}
```

上面这段代码的意思大致就是：检查类是否已经初始化，如果没有就重新先去初始化，然后重新执行一次NEW指令。

``` go

	ref := class.NewObject() // 实例化
	frame.OperandStack().PushRef(ref)
```

上面这段代码是NEW指令的最后一部分，实例化该类的一个对象，然后压入函数栈帧的操作数栈。

``` go
func (self *Class) NewObject() *Object {
	return newObject(self)
}

type Object struct {
	class *Class
	data interface{}
	extra interface{} // 记录Object结构体实例的额外信息
}

func newObject(class *Class) *Object {
	return &Object{
		class: class,
		data: newSlots(class.instanceSlotCount), // 分配类实例内存
	}
}
```

这样，一个有类信息，分配了内存的类实例就创建好了。

## PUTSTATIC

``` java
    MyObject.staticVar = x;              // putstatic
```

`staticVar`是MyObject类的静态字段，不需要实例化就可以访问，这是如何实现的呢？

``` go
// putstatic指令给类的某个静态变量赋值
type PUT_STATIC struct {
	base.Index16Instruction
}

func (self *PUT_STATIC) Execute(frame *rtda.Frame) {
	currentMethod := frame.Method()
	currentClass := currentMethod.Class()
	cp := currentClass.ConstantPool()
	// 通过这个索引可以从当前类的运行时常量池中找到一个字段符号引用
	fieldRef := cp.GetConstant(self.Index).(*heap.FieldRef)
	field := fieldRef.ResolvedField()
	class := field.Class()
	// init class
	if !class.InitStarted() {
		frame.RevertNextPC()
		base.InitClass(frame.Thread(), class)
		return
	}

	if !field.IsStatic() {
		panic("java.lang.IncompatiableClassChangeError")
	}
	// 如果是final字段,则实际操作的是静态常量,只能在类初始化方法中给它赋值
	// 类初始化方法由编译器生成,名字是<clinit>
	if field.IsFinal() {
		if currentClass != class || currentMethod.Name() != "<clinit>" {
			panic("java.lang.IllegalAccessError")
		}
	}

	descriptor := field.Descriptor()
	slotId := field.SlotId()
	slots := class.StaticVars()
	stack := frame.OperandStack()

	switch descriptor[0] {
	case 'Z', 'B', 'C', 'S', 'I':
		slots.SetInt(slotId, stack.PopInt())
	case 'F':
		slots.SetFloat(slotId, stack.PopFloat())
	case 'J':
		slots.SetLong(slotId, stack.PopLong())
	case 'D':
		slots.SetDouble(slotId, stack.PopDouble())
	case 'L', '[':
		slots.SetRef(slotId, stack.PopRef())
	default:
		// TODO
	}
}
```

我们来一段一段的看：

``` go
	currentMethod := frame.Method()
	currentClass := currentMethod.Class()
	cp := currentClass.ConstantPool()
	// 通过这个索引可以从当前类的运行时常量池中找到一个字段符号引用
	fieldRef := cp.GetConstant(self.Index).(*heap.FieldRef)

```

以上代码通过函数栈帧，获取函数，进而获取函数的类，再用类的常量池指针找到指定的Field引用。

``` go
	field := fieldRef.ResolvedField()
	class := field.Class()
	// init class
	if !class.InitStarted() {
		frame.RevertNextPC()
		base.InitClass(frame.Thread(), class)
		return
	}
```

以上代码通过对Field引用的解引用获取该Field的具体信息，进而获取到它对应的类。字段的解引用实现如下：

``` go
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
```

这段解引用的代码比较长，但如果你看过我们上一篇文章就会发现其实很简单，总结一下就是FieldRef从创建之初就绑定了一个运行时常量池指针，这个指针所在的类可能是FieldRef的类，也可能是它的父类，在对FieldRef解引用后获取实际对应的类。再从这个类中寻找是否包含该Field，包括类本身、类实现的接口以及它的所有父类。

在通过FieldRef找到实际的Field后，通过Field尝试去初始化类。

``` go
	if !field.IsStatic() {
		panic("java.lang.IncompatiableClassChangeError")
	}
	// 如果是final字段,则实际操作的是静态常量,只能在类初始化方法中给它赋值
	// 类初始化方法由编译器生成,名字是<clinit>
	if field.IsFinal() {
		if currentClass != class || currentMethod.Name() != "<clinit>" {
			panic("java.lang.IllegalAccessError")
		}
	}
```

上面的代码会检查Field是否是Static以及Final,并进行对应的处理。

``` go
	descriptor := field.Descriptor()
	slotId := field.SlotId()
	slots := class.StaticVars()
	stack := frame.OperandStack()

	switch descriptor[0] {
	case 'Z', 'B', 'C', 'S', 'I':
		slots.SetInt(slotId, stack.PopInt())
	case 'F':
		slots.SetFloat(slotId, stack.PopFloat())
	case 'J':
		slots.SetLong(slotId, stack.PopLong())
	case 'D':
		slots.SetDouble(slotId, stack.PopDouble())
	case 'L', '[':
		slots.SetRef(slotId, stack.PopRef())
	default:
		// TODO
	}

```

以上代码就是实际上给静态变量赋值的过程了，找到该Field对应的SlotId，找到Field的类的Slots，并在对应的Slot设置从操作数栈顶弹出的数据。


## GETSTATIC

``` java
    x = MyObject.staticVar;              // getstatic
```

GETSTATIC的实现与PUTSTATIC高度相似，唯一不同的地方就是，这次是在找到FieldRef对应的Field后，将其类中对应Slot保存的数据压入操作数栈。

``` go
// Get static field from class
type GET_STATIC struct{ base.Index16Instruction }

func (self *GET_STATIC) Execute(frame *rtda.Frame) {
	cp := frame.Method().Class().ConstantPool()
	fieldRef := cp.GetConstant(self.Index).(*heap.FieldRef)
	field := fieldRef.ResolvedField()
    
    // ...

	descriptor := field.Descriptor()
	slotId := field.SlotId()
	slots := class.StaticVars()
	stack := frame.OperandStack()

	switch descriptor[0] {
	case 'Z', 'B', 'C', 'S', 'I':
		stack.PushInt(slots.GetInt(slotId))
	case 'F':
		stack.PushFloat(slots.GetFloat(slotId))
	case 'J':
		stack.PushLong(slots.GetLong(slotId))
	case 'D':
		stack.PushDouble(slots.GetDouble(slotId))
	case 'L', '[':
		stack.PushRef(slots.GetRef(slotId))
	default:
		// TODO
	}
}

```


## PUTFIELD

``` java
    myObj.instanceVar = x;               // putfield
```

``` go
type PUT_FIELD struct {
	base.Index16Instruction
}

func (self *PUT_FIELD) Execute(frame *rtda.Frame) {
	currentMethod := frame.Method()
	currentClass := currentMethod.Class()
	cp := currentClass.ConstantPool()
	fieldRef := cp.GetConstant(self.Index).(*heap.FieldRef)
	field := fieldRef.ResolvedField()

	if field.IsStatic() {
		panic("java.lang.IncompatibleClassChangeError")
	}
	if field.IsFinal() {
		if currentClass != field.Class() || currentMethod.Name() != "<init>" {
			panic("java.lang.IllegalAccessError")
		}
	}

	descriptor := field.Descriptor()
	slotId := field.SlotId()
	stack := frame.OperandStack()

	switch descriptor[0] {
	case 'Z', 'B', 'C', 'S', 'I':
		val := stack.PopInt()
		ref := stack.PopRef()
		if ref == nil {
			panic("java.lang.NullPointerException")
		}
		ref.Fields().SetInt(slotId, val)
    // ...
	default:
		// TODO
	}
}

```

其实可以发现，PUTFIELD的实现与PUTSTATIC很相似，下面的代码同样是从当前函数栈帧中获取当前函数的类，再找到对应的fieldRef并解引用，获取fieldRef对应的Field，检查其Static和Final并进行对应的处理。

``` go
	currentMethod := frame.Method()
	currentClass := currentMethod.Class()
	cp := currentClass.ConstantPool()
	fieldRef := cp.GetConstant(self.Index).(*heap.FieldRef)
	field := fieldRef.ResolvedField()

	if field.IsStatic() {
		panic("java.lang.IncompatibleClassChangeError")
	}
	if field.IsFinal() {
		if currentClass != field.Class() || currentMethod.Name() != "<init>" {
			panic("java.lang.IllegalAccessError")
		}
	}

```

不同点在于下面：这次我们不是直接去修改Class的StaticVars，而是去修改了当前实例的对应的内存

``` go
	descriptor := field.Descriptor()
	slotId := field.SlotId()
	stack := frame.OperandStack()

	switch descriptor[0] {
	case 'Z', 'B', 'C', 'S', 'I':
		val := stack.PopInt()
		ref := stack.PopRef()
		if ref == nil {
			panic("java.lang.NullPointerException")
		}
		ref.Fields().SetInt(slotId, val)
    // ...
	default:
		// TODO
	}

```

我们分开来说：

``` go
		val := stack.PopInt()
		ref := stack.PopRef()
```

`val`是我们将要赋予的值，而`ref`是一个`Object`，是当前类实例的引用，没错就是`this`指针，让我们来回顾一下`Object`的结构

``` go
type Object struct {
	class *Class
	data interface{}
	extra interface{} // 记录Object结构体实例的额外信息
}

func newObject(class *Class) *Object {
	return &Object{
		class: class,
		data: newSlots(class.instanceSlotCount), // 分配类实例内存
	}
}
```

`Object.data`是与`Class.instanceSlotCount`占用量相同的数据区。

``` go
func (self *Object) Fields() Slots {
	return self.data.(Slots)
}

func (self Slots) SetInt(index uint, val int32) {
	self[index].num = val
}
```

`ref.Fields().SetInt()`就是在Object的对应数据区设置val。


## GETFIELD

``` java
    x = myObj.instanceVar;               // getfield
```

``` go
type GET_FIELD struct {
	base.Index16Instruction
}

func (self *GET_FIELD) Execute(frame *rtda.Frame) {
	cp := frame.Method().Class().ConstantPool()
	fieldRef := cp.GetConstant(self.Index).(*heap.FieldRef)
	field := fieldRef.ResolvedField()

	if field.IsStatic() {
		panic("java.lang.IncompatibleClassChangeError")
	}

	stack := frame.OperandStack()
	ref := stack.PopRef() // this指针
	if ref == nil {
		panic("java.lang.NullPointerException")
	}

	descriptor := field.Descriptor()
	slotId := field.SlotId()
	slots := ref.Fields() // 当前实例所有的字段

	switch descriptor[0] {
    // ...
	case 'L', '[':
		stack.PushRef(slots.GetRef(slotId)) // 从指定的字段中获取数据 并压入操作数栈
	default:
		// TODO
	}
}
```

## INSTANCEOF



## CHECKCAST

``` java

```

### LocalVars

``` go
type LocalVars []Slot

type Slot struct {
	num int32
	ref *heap.Object
}
```

LocalVars用来表示局部变量表。从逻辑上来看， LocalVars实例就像一个数组，这个数组的每一个元素都足够容纳一个int、float或引用值(要放入double或者long值，需要相邻的两个元素)

