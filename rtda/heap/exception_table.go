package heap

import "jvm/classfile"

// 异常处理是通过try-catch句实现的
// 异常处理表是Code属性的一部分,它记录了方法是否有能力处理某种异常
// 异常处理表的每一项都包含3个信息
// 1. 处理哪部分代码抛出的异常
// 2. 哪类异常
// 3. 异常处理代码在哪里

// void catchOne() {
//   try {
//     tryItOut();
//   } catch (TestExc e) {
//     handleExc(e);
//   }
// }

// 1 aload_0 // 把局部变量0(this)推入操作数栈顶
// 2 invokevirtual #4 // 调用tryItOut()方法
// 4 goto 13 // 如果try{} 没有抛出异常,直接执行return指令
// 7 astore_1 // 否则,异常对象引用在操作数栈顶,把它弹出,并放入局部变量1
// 8 aload_0 // 把this推入栈顶(将作为handleExc() 方法的参数0)
// 9 aload_1 // 把异常对象引用推入栈顶(将作为handleExc() 方法的参数1)
// 10 invokevirtual #5 // 调用handleExc() 方法
// 13 return // 方法返回

// 当tryItOut()方法通过athrow指令抛出TestExc异常时,Java虚拟
// 机首先会查找tryItOut()方法的异常处理表,看它能否处理该异常。
// 如果能,则跳转到相应的字节码开始异常处理。假设tryItOut()方法
// 无法处理异常,Java虚拟机会进一步查看它的调用者,也就是
// catchOne()方法的异常处理表。catchOne()方法刚好可以处理
// TestExc异常,使catch{}块得以执行。
// 假设catchOne()方法也无法处理TestExc异常,Java虚拟机会继
// 续查找catchOne()的调用者的异常处理表。这个过程会一直继续下
// 去,直到找到某个异常处理项,或者到达Java虚拟机栈的底部。

type ExceptionTable []*ExceptionHandler

type ExceptionHandler struct {
	// 具体来说,start_pc和end_pc可以锁定一部分字节码
	// 这部分字节码对应某个可能抛出异常的try{}代码块
	startPc int
	endPC int

	// 如果位于start_pc和end_pc之间的指令抛出异常x,且x是X(或者X的子类)的实例,
	// handler_pc就指出负责异常处理的catch{}块在哪里。
	handlerPc int

	// catch_type是个索引,通过它可以从运行时常量池中查到一个类符号引用,
	// 解析后的类是个异常类
	catchType *ClassRef
}

func newExcetionTable(entries []*classfile.ExceptionTableEntry,
	cp *ConstantPool) ExceptionTable{
	table := make([]*ExceptionHandler, len(entries))
	for i, entry := range entries {
		table[i] = &ExceptionHandler{
			startPc: int(entry.StartPc()),
			endPC: int(entry.EndPc()),
			handlerPc: int(entry.HandlePc()),
			catchType: getCatchType(uint(entry.CatchType()), cp),
		}
	}

	return table
}

// 我们知道0是无效的常量池索引,但是在这里0并非表示catch-none,而是表示catch-all
func getCatchType(index uint, cp *ConstantPool) *ClassRef {
	if index == 0 {
		return nil // catch all
	}
	return cp.GetConstant(index).(*ClassRef)
}

func (self ExceptionTable) findExceptionHandler(exClass *Class, pc int) *ExceptionHandler {
	for _, handler := range self {
		// 如果位于start_pc和end_pc之间的指令抛出异常x,且x是X(或者X的子类)的实例
		if pc >= handler.startPc && pc < handler.endPC {
			if handler.catchType == nil {
				return handler
			}
			catchClass := handler.catchType.ResolvedClass()
			if catchClass == exClass || catchClass.IsSuperClassOf(exClass) {
				return handler
			}
		}
	}
	return nil
}
