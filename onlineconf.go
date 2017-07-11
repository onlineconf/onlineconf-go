package onlineconf

import (
	"bufio"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
)

type row struct {
	i int
	c bool
	s string
}

var cfg map[string]*row = make(map[string]*row)

//var file string = "/usr/local/etc/onlineconf/TREE.conf"
var config_file string = fmt.Sprintf("/usr/local/etc/onlineconf/%s.conf", path.Base(os.Args[0]))

func init() {
	read()

	go func() {
		watcher, err := fsnotify.NewWatcher()

		if err != nil {
			panic(err)
		}
		defer watcher.Close()

		err = watcher.Add(config_file)

		if err != nil {
			panic(err)
		}

		for {
			select {
			case ev := <-watcher.Events:
				//log.Println("fsnotify event:", ev)

				if ev.Op&fsnotify.Remove == fsnotify.Remove {

					err = watcher.Add(config_file)
					if err != nil {
						panic(err)
					}
					read()
				}

				if (ev.Op&fsnotify.Create == fsnotify.Create) || (ev.Op&fsnotify.Write == fsnotify.Write) {
					read()
				}

			case err := <-watcher.Errors:
				log.Printf("Watch %v error: %v\n", config_file, err)
			}
		}
	}()
}

func read() {
	log.Printf("Start read file %v\n", config_file)

	f, err := os.Open(config_file)

	if err != nil {
		log.Printf("Open file %v error %v\n", config_file, err)
		return
	}

	defer f.Close()

	buff := make([]string, 2)
	scan := bufio.NewScanner(f)

	for scan.Scan() {
		if buff = strings.SplitN(scan.Text(), " ", 2); len(buff) == 2 {
			if val, ok := cfg[buff[0]]; ok {
				if val.s != buff[1] {
					val.c = false
					val.s = buff[1]
				}
			} else {
				cfg[buff[0]] = &row{
					c: false,
					s: buff[1],
				}
			}
		}
	}

	log.Printf("Finish read file %v\n", config_file)
}

func GetAsInt(k string, d ...int) int {
	if val, ok := cfg[k]; ok {
		if val.c == true {
			return val.i
		}

		i, err := strconv.Atoi(val.s)

		if err != nil {
			log.Printf("strconv.Atoi key %s error %v\n", k, config_file)
		} else {
			val.i = i
			val.c = true

			return i
		}
	}

	if len(d) >= 1 {
		return d[0]
	}

	panic(fmt.Sprintf("%v key not exist and default not found", k))
}

func GetAsString(k string, d ...string) string {
	if val, ok := cfg[k]; ok {
		return val.s
	}

	if len(d) >= 1 {
		return d[0]
	}

	panic(fmt.Sprintf("%v key not exist and default not found", k))
}

func IfExistAsString(k string) (string, bool) {
	if val, ok := cfg[k]; ok {
		return val.s, ok
	}

	return "", false
}
