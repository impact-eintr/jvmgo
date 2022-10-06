# jvmgo
jvm implement by golang

在加载一个类之前要先加载它的超类，也就是java.lang.Object

多线程共享的内存区域主要存放两种数据：类数据和类实例（对象）。对象数据存放在Heap中，类数据存放在方法区中。堆由垃圾收集器定时清理，类数据包括字段和方法信息、方法的字节码、运行时常量池等等。从逻辑上讲，其实方法区也是堆的一部分。

线程私有的运行是数据区用于辅助执行java字节码。每个线程都有自己的pc寄存器和java虚拟机栈。

JVM Stack 又由 Stack Frame 构成，栈帧中保存方法执行的状态，包括局部变量表和操作数栈(Operand Stack)等。在任一时刻，某一线程肯定是在执行某个方法。这个方法叫做该线程的当前方法；执行该方法的帧叫做线程的当前帧；声明该方法的类叫做当前类。

如果当前方法是java方法。则pc寄存器中存发当前正在执行的java虚拟机指令的地址，否则，当前方法是本地方法，pc寄存器中的值没有明确定义。


Java虚拟机规范对于运行时数据区的规定是相当宽松的。以堆为例:堆可以是连续空间,也可以不连续。堆的大小可以固定,也可以在运行时按需扩展。虚拟机实现者可以使用任何垃圾回收算法管理堆,甚至完全不进行垃圾收集也是可以的。由于Go本身也有垃圾回收功能,所以可以直接使用Go的堆和垃圾收集器,这大大简化了我们的工作。

Java虚拟机可以操作两类数据:基本类型(primitive type)和引用类型(reference type)。基本类型的变量存放的就是数据本身,引用类型的变量存放的是对象引用,真正的对象数据是在堆里分配的。这里所说的变量包括类变量(静态字段)、实例变量(非静态字段)、数组元素、方法的参数和局部变量,等等。



基本类型可以进一步分为布尔类型(boolean type)和数字类型(numeric type),数字类型又可以分为整数类型(integral type)和浮点数类型(floating-point type)。引用类型可以进一步分为3种:类类型、接口类型和数组类型。类类型引用指向类实例,数组类型引用指向数组实例,接口类型引用指向实现了该接口的类或数组实例。引用类型有一个特殊的值——null,表示该引用不指向任何对象。

接下来，我们按照函数的执行顺序追踪一下JVM的实现，我们需要向计算机一样压很多层栈。当然我会保存一些上下文，免得阅读时读到一半又要跳回来。

