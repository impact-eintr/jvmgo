package classfile

/*
	 Code_attribute {
	 u2 attribute_name_index;
	 u4 attribute_length;
	 u2 max_stack;
	 u2 max_locals;
	 u4 code_length;
	 u1 code[code_length];
	 u2 exception_table_length;
	 {
			 u2 start_pc;
			 u2 end_pc;
			 u2 handler_pc;
			 u2 catch_type;
	 } exception_table[exception_table_length];
	 u2 attributes_count;
	 attribute_info attributes[attributes_count];
	 }
*/
// Code 是变长属性 只存在于method_info结构中
// Code属性中存放字节码等方法相关信息
type CodeAttribute struct {
	cp ConstantPool
	maxStack uint16 // 操作数栈的最大深度
	maxLocals uint16 // 局部变量表大小
	code []byte
	exceptionTable []*ExceptionTableEntry // 异常处理表
	attributes []AttributeInfo
}

func (self *CodeAttribute) readInfo(reader *ClassReader) {
	self.maxStack = reader.readUint16()
	self.maxLocals = reader.readUint16()
	codeLength := reader.readUint32()
	self.code = reader.readBytes(codeLength)
	self.exceptionTable = readExceptionTable(reader)
	self.attributes = readAttributes(reader, self.cp)
}

func (self *CodeAttribute) MaxStack() uint {
	return uint(self.maxStack)
}

func (self *CodeAttribute) MaxLocals() uint {
	return uint(self.maxLocals)
}

func (self *CodeAttribute) Code() []byte {
	return self.code
}

func (self *CodeAttribute) ExceptionTable() []*ExceptionTableEntry {
	return self.exceptionTable
}

func (self *CodeAttribute) LineNumberTableAttribute() *LineNumberTableAttribute {
	for _, attrInfo := range self.attributes {
		switch attrInfo.(type) {
		case *LineNumberTableAttribute:
			return attrInfo.(*LineNumberTableAttribute)
		}
	}
	return nil
}

type ExceptionTableEntry struct {
	startPc uint16
	endPc uint16
	handlePc uint16
	catchType uint16
}

func readExceptionTable(reader *ClassReader) []*ExceptionTableEntry {
	exceptionTableLength := reader.readUint16()
	exceptionTable := make([]*ExceptionTableEntry, exceptionTableLength)
	for i := range exceptionTable {
		exceptionTable[i] = &ExceptionTableEntry{
			startPc: reader.readUint16(),
			endPc: reader.readUint16(),
			handlePc: reader.readUint16(),
			catchType: reader.readUint16(),
		}
	}
	return exceptionTable
}

func (self *ExceptionTableEntry) StartPc() uint16 {
	return self.startPc
}

func (self *ExceptionTableEntry) EndPc() uint16 {
	return self.endPc
}

func (self *ExceptionTableEntry) HandlePc() uint16 {
	return self.handlePc
}

func (self *ExceptionTableEntry) CatchType() uint16 {
	return self.catchType
}
