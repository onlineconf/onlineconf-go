package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/onlineconf/onlineconf-go"
)

func main() {

	isInteractive := flag.Bool("interactive", false, "run get onlineconf in interactive mode")
	ocModuleName := flag.String("module", "TREE", "onlineconf module")

	flag.Parse()

	module := onlineconf.GetModule(*ocModuleName)

	if !*isInteractive {
		if flag.CommandLine.NArg() != 1 {
			fmt.Println("path argiment is required for non interactive mode")
			os.Exit(1)
		}
		onlineconfPath := flag.CommandLine.Arg(0)
		readOCPath(module, onlineconfPath)

		return
	}

	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter onlineconf path: ")
		onlineconfPath, _ := reader.ReadString('\n')
		onlineconfPath = strings.Trim(onlineconfPath, "\n")
		readOCPath(module, onlineconfPath)
	}

}

func readOCPath(module *onlineconf.Module, path string) {
	value, ok := module.GetStringIfExists(path)
	if ok {
		fmt.Println(value)
		return
	}

	fmt.Println("No such key")
}
