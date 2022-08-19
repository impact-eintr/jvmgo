package classfile

/*
 SourceFile_attribute {
   u2 attribute_name_index;
   u4 attribute_length;
   u2 sourcefile_index;
 }
 */
// 用于指出源文件名 具体值保存在常量池中
type SourceFileAttribute struct {
	cp ConstantPool
	sourceFileIndex uint16
}

func (self *SourceFileAttribute) readInfo(reader *ClassReader) {
	self.sourceFileIndex = reader.readUint16()
}

func (self *SourceFileAttribute) FileName() string {
	return self.cp.getUtf8(self.sourceFileIndex)
}
