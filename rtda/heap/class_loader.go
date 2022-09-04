package heap

import (
	"jvm/classpath"
)

type ClassLoader struct {
	cp *classpath.Classpath
	classMap map[string]*Class
}

func NewClassLoader(cp *classpath.Classpath) *ClassLoader {
	return &ClassLoader{
		cp: cp,
		classMap: make(map[string]*Class),
	}
}
