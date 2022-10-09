# 字节码指令

``` go
type Instruction interface {
	FetchOperands(reader *BytecodeReader)
	Execute(frame *rtda.Frame)
}
```

## NOP

``` go
type NOP struct {
	base.NoOperandsInstruction
}

func (self *NOP) Execute(frame *rtda.Frame) {
	// really do nothing
}
```

## XCONST_X系列

``` go
// Push null
type ACONST_NULL struct {
	base.NoOperandsInstruction
}

func (self *ACONST_NULL) Execute(frame *rtda.Frame) {
	frame.OperandStack().PushRef(nil)
}
```


``` go
// Push float
type FCONST_0 struct{ base.NoOperandsInstruction }

func (self *FCONST_0) Execute(frame *rtda.Frame) {
	frame.OperandStack().PushFloat(0.0)
}

type FCONST_1 struct{ base.NoOperandsInstruction }

func (self *FCONST_1) Execute(frame *rtda.Frame) {
	frame.OperandStack().PushFloat(1.0)
}

type FCONST_2 struct{ base.NoOperandsInstruction }

func (self *FCONST_2) Execute(frame *rtda.Frame) {
	frame.OperandStack().PushFloat(2.0)
}
```


## BIPUSH SIPUSH

``` go
// 从操作数中获取一个byte型整数 扩展成int型 然后推入栈顶
type BIPUSH struct {
	val int8
}

func (self *BIPUSH) FetchOperands(reader *base.BytecodeReader) {
	self.val = reader.ReadInt8()
}

func (self *BIPUSH) Execute(frame *rtda.Frame) {
	i := int32(self.val)
	frame.OperandStack().PushInt(i)
}


// 从操作数中获取一个short型整数 扩展成int型 然后推入栈顶
type SIPUSH struct {
	val int16
}

func (self *SIPUSH) FetchOperands(reader *base.BytecodeReader) {
	self.val = reader.ReadInt16()
}

func (self *SIPUSH) Execute(frame *rtda.Frame) {
	i := int32(self.val)
	frame.OperandStack().PushInt(i)
}
```


## LDC系列

从与运行时常量池加载变量

``` go
// ldc系列指令从运行时常量池中加载常量值,并把它推入操作数栈
//
// ldc系列指令属于常量类指令,共3条。
//
// 其中ldc和ldc_w指令用于加载int、float和字符串常量,
// java.lang.Class实例或者MethodType和MethodHandle实例。
//
// ldc2_w指令用于加载long和double常量。
//
// ldc 和ldc_w指令的区别仅在于操作数的宽度。

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


## XLOAD_X系列

从局部变量加载变量

``` go
// iload 指令的索引来自于操作数
type ILOAD struct {
	base.Index8Instruction
}

func (self *ILOAD) Execute(frame *rtda.Frame) {
	_iload(frame, self.Index)
}

// 其余指令的索引隐含于操作码中
type ILOAD_0 struct {
	base.NoOperandsInstruction
}
func (self *ILOAD_0) Execute(frame *rtda.Frame) {
	_iload(frame, 0)
}
// 加载指令从局部变量表中获取变量 然后推入操作数栈顶
func _iload(frame *rtda.Frame, index uint) {
	val := frame.LocalVars().GetInt(index)
	frame.OperandStack().PushInt(val)
}
```

## XSTORE_X系列

``` go
// Store int into local variable
type ISTORE struct{ base.Index8Instruction }

func (self *ISTORE) Execute(frame *rtda.Frame) {
	_istore(frame, uint(self.Index))
}

type ISTORE_0 struct{ base.NoOperandsInstruction }

func (self *ISTORE_0) Execute(frame *rtda.Frame) {
	_istore(frame, 0)
}

// 存储指令把变量从操作数栈顶弹出 然后存入局部变量表
func _istore(frame *rtda.Frame, index uint) {
	val := frame.OperandStack().PopInt()
	frame.LocalVars().SetInt(index, val)
}
```


## XALOAD系列

``` go
// Load reference from array
type AALOAD struct{ base.NoOperandsInstruction }

