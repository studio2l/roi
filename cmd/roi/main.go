package main

import (
	"fmt"
	"os"
)

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "roi [command] [command-args]")
		fmt.Fprintln(os.Stderr, "commands: server, shot")
		os.Exit(1)
	}
	subCmd := args[0]
	subArgs := args[1:]
	switch subCmd {
	case "server":
		serverMain(subArgs)
	case "shot":
		shotMain(subArgs)
	}
}
