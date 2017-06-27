package onlineconf

import(
    "os"
    "fmt"
    "log"
    "bufio"
    "strings"
    "strconv"
    "github.com/howeyc/fsnotify"
)

type row struct {
    i int
    c bool
    s string
}

var cfg map[string]*row = make(map[string]*row)
var file string = "/usr/local/etc/onlineconf/TREE.conf"

func init() {
    read()

    go func() {
        watcher, err := fsnotify.NewWatcher()

        defer watcher.Close()

        if err != nil {
            panic(err)
        }

        err = watcher.Watch(file)

        if err != nil {
            panic(err)
        }

        for {
            select {
                case ev := <-watcher.Event:
                    err := watcher.RemoveWatch(file)

                    if err != nil {
                        panic(err)
                    }

                    if ev.IsCreate() || ev.IsModify() {
                        read()
                    }

                    err = watcher.Watch(file)

                    if err != nil {
                        panic(err)
                    }
                case err := <-watcher.Error:
                    log.Printf("Watch %v error: %v\n", file, err)
            }
        }
    }()
}

func read() {
    log.Printf("Start read file %v\n", file)

    f, err := os.Open(file)

    if err != nil {
        log.Printf("Open file %v error %v\n", file, err)
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

    log.Printf("Finish read file %v\n", file)
}

func GetAsInt(k string, d ...int) int {
    if val, ok := cfg[k]; ok {
        if val.c == true {
            return val.i
        }

        i, err := strconv.Atoi(val.s)

        if err != nil {
            log.Printf("strconv.Atoi key %s error %v\n", k, file)
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
