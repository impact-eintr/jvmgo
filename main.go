package main

import (
	"fmt"
)



func main() {
	cmd := parseCmd()

	if cmd.versionFlag {
		fmt.Println("version: v0.0.1")
	} else if cmd.helpFlag || cmd.class == "" {
		printUsage()
	} else {
		startJVM(cmd)
	}
}

func startJVM(cmd *Cmd) {
	cp := classpath.Prse(cmd.XjreOption, cmd.cpOption)
	fmt.Printf("classpath: %s class:%s args:%v\n",
		cmd.cpOption, cmd.class, cmd.args)
}
