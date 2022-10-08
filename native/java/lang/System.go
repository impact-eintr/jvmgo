package lang

import (
	"jvm/native"
	"jvm/rtda"
	"jvm/rtda/heap"
	"runtime"
	"time"
)

const jlSystem = "java/lang/System"

func init() {
	native.Register(jlSystem, "arraycopy",
		"(Ljava/lang/Object;ILjava/lang/Object;II)V", arraycopy)
	native.Register(jlSystem, "initProperties",
		"(Ljava/util/Properties;)Ljava/util/Properties;", initProperties)
  native.Register(jlSystem, "setIn0", "(Ljava/io/InputStream;)V", setIn0)
  native.Register(jlSystem, "setOut0", "(Ljava/io/PrintStream;)V", setOut0)
  native.Register(jlSystem, "setErr0", "(Ljava/io/PrintStream;)V", setErr0)
  native.Register(jlSystem, "currentTimeMillis", "()J", currentTimeMillis)
}

func arraycopy(frame *rtda.Frame) {
	vars := frame.LocalVars()
	src := vars.GetRef(0)
	srcPos := vars.GetInt(1)
	dst := vars.GetRef(2)
	dstPos := vars.GetInt(3)
	length := vars.GetInt(4)

	if src == nil || dst == nil {
		panic("java.lang.NullPointerException")
	}
	if !checkArrayCopy(src, dst) {
		panic("java.lang.ArrayStoreException")
	}
	if srcPos < 0 || dstPos < 0 || length < 0 ||
		srcPos+length > src.ArrayLength() ||
		dstPos+length > dst.ArrayLength() {
		panic("java.lang.IndexOutOfBoundsException")
	}
	heap.ArrayCopy(src, dst, srcPos, dstPos, length)
}

func checkArrayCopy(src, dst *heap.Object) bool {
	srcClass := src.Class()
	dstClass := dst.Class()

	if !srcClass.IsArray() || !dstClass.IsArray() {
		return false
	}
	// 如果是基本类型数组 两种数组应该同类型
	if srcClass.ComponentClass().IsPrimitive() ||
		dstClass.ComponentClass().IsPrimitive() {
		return srcClass == dstClass
	}
	return true
}

// private static native Properties initProperties(Properties props);
// (Ljava/util/Properties;)Ljava/util/Properties;
func initProperties(frame *rtda.Frame) {
	vars := frame.LocalVars()
	props := vars.GetRef(0)

	stack := frame.OperandStack()
	stack.PushRef(props)

	//setPropMethod := props.Class().GetInstanceMethod("setProperty",
	//	"(Ljava/lang/String;Ljava/lang/String;)Ljava/lang/Object;")
	//thread := frame.Thread()
	for k, v := range _sysProps() {
		jKey := heap.JString(frame.Method().Class().Loader(), k)
		jVal := heap.JString(frame.Method().Class().Loader(), v)
		ops := rtda.NewOperandStack(3)
		ops.PushRef(props)
		ops.PushRef(jKey)
		ops.PushRef(jVal)
		// TODO
		println("可爱捏")
		//base.InvokeMethod(frame, setPropMethod) // FIXME
	}
}

func _sysProps() map[string]string {
	return map[string]string{
		"java.version":         "1.8.0",
		"java.vendor":          "jvm.go",
		"java.vendor.url":      "https://github.com/impact-eintr/jvm",
		"java.home":            "todo",
		"java.class.version":   "52.0",
		"java.class.path":      "todo",
		"java.awt.graphicsenv": "sun.awt.CGraphicsEnvironment",
		"os.name":              runtime.GOOS,   // todo
		"os.arch":              runtime.GOARCH, // todo
		"os.version":           "",             // todo
		"file.separator":       "/",            // todo os.PathSeparator
		"path.separator":       ":",            // todo os.PathListSeparator
		"line.separator":       "\n",           // todo
		"user.name":            "",             // todo
		"user.home":            "",             // todo
		"user.dir":             ".",            // todo
		"user.country":         "CN",           // todo
		"file.encoding":        "UTF-8",
		"sun.stdout.encoding":  "UTF-8",
		"sun.stderr.encoding":  "UTF-8",
	}
}


// private static native void setIn0(InputStream in);
// (Ljava/io/InputStream;)V
func setIn0(frame *rtda.Frame) {
	vars := frame.LocalVars()
	in := vars.GetRef(0)

	sysClass := frame.Method().Class()
	sysClass.SetRefVar("in", "Ljava/io/InputStream;", in)
}

// private static native void setIn0(PrintStream out);
// (Ljava/io/PrintStream;)V
func setOut0(frame *rtda.Frame) {
	vars := frame.LocalVars() // []Slot => []Object
	out := vars.GetRef(0) // an Object

	sysClass := frame.Method().Class() // java/lang/System
	sysClass.SetRefVar("out", "Ljava/io/PrintStream;", out)
}

// private static native void setIn0(PrintStream err);
// (Ljava/io/PrintStream;)V
func setErr0(frame *rtda.Frame) {
	vars := frame.LocalVars()
	err := vars.GetRef(0)

	sysClass := frame.Method().Class()
	sysClass.SetRefVar("err", "Ljava/io/PrintStream;", err)
}

// public static native long currentTimeMillis();
// ()J
func currentTimeMillis(frame *rtda.Frame) {
	millis := time.Now().UnixNano() / int64(time.Millisecond)
	stack := frame.OperandStack()
	stack.PushLong(millis)
}
