package main

import "fmt"

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