func (self *AALOAD) Execute(frame *rtda.Frame) {
	stack := frame.OperandStack()
	index := stack.PopInt()
	arrRef := stack.PopRef()

    // 非空检测
	checkNotNil(arrRef)
	refs := arrRef.Refs()
	checkIndex(len(refs), index)
	stack.PushRef(refs[index])
}
```

## XSTORE系列

``` go
// Store into reference array
type AASTORE struct{ base.NoOperandsInstruction }

func (self *AASTORE) Execute(frame *rtda.Frame) {
	stack := frame.OperandStack()
	ref := stack.PopRef()
	index := stack.PopInt()
	arrRef := stack.PopRef()

    // 非空检测
	checkNotNil(arrRef)
	refs := arrRef.Refs()
	checkIndex(len(refs), index)
	refs[index] = ref
}
```

## POP指令

``` go
type POP struct {
	base.NoOperandsInstruction
}

// pop int float ...
func (self *POP) Execute(frame *rtda.Frame) {
	stack := frame.OperandStack()
	stack.PopSlot()
}

type POP2 struct {
	base.NoOperandsInstruction
}

// pop double long
func (self *POP2) Execute(frame *rtda.Frame) {
	stack := frame.OperandStack()
	stack.PopSlot()
	stack.PopSlot()
}

```


## DUP系列

``` go
type DUP struct {
	base.NoOperandsInstruction
}

// dup 指令复制栈顶的单个变量
func (self *DUP) Execute(frame *rtda.Frame) {
	stack := frame.OperandStack()
	slot := stack.PopSlot()
	stack.PushSlot(slot)
	stack.PushSlot(slot)
}

// dup 指令复制栈顶的两个变量
type DUP2 struct {
	base.NoOperandsInstruction
}

func (self *DUP2) Execute(frame *rtda.Frame) {
	stack := frame.OperandStack()
	slot1 := stack.PopSlot()
	slot2 := stack.PopSlot()
	stack.PushSlot(slot2)
	stack.PushSlot(slot1)
	stack.PushSlot(slot2)
	stack.PushSlot(slot1)
}
```


## SWAP指令

``` go
type SWAP struct {
	base.NoOperandsInstruction
}

// 交换栈顶的两个变量
func (self *SWAP) Execute(frame *rtda.Frame) {
	stack := frame.OperandStack()
	slot1 := stack.PopSlot()
	slot2 := stack.PopSlot()
	stack.PushSlot(slot1)
	stack.PushSlot(slot2)
}
```



## ADD系列

``` go
// Int Add
type IADD struct {
	base.NoOperandsInstruction
}

func (self *IADD) Execute(frame *rtda.Frame) {
	stack := frame.OperandStack()
	v2 := stack.PopInt()
	v1 := stack.PopInt()
	result := v1 + v2
	stack.PushInt(result)
}
```

## SUB系列

``` go
// Subtract int
type ISUB struct{ base.NoOperandsInstruction }

func (self *ISUB) Execute(frame *rtda.Frame) {
	stack := frame.OperandStack()
	v2 := stack.PopInt()
	v1 := stack.PopInt()
	result := v1 - v2
	stack.PushInt(result)
}

```

## MULT系列

``` go
// Multiply int
type IMUL struct{ base.NoOperandsInstruction }

func (self *IMUL) Execute(frame *rtda.Frame) {
	stack := frame.OperandStack()
	v2 := stack.PopInt()
	v1 := stack.PopInt()
	result := v1 * v2
	stack.PushInt(result)
}
```

## DIV系列

``` go
// Divide int
type IDIV struct{ base.NoOperandsInstruction }

func (self *IDIV) Execute(frame *rtda.Frame) {
	stack := frame.OperandStack()
	v2 := stack.PopInt()
	v1 := stack.PopInt()
	if v2 == 0 {
		panic("java.lang.ArithmeticException: / by zero")
	}

	result := v1 / v2
	stack.PushInt(result)
}
```


## REM系列 (mod)

``` go
// Remainder Int
type IREM struct {
	base.NoOperandsInstruction
}

