package main

import (
	"fmt"
	"os"

	"hago-plat/pcom/quintus/cli/quintus/cmd"
)

func main() {
	command := cmd.NewQuintusCommand()
	err := command.Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
