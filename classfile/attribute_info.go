package classfile

/*
attribute_info {
   u2 attribyte_name_index;
   u4 attribute_length;
   u1 info[attribute_length];
}
**/
type AttributeInfo interface {
	readInfo(reader *ClassReader)
}

func readAttributes(reader *ClassReader, cp ConstantPool) []AttributeInfo {
	attributesCount := reader.readUint16()
	attributes := make([]AttributeInfo, attributesCount)
	for i := range attributes {
		attributes[i] = readAttribute(reader, cp)
	}
	return attributes
}

func readAttribute(reader *ClassReader, cp ConstantPool) AttributeInfo {
	attrNameIndex := reader.readUint16()
	attrName := cp.getUtf8(attrNameIndex)
	attrLen := reader.readUint32()
	attrInfo := newAttributeInfo(attrName, attrLen, cp)
	attrInfo.readInfo(reader)
	return attrInfo
}

func newAttributeInfo(attrName string, attrLen uint32,
	cp ConstantPool) AttributeInfo {
	switch attrName {
	// Code是变长属性，只存在于method_info结构中。Code属性中存放字节码等方法相关信息
	case "Code":
		return &CodeAttribute{cp: cp}
	// ConstantValue是定长属性，只会出现在field_info结构中，用于表示常量表达式的值
	case "ConstantValue":
    return &ConstantValueAttribute{}
	//  Deprecated，仅起标记作用，不包含任何数据
  case "Deprecated":
    return &DeprecatedAttribute{}
	// Exceptions是变长属性，记录方法抛出的异常表
  case "Exceptions":
    return &ExceptionsAttribute{}
	// LineNumberTable属性表存放方法的行号信息
  case "LineNumberTable":
    return &LineNumberTableAttribute{}
	// LocalVariableTable属性表中存放方法的局部变量信息
  case "LocalVariableTable":
    return &LocalVariableTableAttribute{}
	// SourceFile是可选定长属性，只会出现在ClassFile结构中，用于指出源文件名
  case "SourceFile":
    return &SourceFileAttribute{cp: cp}
	//  Synthetic，仅起标记作用，不包含任何数据
  case "Synthetic":
    return &SyntheticAttribute{}
  default:
    return &UnparsedAttribute{attrName, attrLen, nil}
	}
}
