package onlineconf

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/jbarham/go-cdb"
	"io"
	"log"
	"strconv"
	"sync"
)

const configDir = "/usr/local/etc/onlineconf"

var watcher *fsnotify.Watcher

func init() {
	modules.byName = make(map[string]*Module)
	modules.byFile = make(map[string]*Module)

	var err error
	watcher, err = fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}

	err = watcher.Add(configDir)
	if err != nil {
		panic(err)
	}

	go func() {
		for {
			select {
			case ev := <-watcher.Events:
				//log.Println("fsnotify event:", ev)

				if ev.Op&fsnotify.Create == fsnotify.Create {
					modules.Lock()
					module, ok := modules.byFile[ev.Name]
					modules.Unlock()

					if ok {
						module.reopen()
					}
				}

			case err := <-watcher.Errors:
				log.Printf("Watch %v error: %v\n", configDir, err)
			}
		}
	}()
}

func SetOutput(w io.Writer) {
	log.SetOutput(w)
}

type Module struct {
	sync.RWMutex
	name string
	file string
	cdb  *cdb.Cdb
}

func newModule(name string) *Module {
	file := fmt.Sprintf("%s/%s.cdb", configDir, name)
	cdb, err := cdb.Open(file)
	if err != nil {
		panic(err)
	}
	return &Module{name: name, file: file, cdb: cdb}
}

func (m *Module) reopen() {
	log.Printf("Reopen %s\n", m.file)
	m.Lock()
	defer m.Unlock()
	cdb, err := cdb.Open(m.file)
	if err != nil {
		log.Printf("Reopen file %v error: %v\n", m.file, err)
	} else {
		m.cdb.Close()
		m.cdb = cdb
	}
}

func (m *Module) get(path string) (byte, []byte) {
	m.RLock()
	defer m.RUnlock()
	data, err := m.cdb.Data([]byte(path))
	if err != nil || len(data) == 0 {
		if err != io.EOF {
			log.Printf("Get %v:%v error: %v", m.file, path, err)
		}
		return 0, data
	}
	return data[0], data[1:]
}

func (m *Module) GetStringIfExists(path string) (string, bool) {
	format, data := m.get(path)
	switch format {
	case 0:
		return "", false
	case 's':
		return string(data), true
	default:
		log.Printf("%s:%s: format is not string\n", m.name, path)
		return "", false
	}
}

func (m *Module) GetIntIfExists(path string) (int, bool) {
	str, ok := m.GetStringIfExists(path)
	if !ok {
		return 0, false
	}

	i, err := strconv.Atoi(str)
	if err != nil {
		log.Printf("%s:%s: value is not an integer: %s\n", m.name, path, str)
		return 0, false
	}

	return i, true
}

func (m *Module) GetString(path string, d ...string) string {
	if val, ok := m.GetStringIfExists(path); ok {
		return val
	} else if len(d) > 0 {
		return d[0]
	} else {
		panic(fmt.Sprintf("%s:%s key not exists and default not found", m.name, path))
	}
}

func (m *Module) GetInt(path string, d ...int) int {
	if val, ok := m.GetIntIfExists(path); ok {
		return val
	} else if len(d) > 0 {
		return d[0]
	} else {
		panic(fmt.Sprintf("%s:%s key not exists and default not found", m.name, path))
	}
}

var modules struct {
	sync.Mutex
	byName map[string]*Module
	byFile map[string]*Module
}

func GetModule(name string) *Module {
	modules.Lock()
	defer modules.Unlock()

	if module, ok := modules.byName[name]; ok {
		return module
	}

	module := newModule(name)

	modules.byName[module.name] = module
	modules.byFile[module.file] = module

	return module
}

var tree struct {
	sync.Mutex
	module *Module
}

func getTree() *Module {
	if tree.module != nil {
		return tree.module
	}

	tree.Lock()
	defer tree.Unlock()

	if tree.module != nil {
		return tree.module
	}

	tree.module = GetModule("TREE")
	return tree.module
}

func GetStringIfExists(path string) (string, bool) {
	return getTree().GetStringIfExists(path)
}

func GetIntIfExists(path string) (int, bool) {
	return getTree().GetIntIfExists(path)
}

func GetString(path string, d ...string) string {
	return getTree().GetString(path, d...)
}

func GetInt(path string, d ...int) int {
	return getTree().GetInt(path, d...)
}
