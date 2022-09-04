package heap

import (
	"jvm/classpath"
)

type ClassLoader struct {
	cp *classpath.Classpath
	classMap map[string]*Class
}

// 1 首先找到 class 文件然后把数据读取到内存中
// 2 解析 class 文件 生成虚拟机可以使用的类数据 并放入方法区
// 3 进行链接

func NewClassLoader(cp *classpath.Classpath) *ClassLoader {
	return &ClassLoader{
		cp: cp,
		classMap: make(map[string]*Class),
	}
}