func (self *IREM) Execute(frame *rtda.Frame) {
	stack := frame.OperandStack()
	v2 := stack.PopInt()
	v1 := stack.PopInt()
	if v2 == 0 {
		panic("java.lang.ArithmeticException: / by zero")
	}
	reslut := v1 % v2
	stack.PushInt(reslut)
}
```

## NEG系列

``` go
// Negate int
type INEG struct{ base.NoOperandsInstruction }

func (self *INEG) Execute(frame *rtda.Frame) {
	stack := frame.OperandStack()
	val := stack.PopInt()
	stack.PushInt(-val)
}
```

## 各种位移

``` go
// Shitf Left Int
type ISHL struct {
	base.NoOperandsInstruction
}

func (self *ISHL) Execute(frame *rtda.Frame) {
	stack := frame.OperandStack()
	v2 := stack.PopInt()
	v1 := stack.PopInt()
	s := uint32(v2) & 0x1f // 取最后五位 2**5 = 32 足以表示位移
	result := v1 << s
	stack.PushInt(result)
}

// Int 算术右移 有符号右移
type ISHR struct {
	base.NoOperandsInstruction
}

func (self *ISHR) Execute(frame *rtda.Frame) {
	stack := frame.OperandStack()
	v2 := stack.PopInt()
	v1 := stack.PopInt()
	s := uint32(v2) & 0x1f
	result := v1 >> s
	stack.PushInt(result)
}

// 逻辑右移 无符号右移
type IUSHR struct {
	base.NoOperandsInstruction
}

func (self *IUSHR) Execute(frame *rtda.Frame) {
	stack := frame.OperandStack()
	v2 := stack.PopInt()
	v1 := stack.PopInt()
	s := uint32(v2) & 0x1f
	result := int32((uint32(v1) >> s))
	stack.PushInt(result)
}

// Shift Left Long
type LSHL struct {
	base.NoOperandsInstruction
}

func (self *LSHL) Execute(frame *rtda.Frame) {
	stack := frame.OperandStack()
	v2 := stack.PopInt()
	v1 := stack.PopInt()
	s := uint32(v2) & 0x3f
	result := v1 << s
	stack.PushInt(result)
}

// Shift Right Long
type LSHR struct {
	base.NoOperandsInstruction
}

func (self *LSHR) Execute(frame *rtda.Frame) {
	stack := frame.OperandStack()
	v2 := stack.PopInt()
	v1 := stack.PopInt()
	s := uint32(v2) & 0x3f
	result := v1 >> s
	stack.PushInt(result)
}

// Long 算术右移 有符号右移
type LUSHR struct {
	base.NoOperandsInstruction
}

func (self *LUSHR) Execute(frame *rtda.Frame) {
	stack := frame.OperandStack()
	v2 := stack.PopInt()
	v1 := stack.PopInt()
	s := uint32(v2) & 0x3f // 0x0011 1111 2**6=64
	result := int64(uint64(v1) >> s)
	stack.PushLong(result)
}


```


## IINC指令

``` go
// iinc 指令给局部变量表中的int变量增加常量值
//      局部变量表索引和常量值都由指令的操作数提供
type IINC struct {
	Index uint
	Const int32
}

func (self *IINC) FetchOperands(reader *base.BytecodeReader) {
	self.Index = uint(reader.ReadInt8())
	self.Const = int32(reader.ReadInt8())
}

func (self *IINC) Execute(frame *rtda.Frame) {
	localVars := frame.LocalVars()
	val := localVars.GetInt(self.Index)
	val += self.Const
	localVars.SetInt(self.Index, val)
}

```


## 强制类型转换

从操作栈弹出数据，用go的类型强制转换，再将其压入回栈

``` go
// Convert int to byte
type I2B struct{ base.NoOperandsInstruction }

func (self *I2B) Execute(frame *rtda.Frame) {
	stack := frame.OperandStack()
	i := stack.PopInt()
	b := int32(int8(i))
	stack.PushInt(b)
}

// Convert int to char
type I2C struct{ base.NoOperandsInstruction }

func (self *I2C) Execute(frame *rtda.Frame) {
	stack := frame.OperandStack()
	i := stack.PopInt()
	c := int32(uint16(i))
	stack.PushInt(c)
}

// Convert int to short
type I2S struct{ base.NoOperandsInstruction }

