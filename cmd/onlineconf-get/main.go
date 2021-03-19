package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/onlineconf/onlineconf-go"
)

func main() {

	isInteractive := flag.Bool("interactive", false, "Run get onlineconf in interactive mode")
	ocModuleDir := flag.String("module-dir", "/usr/local/etc/onlineconf", "Onlineconf module dir")
	ocModuleName := flag.String("module", "TREE", "Onlineconf module")
	asBool := flag.Bool("bool", false, "Interpret value as boolean and exit with code 0 on true and 1 on false. Only non-interactive mode")

	flag.Parse()

	if *asBool && *isInteractive {
		fmt.Fprintln(os.Stderr, "-bool option i not available in interactive mode")
		os.Exit(2)
	}

	onlineconf.Initialize(*ocModuleDir)
	onlineconf.SetOutput(os.Stderr)
	module := onlineconf.GetModule(*ocModuleName)

	if !*isInteractive {
		if flag.CommandLine.NArg() != 1 {
			fmt.Println("path argiment is required for non interactive mode")
			os.Exit(2)
		}
		onlineconfPath := flag.CommandLine.Arg(0)

		if *asBool {
			value := module.GetBool(onlineconfPath, false)
			if value {
				os.Exit(0)
			}
			os.Exit(1)
		}

		readOCPath(module, onlineconfPath)

		return
	}

	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter onlineconf path: ")
		onlineconfPath, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}
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

	fmt.Fprintln(os.Stderr, "No such key")
}
