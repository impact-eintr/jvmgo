package classfile

/*
ConstantValue_attribute {
   u2 attribute_name_index;
   u4 attribute_length;
   u2 constantvalue_index;
}
*/
// 定长属性 指挥出现在field_info结构中 用于表示常量表达式的值
type ConstantValueAttribute struct {
	constantValueIndex uint16
}

func (self *ConstantValueAttribute) readInfo(reader *ClassReader) {
	self.constantValueIndex = reader.readUint16()
}

func (self *ConstantValueAttribute) ConstantValueIndex() uint16 {
	return self.constantValueIndex
}
