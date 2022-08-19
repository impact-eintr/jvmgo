package classfile

/*
Deprecated_attribute {
   u2 attribute_name_index;
   u4 attribute_length;
}
**/
// @Deprecated 被废弃的 注解
type DeprecatedAttribute struct {
	MarkerAttribute
}

/*
Deprecated_attribute {
   u2 attribute_name_index;
   u4 attribute_length;
}
**/
// @Synthetic 源文件中不存在的 由编译器生成的类成员
type SyntheticAttribute struct {
	MarkerAttribute
}

type MarkerAttribute struct {}

func (self *MarkerAttribute) readInfo(reader *ClassReader) {
	// read nothing
}