func (self *I2S) Execute(frame *rtda.Frame) {
	stack := frame.OperandStack()
	i := stack.PopInt()
	s := int32(int16(i))
	stack.PushInt(s)
}

// Convert int to long
type I2L struct{ base.NoOperandsInstruction }

func (self *I2L) Execute(frame *rtda.Frame) {
	stack := frame.OperandStack()
	i := stack.PopInt()
	l := int64(i)
	stack.PushLong(l)
}

// Convert int to float
type I2F struct{ base.NoOperandsInstruction }

func (self *I2F) Execute(frame *rtda.Frame) {
	stack := frame.OperandStack()
	i := stack.PopInt()
	f := float32(i)
	stack.PushFloat(f)
}

// Convert int to double
type I2D struct{ base.NoOperandsInstruction }

func (self *I2D) Execute(frame *rtda.Frame) {
	stack := frame.OperandStack()
	i := stack.PopInt()
	d := float64(i)
	stack.PushDouble(d)
}

```

## CMP系列

- LCMP 两个Long比较

``` go
type LCMP struct {
	base.NoOperandsInstruction
}

func (self *LCMP) Execute(frame *rtda.Frame) {
	stack := frame.OperandStack()
	v2 := stack.PopLong()
	v1 := stack.PopLong()
	if v1 > v2 {
		stack.PushInt(1)
	} else if v1 == v2 {
		stack.PushInt(0)
	} else {
		stack.PushInt(-1)
	}
}

```

- FCMP fcmpg和fcmpl指令用于比较float变量

``` go
// Compare float
type FCMPG struct{ base.NoOperandsInstruction }

func (self *FCMPG) Execute(frame *rtda.Frame) {
	_fcmp(frame, true)
}

type FCMPL struct{ base.NoOperandsInstruction }

func (self *FCMPL) Execute(frame *rtda.Frame) {
	_fcmp(frame, false)
}

func _fcmp(frame *rtda.Frame, gFlag bool) {
	stack := frame.OperandStack()
	v2 := stack.PopFloat()
	v1 := stack.PopFloat()
	if v1 > v2 {
		stack.PushInt(1)
	} else if v1 == v2 {
		stack.PushInt(0)
	} else if v1 < v2 {
		stack.PushInt(-1)
	} else if gFlag {
		stack.PushInt(1)
	} else {
		stack.PushInt(-1)
	}
}

```

- cmpg和dcmpl指令用来比较double变量, 这两条指令和fcmpg、fcmpl指令除了比较的变量类型不同之外，代码基本上完全一样，这里就不详细介绍了

## if<cond>指令

``` go
// if<cond> 指令操作数栈顶的int变量弹出 然后跟0进行比较 满足条件则跳转

type IFEQ struct {
	base.BranchInstruction
}

func (self *IFEQ) Execute(frame *rtda.Frame) {
	val := frame.OperandStack().PopInt()
	if val == 0 {
		base.Branch(frame, self.Offset)
	}
}

type IFNE struct {
	base.BranchInstruction
}

func (self *IFNE) Execute(frame *rtda.Frame) {
	val := frame.OperandStack().PopInt()
	if val != 0 {
		base.Branch(frame, self.Offset)
	}
}
```

其中的base.Branch的实现如下，使用pc+offset实现了代码的跳转

``` go
func Branch(frame *rtda.Frame, offset int) {
	pc := frame.Thread().PC()
	nextPC := pc + offset
	frame.SetNextPC(nextPC)
}
```

## if_icmp<cond>指令
比较两个Int是否相同

``` go
// Branch if int comparison succeeds
type IF_ICMPEQ struct{ base.BranchInstruction }

func (self *IF_ICMPEQ) Execute(frame *rtda.Frame) {
	if val1, val2 := _icmpPop(frame); val1 == val2 {
		base.Branch(frame, self.Offset)
	}
}

type IF_ICMPNE struct{ base.BranchInstruction }

func (self *IF_ICMPNE) Execute(frame *rtda.Frame) {
	if val1, val2 := _icmpPop(frame); val1 != val2 {
		base.Branch(frame, self.Offset)
	}
}