代码详见[https://github.com/impact-eintr/jvmgo](https://github.com/impact-eintr/jvmgo)

## 主函数

``` go
func main() {
	cmd := parseCmd()
	if cmd.versionFlag {
		fmt.Println("version: v0.0.1")
	} else if cmd.helpFlag || cmd.class == "" {
		printUsage()
	} else {
		newJVM(cmd).start()
	}
}
```


### 命令行解析

``` go
type Cmd struct {
	helpFlag bool
	versionFlag bool
	verboseClassFlag bool
	verboseInstFlag bool
	cpOption string
	XjreOption string
	class string
	args []string
}
```

- 解析命令

``` go
func parseCmd() *Cmd {
	cmd := &Cmd{}

	flag.Usage = printUsage
	flag.BoolVar(&cmd.helpFlag, "help", false, "print help message")
	flag.BoolVar(&cmd.helpFlag, "?", false, "print help message")
	flag.BoolVar(&cmd.versionFlag, "version", false, "print version and exit")
	flag.BoolVar(&cmd.verboseClassFlag, "verbose:class", false, "enable verbose output")
	flag.BoolVar(&cmd.verboseInstFlag, "verbose:inst", false, "enable verbose output")
	flag.StringVar(&cmd.cpOption, "classpath", "", "classpath")
	flag.StringVar(&cmd.cpOption, "cp", "", "classpath")
	flag.StringVar(&cmd.XjreOption, "Xjre", "", "path to jre")
	flag.Parse()

	args := flag.Args()
	if (len(args) > 0) {
		cmd.class = args[0]
		cmd.args = args[1:]
	}

	return cmd
}
```

捕获对应的字符，绑定到Cmd的字段。


### 构造虚拟机

``` go
		newJVM(cmd).start()
```

- 虚拟机对象

``` go
type JVM struct {
	cmd *Cmd
	classLoader *heap.ClassLoader
	mainThread *rtda.Thread
}
```

1. cmd cli解析器
2. classLoader 类加载器
3. mainThread 运行时主线程

- 构造JVM对象

``` go
func newJVM(cmd *Cmd) *JVM {
	cp := classpath.Parse(cmd.XjreOption, cmd.cpOption)
	classLoader := heap.NewClassLoader(cp, cmd.verboseClassFlag)
	return &JVM{
		cmd: cmd,
		classLoader: classLoader,
		mainThread: rtda.NewThread(),
	}
}
```


#### classpath
首先来看classpath对象

``` go
type Classpath struct {
	bootClasspath Entry
	extClasspath Entry
	userClasspath Entry
}
```

- `Entry` 是一个 interface

``` go
type Entry interface {
	readClass(className string) ([]byte, Entry, error)
	String() string
}
```

`readClass` 是不同类型classpath对应的读取class对象的实现方法。

`String` 不同类型classpath对应的路径

``` go
func newEntry(path string) Entry {
	if strings.Contains(path, pathListSeparator) {
		return newCompositeEntry(path)
	}

	if strings.HasSuffix(path, "*") {
		return newWildcardEntry(path)
	}

	if strings.HasSuffix(path, ".jar") || strings.HasSuffix(path, ".JAR") ||
		strings.HasSuffix(path, ".zip") || strings.HasSuffix(path, ".ZIP") {
		return newZipEntry(path)
	}

	return newDirEntry(path)
}
```

可以看到有4种Entry
1. WildcardEntry 路径中包含通配符*

``` go
func newWildcardEntry(path string) CompositeEntry {
	baseDir := path[:len(path)-1]
	compositeEntry := []Entry{}

	// 对某个路径下的所有文件执行此函数
	walkFn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && path != baseDir {
			return filepath.SkipDir
		}
		// 对jar包构建Entry
		if strings.HasSuffix(path, ".jar") || strings.HasSuffix(path, ".JAR") {
			jarEntry := newZipEntry(path)
			compositeEntry = append(compositeEntry, jarEntry)
		}
		return nil
	}

	filepath.Walk(baseDir, walkFn)

	return compositeEntry
}
```

2. ZipEntry 路径中包含压缩包

``` go
type ZipEntry struct {
	absPath string
	zipRC *zip.ReadCloser
}

func newZipEntry(path string) *ZipEntry {
	absPath, err := filepath.Abs(path)
	if err != nil {
		panic(err)
	}
	return &ZipEntry{absPath: absPath, zipRC: nil}
}

func (self *ZipEntry) readClass(className string) ([]byte, Entry, error) {
	if self.zipRC == nil {
		err := self.openJar() // 打开jar包
		if err != nil {
			return nil, nil, errors.New("class not found: " + className)
		}
	}

	classFile := self.findClass(className) // 从已经打开的jar包中寻找对应的class文件
	if classFile == nil {
		return nil, nil, errors.New("class not found: " + className)
	}

	data, err := readClass(classFile) // 读取class文件的内容
	return data, self, err
}
```


3. DirEntry 普通的目录

``` go
type DirEntry struct {
	absDir string
}

func newDirEntry(path string) *DirEntry {
	absDir, err := filepath.Abs(path)
	if err != nil {
		panic(err)
	}
	return &DirEntry{absDir: absDir}
}


func (self *DirEntry) readClass(className string) ([]byte, Entry, error) {
	fileName := filepath.Join(self.absDir, className)
	data, err := ioutil.ReadFile(fileName)
	return data, self, err
}
```

4. CompositeEntry 路径中包含 `;`对分割出的路径中的每一部分进行构造

``` go
type CompositeEntry []Entry

func newCompositeEntry(pathList string) CompositeEntry {
	CompositeEntry := []Entry{}

	for _, path := range strings.Split(pathList, pathListSeparator) {
		entry := newEntry(path)
		CompositeEntry = append(CompositeEntry, entry)
	}
	return CompositeEntry
}

func (self CompositeEntry) readClass(className string) ([]byte, Entry, error) {
	for _, entry := range self {
		data, from, err := entry.readClass(className)
		if err == nil {
			return data, from, nil
		}
	}
	return nil, nil, errors.New("class not found: " + className)
}
```

- classpath的解析 (现在，我们在构造JVM对象，我们需要先构造一个classpath对象)

**Classpath结构体有三个字段，分别存放三种类路径。Parse()函数使用-Xjre选项解析启动类路径和扩展类路径，使用-classpath/-cp选项解析用户类路径**

``` go
func Parse(jreOption, cpOption string) *Classpath {
	cp := &Classpath{}
	cp.parseBootAndExtClasspath(jreOption)
	cp.parseUserClasspath(cpOption)
	return cp
}
```


**优先使用用户输入的-Xjre选项作为jre目录。如果没有输入该选项，则在当前目录下寻找jre目录。如果找不到，尝试使用JAVA_HOME环境变量**
``` go
func (self *Classpath) parseBootAndExtClasspath(jreOption string) {
	jreDir := getJreDir(jreOption)

	// jre/lib*
	jreLibPath := filepath.Join(jreDir, "lib", "*")
	self.bootClasspath = newWildcardEntry(jreLibPath)
	// jre/lib/ext/*
	jreExtPath := filepath.Join(jreDir, "lib", "ext","*")
	self.extClasspath = newWildcardEntry(jreExtPath)
}

func getJreDir(jreOption string) string {
	if jreOption != "" && exists(jreOption) {
		return jreOption
	}
	if exists("./jre") {
		return "./jre"
	}
	if jh := os.Getenv("JAVA_HOME"); jh != "" {
		return filepath.Join(jh, "jre")
	}
	panic("Can not find jre folder")
}
```

**解析用户classPath**
``` go
func (self *Classpath) parseUserClasspath(cpOption string) {
	if cpOption == "" {
		cpOption = "."
	}
	self.userClasspath = newEntry(cpOption)
}
```


**如果用户没有提供-classpath/-cp选项，则使用当前目录作为用户类路径。ReadClass()方法依次从启动类路径、扩展类路径和用户类路径中搜索class文件**

``` go
func (self *Classpath) ReadClass(className string) ([]byte, Entry, error) {
	className = className + ".class"
	if data, entry, err := self.bootClasspath.readClass(className); err == nil {
		return data, entry, nil
	}

	if data, entry, err := self.extClasspath.readClass(className); err == nil {
		return data, entry, nil
	}
	return self.userClasspath.readClass(className)
}
```

#### classLoader 类加载器对象 (现在我们在构造JVM对象，构造好classpath对象后，我们使用它来构造类加载器对象,注意：这个函数设计的内容非常多)

``` go
// 类加载器
type ClassLoader struct {
	cp       *classpath.Classpath
	verboseFlag bool
	classMap map[string]*Class
}
```

Classpath上面我们刚讲过，其实可以理解为classLoader的数据源，verboseFlag是一个调试用的标志，可以不细究，classMap是类加载器中的数据缓存一样的存在。

``` go
// 构造函数
func NewClassLoader(cp *classpath.Classpath, verboseFlag bool) *ClassLoader {
	loader := &ClassLoader{
		cp:       cp,
		verboseFlag: verboseFlag,
		classMap: make(map[string]*Class),
	}

	loader.loadBasicClasses()
	loader.loadPrimitiveClasses()
	return loader
}
```

classLoader的构造函数调用了一下两个函数，分别加载了`java/lang/Class`和`int` `long` `double` `byte` `char`... 等基本数据类型

``` go
func (self *ClassLoader) loadBasicClasses() {
	jlClassClass := self.LoadClass("java/lang/Class")
	for _, class := range self.classMap {
		if class.jClass == nil {
			class.jClass = jlClassClass.NewObject()
			class.jClass.extra = class
		}
	}
}

// 加载基本数据类型
func (self *ClassLoader) loadPrimitiveClasses() {
	for primitiveType, _ := range primitiveTypes {
		self.loadPrimitiveClass(primitiveType)
	}
}

func (self *ClassLoader) loadPrimitiveClass(className string) {
	class := &Class{
		accessFlags: ACC_PUBLIC,
		name: className,
		loader: self,
		initStarted: true,
	}
	class.jClass = self.classMap["java/lang/Class"].NewObject()
	class.jClass.extra = class
	self.classMap[className] = class
}
```

`jlClassClass := self.LoadClass("java/lang/Class")` 中LoadClass()的实现如下

``` go
func (self *ClassLoader) LoadClass(name string) (class *Class) {
	if class, ok := self.classMap[name]; ok { // 检测缓存中有没有
		// alreay loaded
		return class
	}

	if name[0] == '[' { // 判断是否是数组类
		class = self.loadArrayClass(name)
	} else {
		class = self.loadNonArrayClass(name)
	}

	// 任意一个class加载时都会关联java.lang.Class的一个实例
	// 使之jClass为java.lang.Class的一个实例
	// 使之jClass.extra 为其自身
	if jlClassClass, ok := self.classMap["java/lang/Class"]; ok {
		class.jClass = jlClassClass.NewObject()
		class.jClass.extra = class
	}
	return
}
```

- **对于非数组类，类的加载分为一下三步(构造了类加载器对象后，需要用它去加载一些类，没错，这是包含在classloader的构造函数中的)**
1. 首先找到 class 文件然后把数据读取到内存中
2. 解析 class 文件 生成虚拟机可以使用的类数据 并放入方法区
3. 进行链接

``` go
func (self *ClassLoader) loadNonArrayClass(name string) *Class {
	data, entry := self.readClass(name) // 读取类信息
	class := self.defineClass(data) // 解析类信息
	link(class)
	if self.verboseFlag {
		fmt.Printf("[Loaded %s from %s]\n", name, entry)
	}
	return class
}
```

1. 首先找到 class 文件然后把数据读取到内存中

``` go
func (self *ClassLoader) readClass(name string) ([]byte, classpath.Entry) {
	data, entry, err := self.cp.ReadClass(name)
	if err != nil {
		panic("java.lang.ClassNotFoundException: " + name)
	}
	return data, entry
}
```


2. 解析 class 文件 生成虚拟机可以使用的类数据 并放入方法区

解释一下，方法区，它是运行时数据区的一块逻辑区域，由多个线程共享。方法区主要存放从class文件获取的类信息。此外，类变量也存放在方法区中。当Java虚拟机第一次使用某个类时，它会搜索类路径，找到相应的class文件，然后读取并解析class文件，把相关信息放进方法区。至于方法区到底位于何处，是固定大小还是动态调整，是否参与垃圾回收，以及如何在方法区内存放类数据等，Java虚拟机规范并没有明确规定。

``` go
func (self *ClassLoader) defineClass(data []byte) *Class {
	class := parseClass(data)
	class.loader = self // 绑定加载器
	resolveSuperClass(class)
	resolveInterfaces(class)
	self.classMap[class.name] = class // 注册登记
	return class
}
```

``` go
func parseClass(data []byte) *Class {
	cf, err := classfile.Parse(data)
	if err != nil {
		panic(err)
	}
	return newClass(cf)
}
```

#### 简单来看一下class文件的解析(现在我们在构造JVM, 在构造类加载器的过程中，需要去加载class文件)

``` go
type ClassFile struct {
	magic uint32
	minorVersion uint16
	majorVersion uint16
	constantPool ConstantPool // 常量池
	accessFlags uint16
	thisClass uint16
	superClass uint16
	interClass uint16
	interfaces []uint16
	fields []*MemberInfo
	methods []*MemberInfo
	attributes []AttributeInfo
}
```

解析函数的实现很简单：

``` go
func Parse(classData []byte) (cf *ClassFile, err error) {
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("%v", r)
			}
		}
	} ()

	cr := &ClassReader{classData}
	cf = &ClassFile{}
	cf.read(cr)
	return
}
```

解析后对应字段赋值

``` go
func (self *ClassFile) read(reader *ClassReader) {
	self.readAndCheckMagic(reader)
	self.readAndCheckVersion(reader)
	self.constantPool = readConstantPool(reader)
	self.accessFlags = reader.readUint16()
	self.thisClass = reader.readUint16()
	self.superClass = reader.readUint16()
	self.interfaces = reader.readUint16s()
	self.fields = readMembers(reader, self.constantPool)
	self.methods = readMembers(reader, self.constantPool)
	self.attributes = readAttributes(reader, self.constantPool)
}
```

需要注意的是 `readConstantPool` 、`readMembers` 以及 `readAttributes` 的实现

- 常量池 常量池类似于SymbolTable (现在我们在构造JVM,解析了classpath,正在构造classloader,读取了class文件，现在来解析其中的常量池)

**常量池占据了class文件很大一部分数据，里面存放着各式各样的常量信息，包括数字和字符串常量、类和接口名、字段和方法名，等等**

**可以把常量池中的常量分为两类：字面量（literal）和符号引用（symbolic reference）。字面量包括数字常量和字符串常量，符号引用包括类和接口名、字段和方法信息等。除了字面量，其他常量都是通过索引直接或间接指向CONSTANT_Utf8_info常量**

``` java
public class HelloWorld {
    // 静态常量
    public static final double PI = 3.14;
    // 声明成员常量
    final int y = 10;
    public static void main(String[] args) {
        // 声明局部常量
        final double x = 3.3;
    }
}
```

常量池的实现是通过`type ConstantPool []ConstantInfo` 构造一个ConstantInfo的Slice

``` go
type ConstantInfo interface {
	readInfo(reader *ClassReader)
}
```

cp的构造方法：
``` go
func readConstantPool(reader *ClassReader) ConstantPool {
	cpCount := int(reader.readUint16())
	cp := make([]ConstantInfo, cpCount)

	// 遍历类文件中的常量数据 逐个读取并保存在数组中
	for i := 1;i < cpCount;i++ {
		cp[i] = readConstantInfo(reader, cp)

		switch cp[i].(type) {
		case *ConstantLongInfo, *ConstantDoubleInfo: // Long 和 Double 占用两个位置
			i++
		}
	}
	return cp
}
```

``` go
func readConstantInfo(reader *ClassReader, cp ConstantPool) ConstantInfo {
	tag := reader.readUint8()
	c := newConstantInfo(tag, cp) // 根据不同数据类型的tag 构造不同的常量信息
	c.readInfo(reader)
	return c
}
```

经过这样readConstantPool中的循环，ConstantPool中就保存了class文件中的常量信息。通过Index可以获取对应的常量信息

``` go
func (self ConstantPool) getConstantInfo(index uint16) ConstantInfo {
	if cpInfo := self[index]; cpInfo != nil {
		return cpInfo
	}
	panic(fmt.Errorf("Invalid constant pool index: %v!", index))
}
```


- 字段与方法(现在我们在构造JVM,解析了classpath,正在构造classloader,读取了class文件，现在来解析其中的字段和方法)

``` go
func readMembers(reader *ClassReader, cp ConstantPool) []*MemberInfo  {
	memberCount := reader.readUint16()
	members := make([]*MemberInfo, memberCount)
	for i := range members {
		members[i] = readMember(reader, cp)
	}
	return members
}

func readMember(reader *ClassReader, cp ConstantPool) *MemberInfo {
	return &MemberInfo{
		cp: cp,
		accessFlags: reader.readUint16(),
		nameIndex: reader.readUint16(),
		descriptorIndex: reader.readUint16(),
		attributes: readAttributes(reader, cp),
	}
}
```

- 属性表 同样是一个AttributeInfo的Slice(现在我们在构造JVM,解析了classpath,正在构造classloader,读取了class文件，现在来解析其中的属性表)


``` go
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
	attrInfo := newAttributeInfo(attrName, attrLen, cp) // 根据不同的attrName 构造不同的attrInfo
	attrInfo.readInfo(reader)
	return attrInfo
}
```


``` go
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
```

解析过class文件后,调用newClass() 来构造该类(现在我们在构造JVM,解析了classpath,正在构造classloader,读取了class文件,现在使用这些类信息来构造class对象)

``` go
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
```

对于这段代码，需要注意的有几个：newConstantPool() newFields() newMethods()

- `newConstantPool(class, cf.ConstantPool())` 类对象的运行时常量池指针

``` go
func newConstantPool(class *Class, cfCp classfile.ConstantPool) *ConstantPool {
	cpCount := len(cfCp)
	consts := make([]Constant, cpCount)
	rtCp := &ConstantPool{class, consts}

	for i := 1;i < cpCount;i++ {
		cpInfo := cfCp[i]
		switch cpInfo.(type) {
		case *classfile.ConstantIntegerInfo:
			intInfo := cpInfo.(*classfile.ConstantIntegerInfo)
			consts[i] = intInfo.Value()
        // ...
		case *classfile.ConstantStringInfo:
			stringInfo := cpInfo.(*classfile.ConstantStringInfo)
			consts[i] = stringInfo.String()
		case *classfile.ConstantClassInfo:
			classInfo := cpInfo.(*classfile.ConstantClassInfo)
			consts[i] = newClassRef(rtCp, classInfo) // 当前类的一个引用
		case *classfile.ConstantFieldrefInfo:
			fieldrefInfo := cpInfo.(*classfile.ConstantFieldrefInfo)
			consts[i] = newFieldRef(rtCp, fieldrefInfo)
		case *classfile.ConstantMethodrefInfo:
			methodrefInfo := cpInfo.(*classfile.ConstantMethodrefInfo)
			consts[i] = newMethodRef(rtCp, methodrefInfo)
		case *classfile.ConstantInterfaceMethodrefInfo:
			methodrefInfo := cpInfo.(*classfile.ConstantInterfaceMethodrefInfo)
			consts[i] = newInterfaceMethodRef(rtCp, methodrefInfo)
		default:
			// TODO
		}
	}
	return rtCp
}
```
不难看出，其实这个函数就是解析了classfile中的ConstantPool对象，将其中保存的数据进行映射。值得注意的是引用类型的数据:

``` go
// 根据class文件中存储的类常量创建ClassRef实例
func newClassRef(cp *ConstantPool, classInfo *classfile.ConstantClassInfo) *ClassRef {
	ref := &ClassRef{}
	ref.cp = cp
	ref.className = classInfo.Name()
	return ref
}
```

``` go
func newFieldRef(cp *ConstantPool, refInfo *classfile.ConstantFieldrefInfo) *FieldRef{
	ref := &FieldRef{}
	ref.cp = cp
	ref.copyMemberRefInfo(&refInfo.ConstantMemberrefInfo)
	return ref
}
```

Method 与 Interface_methodRef的实现与field_ref的基本一致，都是拷贝一些必要的信息给引用

``` go
func (self *MemberRef) copyMemberRefInfo(refInfo *classfile.ConstantMemberrefInfo) {
	self.className = refInfo.ClassName()
	self.name, self.descriptor = refInfo.NameAndDescriptor()
}
```

- newFields 类对象的运行时字段

``` go
type ClassMember struct {
	accessFlags uint16 // 访问级别
	name string // 类名
	descriptor string // 描述符
	class *Class // 类指针
}
```

``` go
type Field struct {
	ClassMember
	constValueIndex uint
	slotId uint
}

//  classfile.MemberInfo 转换为 Fileds
func newFileds(class *Class, cfFields []*classfile.MemberInfo) []*Field {
	fields := make([]*Field, len(cfFields))
	for i, cfField := range cfFields {
		fields[i] = &Field{}
		fields[i].class = class
		fields[i].copyMemberInfo(cfField)
		fields[i].copyAttributes(cfField)
	}
	return fields
}
```

- newMethods 类对象的运行时方法

``` go
type Method struct {
	ClassMember
	maxStack uint
	maxLocals uint
	code []byte
	exceptionTable ExceptionTable
	lineNumberTable *classfile.LineNumberTableAttribute
	argSlotCount uint
}

func newMethods(class *Class, cfMethods []*classfile.MemberInfo) []*Method {
	methods := make([]*Method, len(cfMethods))
	for i, cfMethod := range cfMethods {
		methods[i] = newMethod(class, cfMethod)
	}
	return methods
}
```

**构造好类对象后，我们要把这个类注册到类加载器的classMap中,提醒一下，我们现在还没有走出classLoader的构造函数**

``` go
func (self *ClassLoader) loadNonArrayClass(name string) *Class {
	data, entry := self.readClass(name) // 读取类信息
	class := self.defineClass(data) // 解析类信息
	link(class)
	if self.verboseFlag {
		fmt.Printf("[Loaded %s from %s]\n", name, entry)
	}
	return class
}
```

``` go
func (self *ClassLoader) defineClass(data []byte) *Class {
	class := parseClass(data)
	class.loader = self // 绑定加载器
	resolveSuperClass(class)
	resolveInterfaces(class)
	self.classMap[class.name] = class // 注册登记
	return class
}
```

**以上函数是类解析过程的第二步, 看一下resolveSuperClass(class)和resolveInterfaces(class)的实现**

``` go
func resolveSuperClass(class *Class) {
	if class.name != "java/lang/Object" { // Object没有父类
        // 递归向上加载类
		class.superClass = class.loader.LoadClass(class.superClassName)
	}
}

func resolveInterfaces(class *Class) {
	interfaceCount := len(class.interfaceNames)
	if interfaceCount > 0 {
		class.interfaces = make([]*Class, interfaceCount)
		for i, interfaceName := range class.interfaceNames {
            // 逐个加载接口类
			class.interfaces[i] = class.loader.LoadClass(interfaceName)
		}
	}
}
```


3. 进行链接

``` go
func link(class *Class) {
	prepare(class)
}

func prepare(class *Class) {
	calcInstanceFieldSlotIds(class)
	calcStaticFieldSlotIds(class)
	allocAndInitStaticVars(class)
}

```

计算类实例数据以及静态数据占用的槽量
``` go
// calculate how many Instantce vars do we need
func calcInstanceFieldSlotIds(class *Class) {
	slotId := uint(0)
	if class.superClass != nil {
		slotId = class.superClass.instanceSlotCount
	}
	for _, field := range class.fields {
		if !field.IsStatic() {
			field.slotId = slotId
			slotId++
			if field.isLongOrDouble() {
				slotId++
			}
		}
	}
	class.instanceSlotCount = slotId // how many slot we need
}

// calculate how many STATIC vars do we need
func calcStaticFieldSlotIds(class *Class) {
	slotId := uint(0)
	for _, field := range class.fields {
		if field.IsStatic() {
			field.slotId = slotId
			slotId++
			if field.isLongOrDouble() {
				slotId++
			}
		}
	}
	class.staticSlotCount = slotId // how many slot we need
}
```


分配内存并初始化静态变量
``` go
func allocAndInitStaticVars(class *Class) {
	class.staticVars = newSlots(class.staticSlotCount) // allocate mem for static vars
	for _, field := range class.fields {
		if field.IsStatic() && field.IsFinal() {
			initStaticFinalVar(class, field) // init the satic final vars
		}
	}
}

func initStaticFinalVar(class *Class, field *Field) {
	vars := class.staticVars
	cp := class.constantPool
	cpIndex := field.ConstValueIndex()
	slotId := field.SlotId()

	if cpIndex > 0 {
		switch field.Descriptor() {
		case "Z", "B", "C", "S", "I":
			val := cp.GetConstant(cpIndex).(int32)
			vars.SetInt(slotId, val)
        // ...
		}
	}
}
```


- **对于数组类** (我们正在构造JVM对象，构造了classpath后，正在构造classloader，构造classloader时需要去预加载一些class，在这个过程中分为非数组类和数组类)

``` go
func (self *ClassLoader) LoadClass(name string) (class *Class) {

	if name[0] == '[' { // 判断是否是数组类
		class = self.loadArrayClass(name)
	} else {
		class = self.loadNonArrayClass(name)
	}

}
```


``` go
func (self *ClassLoader) loadArrayClass(name string) *Class {
	// int[]{1, 2, 3, 4} 这就是一个数组类 [I
	class := &Class{
		accessFlags: ACC_PUBLIC,
		name: name,
		loader: self,
		initStarted: true, // 数组类不需要初始化
		superClass: self.LoadClass("java/lang/Object"),
		interfaces: []*Class{ // 实现了 以下两个类
			self.LoadClass("java/lang/Cloneable"),
			self.LoadClass("java/io/Serializable"),
		},
	}
	self.classMap[name] = class
	return class
}
```


#### 类的实例化

**为加载好的对象创建一个java/lang/Class的实例作为类信息**

``` go
func (self *ClassLoader) loadBasicClasses() {
	jlClassClass := self.LoadClass("java/lang/Class")
	for _, class := range self.classMap {
		if class.jClass == nil {
			class.jClass = jlClassClass.NewObject()
			class.jClass.extra = class
		}
	}
}
```

``` go
func (self *Class) NewObject() *Object {
	return newObject(self)
}
```

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

Slot是实际存储数据的槽

``` go
type Slot struct {
	num int32 // 数值
	ref *Object // 引用
}

type Slots []Slot

func newSlots(slotCount uint) Slots {
	if slotCount > 0 {
		return make([]Slot, slotCount)
	}
	return nil
}
```

#### 类加载器构造结束

经过漫长的类加载过程，我们终于构造好了类加载器，我们做了什么呢？

1. 我们构造了classpath对象，获取了class文件的数据源
2. 构造了classloader对象，但它里面暂时还没有数据
3. 使用它去加载了java/lang/Class以及一些基本数据类型
4. 类加载器会去从classpath中找到这些类的class文件，读取、解析、链接
5. 预加载过这些类后，为每一个类的jlClass创建一个java/lang/Class实例作为类信息对外暴露

#### jvm的线程

上面，我们构造了类加载器，作为JVM的最后一部分，我们来看一下mainThread主线程

``` go
func newJVM(cmd *Cmd) *JVM {
	cp := classpath.Parse(cmd.XjreOption, cmd.cpOption)
	classLoader := heap.NewClassLoader(cp, cmd.verboseClassFlag)
	return &JVM{
		cmd: cmd,
		classLoader: classLoader,
		mainThread: rtda.NewThread(),
	}
}
```

由于我们只是实现一个简单的玩具，并不考虑实现多线程。

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


type Thread struct {
	pc int
	stack *Stack
}

func NewThread() *Thread {
	return &Thread{
		stack: newStack(1024), // 最多存放1024个栈帧
	}
}
```

jvm中一个抽象的栈

``` go
type Stack struct {
	maxSize uint
	size uint
	_top *Frame
}

func newStack(maxSize uint) *Stack {
	return &Stack{
		maxSize: maxSize,
	}
}
```

### 初始化

``` go
func (self *JVM) start() {
	self.initVM()
	self.execMain()
}

func (self *JVM) initVM() {
	vmClass := self.classLoader.LoadClass("sun/misc/VM")
	base.InitClass(self.mainThread, vmClass)
	interpret(self.mainThread, self.cmd.verboseInstFlag)
}
```

**类的加载已经讲过了，现在去初始化类**

``` go
// 初始化类
func InitClass(thread *rtda.Thread, class *heap.Class) {
    // 标志已经初始化
	class.StartInit()
    // 构造函数压栈
	scheduleClinit(thread, class)
    // 递归构造父类
	initSuperClass(thread, class)
}

func scheduleClinit(thread *rtda.Thread, class *heap.Class) {
	clinit := class.GetClinitMethod()
	if clinit != nil {
		// exec <clinit>
		newFrame := thread.NewFrame(clinit)
		thread.PushFrame(newFrame)
	}
}

func initSuperClass(thread *rtda.Thread, class *heap.Class) {
	if !class.IsInterface() { // not a interface
		superClass := class.SuperClass()
		if superClass != nil && !superClass.InitStarted() {
			InitClass(thread, superClass)
		}
	}
}

```

来说一下 GetClinitMethod 和 Frame的几个操作

``` go
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
```

从当前类及其父类中寻找名为"<clinit>"描述符为"()V"的方法

- Frame的操作

``` go
func (self *Thread) NewFrame(method *heap.Method) *Frame {
	return newFrame(self, method)
}

type Frame struct {
	lower *Frame
	localVars LocalVars
	operandStack *OperandStack
	thread *Thread
	method *heap.Method
	nextPC int // the next instruction after the call
}

func newFrame(thread *Thread, method *heap.Method) *Frame {
	return &Frame {
		thread: thread,
		method: method,
		localVars: newLocalVars(method.MaxLocals()),
		operandStack: newOperandStack(method.MaxStack()),
	}
}
```

这里的Frame是一个链表的节点，我们使用链表来模拟一个Stack. 每个Frame就是一个函数栈帧，函数栈帧中又包含了局部变量和操作数栈。

PushFrame相当于将当前函数栈帧压入线程栈顶，这样执行时就会操作该栈内的数据

`interpret`函数在下一节进行解说

### 执行字节码

在class文件文件中会有Code AttributeInfo 其中包含了jvm字节码，通过解析这些字节码，翻译成对应的指令，再逐步执行这些我们已经实现好了的指令，就实现了一个有效的虚拟机

``` go
func (self *JVM) start() {
	self.initVM()
	self.execMain()
}

func (self *JVM) execMain() {
	className := strings.Replace(self.cmd.class, ".", "/", -1)
	mainClass := self.classLoader.LoadClass(className)
	mainMethod := mainClass.GetMainMethod()
	if mainMethod == nil {
		fmt.Printf("Main method not found in class %s\n", self.cmd.class)
		return
	}

	argsArr := self.createArgsArray()
	frame := self.mainThread.NewFrame(mainMethod)
	frame.LocalVars().SetRef(0, argsArr)
	self.mainThread.PushFrame(frame)
	interpret(self.mainThread, self.cmd.verboseInstFlag)
}
```

执行Main方法时，先加载当前主类，然后找到main方法，为main方法新建一个函数栈帧，设置函数本地变量表的第0位为cli传入的参数

``` go
func interpret(thread *rtda.Thread, logInst bool) {
	defer catchErr(thread)
	loop(thread, logInst)
}

func catchErr(thread *rtda.Thread) {
	if r := recover(); r != nil {
		logFrames(thread)
		panic(r)
	}
}

func loop(thread *rtda.Thread, logInst bool) {
	reader := &base.BytecodeReader{}
	for {
		frame := thread.CurrentFrame() // 当前函数栈帧
		pc := frame.NextPC()
		thread.SetPC(pc)

		// decode
		reader.Reset(frame.Method().Code(), pc)
		opcode := reader.ReadUint8()
		inst := instructions.NewInstruction(opcode)
		inst.FetchOperands(reader)
		frame.SetNextPC(reader.PC())

		if logInst {
			logInstruction(frame, inst)
		}

		// execute
		inst.Execute(frame)
		if thread.IsStackEmpty() {
			break
		}
	}
}
```


当loop循环开始后，位于mainThread栈顶的是mainFrame，向后移动PC后，构造Instruction对象，开始解码当前函数的的字节码，然后执行该指令的Execute()

``` go

func NewInstruction(opcode byte) base.Instruction {
	switch opcode {
	case 0x00:
		return nop
    // ... 很多指令 200条左右
	case 0xbb:
		return &NEW{}
	case 0xbc:
		return &NEW_ARRAY{}
	case 0xbd:
		return &ANEW_ARRAY{}
	case 0xbe:
		return arraylength
	case 0xbf:
		return athrow
	case 0xc0:
		return &CHECK_CAST{}
	case 0xc1:
		return &INSTANCE_OF{}
	case 0xc4:
		return &WIDE{}
	case 0xc5:
		return &MULTI_ANEW_ARRAY{}
	case 0xc6:
		return &IFNULL{}
	case 0xc7:
		return &IFNONNULL{}
	case 0xc8:
		return &GOTO_W{}
	case 0xfe:
		return invoke_native
	default:
		panic(fmt.Errorf("Unsupported opcode: 0x%x!", opcode))
	}
}
```
