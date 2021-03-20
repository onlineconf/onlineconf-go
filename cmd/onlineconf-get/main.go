package main

import (
	"bufio"
	"context"
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
		panic("-bool option i not available in interactive mode")
	}

	reloader, err := onlineconf.NewModuleReloader(&onlineconf.ReloaderOptions{
		Dir:  *ocModuleDir,
		Name: *ocModuleName,
	})
	if err != nil {
		panic(err)
	}

	module := reloader.Module()

	if !*isInteractive {
		if flag.CommandLine.NArg() != 1 {
			panic("path argiment is required for non interactive mode")
		}
		paramPath := flag.CommandLine.Arg(0)

		if *asBool {
			value, err := module.Bool(onlineconf.MustConfigParamBool(paramPath, false))
			if err != nil {
				panic(err)
			}

			if value {
				os.Exit(0)
			}
			os.Exit(1)
		}

		readOCPath(module, paramPath)

		return
	}

	go reloader.RunWatcher(context.Background())

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
		module = reloader.Module()

		readOCPath(module, onlineconfPath)
	}

}

func readOCPath(module *onlineconf.Module, path string) {
	paramPath, err := onlineconf.NewConfigParamString(path, "")
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	value, err := module.String(paramPath)
	if onlineconf.IsErrKeyNotFound(err) {
		fmt.Println("No such key: ", path)
		return
	}
	if err != nil {
		panic(err)
	}

	fmt.Println(value)
	return
}