func _icmpPop(frame *rtda.Frame) (val1, val2 int32) {
	stack := frame.OperandStack()
	val2 = stack.PopInt()
	val1 = stack.PopInt()
	return
}
```


## if_acmp<cond>指令

比较两个引用是否相同

``` go
// Branch if reference comparison succeeds
type IF_ACMPEQ struct{ base.BranchInstruction }

func (self *IF_ACMPEQ) Execute(frame *rtda.Frame) {
	if _acmp(frame) {
		base.Branch(frame, self.Offset)
	}
}

type IF_ACMPNE struct{ base.BranchInstruction }

func (self *IF_ACMPNE) Execute(frame *rtda.Frame) {
	if !_acmp(frame) {
		base.Branch(frame, self.Offset)
	}
}

func _acmp(frame *rtda.Frame) bool {
	stack := frame.OperandStack()
	ref2 := stack.PopRef()
	ref1 := stack.PopRef()
	return ref1 == ref2 // todo
}
```


##  GOTO

goto指令进行无条件跳转

``` go
func (self *GOTO) Execute(frame *rtda.Frame) {
    base.Branch(frame, self.Offset)
}
```


## SWITCH

Java语言中的switch-case语句有两种实现方式：

1. 如果case值可以编码成一个索引表，则实现成tableswitch指令

``` java
int chooseNear(int i) {
    switch (i) {
      case 0:   return 0;
      case 1:   return 1;
      case 2:   return 2;
      default:  return -1;
    }
}
```

先从操作数栈中弹出一个int变量，然后看它是否在low和high给定的范围之内。如果在，则从jumpOffsets表中查出偏移量进行跳转，否则按照defaultOffset跳转

``` go
// 连续 case
type TABLE_SWITCH struct {
	defalutOffset int32
	low int32
	high int32
	jumpOffsets []int32
}

func (self *TABLE_SWITCH) FetchOperands(reader *base.BytecodeReader) {
	reader.SkipPadding()
	self.defalutOffset = reader.ReadInt32()
	self.low = reader.ReadInt32()
	self.high = reader.ReadInt32()
	jumpOffsetCount := self.high - self.low + 1 // 获取跳转
	self.jumpOffsets = reader.ReadInt32s(jumpOffsetCount) // 获取跳转范围
}

func (self *TABLE_SWITCH) Execute(frame *rtda.Frame) {
	index := frame.OperandStack().PopInt()

	var offset int
	if index >= self.low && index <= self.high {
		offset = int(self.jumpOffsets[index-self.low])
	} else {
		offset = int(self.defalutOffset)
	}

	base.Branch(frame, offset)
}

```

2. 否则实现成lookupswitch指令

``` java
int chooseFar(int i) {
    switch (i) {
      case -100: return -1;
      case 0:     return   0;
      case 100:   return   1;
      default:    return -1;
    }
}
```

matchOffsets有点像Map，它的key是case值，value是跳转偏移量。Execute()方法先从操作数栈中弹出一个int变量，然后用它查找matchOffsets，看是否能找到匹配的key。如果能，则按照value给出的偏移量跳转，否则按照defaultOffset跳转

``` go

// 不连续 case
type LOOKUP_SWITCH struct {
	defalutOffset int32
	npairs int32
	matchOffsets []int32
}

func (self *LOOKUP_SWITCH) FetchOperands(reader *base.BytecodeReader) {
	reader.SkipPadding()
	self.defalutOffset = reader.ReadInt32()
	self.npairs = reader.ReadInt32()
	self.matchOffsets = reader.ReadInt32s(self.npairs * 2) // first:key second:value
}

func (self *LOOKUP_SWITCH) Execute(frame *rtda.Frame) {
	key := frame.OperandStack().PopInt()
	for i := int32(0);i < self.npairs * 2;i += 2{
		if self.matchOffsets[i] == key {
			offset := self.matchOffsets[i+1]
			base.Branch(frame, int(offset))
			return
		}
	}
	base.Branch(frame, int(self.defalutOffset))
}

