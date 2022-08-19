package classfile

/*
Exception_attribute {
   u2 attribute_name_index;
   u4 attribute_length;
   u2 number_of_exceptions;
   u2 exception_index_table[number_of_exceptions];
}
*/
// Exceptions 是变长属性 记录方法抛出的异常表
type ExceptionsAttribute struct {
	exceptionIndexTable []uint16
}

func (self *ExceptionsAttribute) readInfo(reader *ClassReader) {
	self.exceptionIndexTable = reader.readUint16s()
}

func (self *ExceptionsAttribute) ExceptionIndexTable() []uint16 {
	return self.exceptionIndexTable
}