```





## 扩展指令

### wide指令

加载类指令、存储类指令、ret指令和iinc指令需要按索引访问局部变量表，索引以uint8的形式存在字节码中。对于大部分方法来说，局部变量表大小都不会超过256，所以用一字节来表示索引就够了。但是如果有方法的局部变量表超过这限制呢？Java虚拟机规范定义了wide指令来扩展前述指令

``` go
type WIDE struct {
	modifiedInstruction base.Instruction
}

func (self *WIDE) FetchOperands(reader *base.BytecodeReader) {
	opcode := reader.ReadUint8()
	switch opcode {
	case 0x15:
		inst := &loads.ILOAD{}
		inst.Index = uint(reader.ReadUint16())
		self.modifiedInstruction = inst
	case 0x16:
		inst := &loads.LLOAD{}
		inst.Index = uint(reader.ReadUint16())
		self.modifiedInstruction = inst
	case 0x17:
		inst := &loads.FLOAD{}
		inst.Index = uint(reader.ReadUint16())
		self.modifiedInstruction = inst
	case 0x18:
		inst := &loads.DLOAD{}
		inst.Index = uint(reader.ReadUint16())
		self.modifiedInstruction = inst
	case 0x19:
		inst := &loads.ALOAD{}
		inst.Index = uint(reader.ReadUint16())
		self.modifiedInstruction = inst
	case 0x36:
		inst := &stores.ISTORE{}
		inst.Index = uint(reader.ReadUint16())
		self.modifiedInstruction = inst
	case 0x37:
		inst := &stores.LSTORE{}
		inst.Index = uint(reader.ReadUint16())
		self.modifiedInstruction = inst
	case 0x38:
		inst := &stores.FSTORE{}
		inst.Index = uint(reader.ReadUint16())
		self.modifiedInstruction = inst
	case 0x39:
		inst := &stores.DSTORE{}
		inst.Index = uint(reader.ReadUint16())
		self.modifiedInstruction = inst
	case 0x3a:
		inst := &stores.ASTORE{}
		inst.Index = uint(reader.ReadUint16())
		self.modifiedInstruction = inst
	case 0x84:
		inst := &math.IINC{}
		inst.Index = uint(reader.ReadUint16())
		inst.Const = int32(reader.ReadInt16())
		self.modifiedInstruction = inst
	case 0xa9: // ret
		panic("Unsupported opcode: 0xa9!")
	}

}

func (self *WIDE) Execute(frame *rtda.Frame) {
	self.modifiedInstruction.Execute(frame)
}

```

### IFNULL指令

根据引用是否是null进行跳转，ifnull和ifnonnull指令把栈顶的引用弹出

``` go
type IFNULL struct {
	base.BranchInstruction
}

func (self *IFNULL) Execute(frame *rtda.Frame) {
	ref := frame.OperandStack().PopRef()
	if ref == nil {
		base.Branch(frame, self.Offset)
	}
}

type IFNONNULL struct {
	base.BranchInstruction
}

func (self *IFNONNULL) Execute(frame *rtda.Frame) {
	ref := frame.OperandStack().PopRef()
	if ref != nil {
		base.Branch(frame, self.Offset)
	}
}
```

### GOTO_W 指令

``` go
type GOTO_W struct {
	offset int
}

func (self *GOTO_W) FetchOperands(reader *base.BytecodeReader) {
	self.offset = int(reader.ReadInt32())
}

func (self *GOTO_W) Execute(frame *rtda.Frame) {
	base.Branch(frame, self.offset)
}
```

## RETURN

直接弹出即将退出当前函数栈帧

``` go
// Return void from method
type RETURN struct {
	base.NoOperandsInstruction
}
func (self *RETURN) Execute(frame *rtda.Frame) {
	frame.Thread().PopFrame()
}
```

弹出即将退出当前函数栈帧，弹出该栈栈顶的数据后，主线程的栈顶函数栈帧将这些数据再次压栈

``` go

// Return int from method
type IRETURN struct {
	base.NoOperandsInstruction
}
func (self *IRETURN) Execute(frame *rtda.Frame) {
	thread := frame.Thread()
	currentFrame := thread.PopFrame()
	invokerFrame := thread.TopFrame()
	val := currentFrame.OperandStack().PopInt()
	invokerFrame.OperandStack().PushInt(val)
}
```
