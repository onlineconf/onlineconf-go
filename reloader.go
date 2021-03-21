package onlineconf

import (
	"context"
	"fmt"
	"path"
	"sync"

	"github.com/fsnotify/fsnotify"
	"golang.org/x/exp/mmap"
)

// DefaultModulesDir defines default directory for modules
const DefaultModulesDir = "/usr/local/etc/onlineconf"

// ReloaderOptions specify loader options
// You can specify either FilePath or Name + Dir.
// If you sprcified only Name, DefaultModulesDir Dir will be used
type ReloaderOptions struct {
	Name     string
	Dir      string // default in `DefaultModulesDir`
	FilePath string
}

// ModuleReloader watchers for module updates and reloads it
type ModuleReloader struct {
	module  *Module
	modMu   *sync.RWMutex // module mutex
	ops     *ReloaderOptions
	cdbFile *mmap.ReaderAt
}

// Reloader returns reloader for specified module
func Reloader(moduleName string) (*ModuleReloader, error) {
	mr, err := NewModuleReloader(&ReloaderOptions{Name: moduleName})
	if err != nil {
		return nil, err
	}
	return mr, nil
}

// MustReloader returns reloader for specified module. Panics on error.
func MustReloader(moduleName string) *ModuleReloader {
	mr, err := NewModuleReloader(&ReloaderOptions{Name: moduleName})
	if err != nil {
		panic(err)
	}
	return mr
}

// GetModule returns current copy of Module by name in default onlineconf module directory.
// This function is STRONGLY NOT RECOMENDED FOR USE. It is very unefficient.
// This module will never be updated.
// This function parses all the parameters in onlineconf and copies it to memory.
// Thats why this operations is expencive. You should prefer to use ModuleReloader if its possible.
func GetModule(moduleName string) (*Module, error) {
	mr, err := Reloader(moduleName)
	if err != nil {
		return nil, err
	}

	return mr.Module(), nil
}

// MustModule returns Module by name in default onlineconf module directory.
// Panics on error.
// This module will never be updated.
// This function parses all the parameters in onlineconf and copies it to memory.
// Thats why this operations is expencive. You should prefer to use ModuleReloader if its possible.
func MustModule(moduleName string) *Module {
	m, err := GetModule(moduleName)
	if err != nil {
		panic(err)
	}
	return m
}

// Tree returns current copy of Module TREE in default onlineconf module directory.
// This module will never be updated.
// This function parses all the parameters in onlineconf and copies it to memory.
// Thats why this operations is expencive. You should prefer to use ModuleReloader if its possible.
func Tree() (*Module, error) {
	return GetModule("TREE")
}

// MustTree returns Module TREE in default onlineconf module directory
// Panics on error.
// This module will never be updated.
// This function parses all the parameters in onlineconf and copies it to memory.
// Thats why this operations is expencive. You should prefer to use ModuleReloader if its possible.
func MustTree() *Module {
	m, err := Tree()
	if err != nil {
		panic(err)
	}
	return m
}

// NewModuleReloader creates new module reloader
func NewModuleReloader(ops *ReloaderOptions) (*ModuleReloader, error) {
	if ops.FilePath == "" {
		if ops.Dir == "" {
			ops.Dir = DefaultModulesDir
		}
		fileName := fmt.Sprintf("%s.cdb", ops.Name)
		filePath := path.Join(ops.Dir, fileName)
		ops.FilePath = filePath
	}

	mr := ModuleReloader{
		ops:   ops,
		modMu: &sync.RWMutex{},
	}
	err := mr.Reload()
	if err != nil {
		return nil, err
	}

	return &mr, nil
}

// Module returns the last successfully updated version of module
func (mr *ModuleReloader) Module() *Module {
	mr.modMu.RLock()
	mod := mr.module
	defer mr.modMu.RUnlock()
	return mod
}

func (mr *ModuleReloader) RunWatcher(ctx context.Context) error {
	var watcher *fsnotify.Watcher

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("can't init inotify watcher: %w", err)
	}

	err = watcher.Add(mr.ops.FilePath)
	if err != nil {
		return fmt.Errorf("can't add inotify watcher for module %s: %w", mr.ops.Name, err)
	}

	var watcherLoopErr error

WATCHERLOOP:
	for {
		select {
		case ev := <-watcher.Events:
			if ev.Name == mr.ops.FilePath && ev.Op&fsnotify.Chmod == fsnotify.Chmod {
				watcherLoopErr = mr.Reload()
				break WATCHERLOOP
			}
		case err := <-watcher.Errors:
			if err != nil {
				watcherLoopErr = fmt.Errorf("onlineconf reloader (%s) fsnotify watcher failed: %w", mr.ops.Name, err)
				break WATCHERLOOP
			}
		case <-ctx.Done():
			break WATCHERLOOP
		}
	}

	watcher.Close()

	return watcherLoopErr
}

func (mr *ModuleReloader) Reload() error {

	cdbFile, err := mmap.Open(mr.ops.FilePath)
	if err != nil {
		return err
	}

	module, err := NewModule(cdbFile)
	if err != nil {
		return err
	}

	if err != nil {
		return fmt.Errorf("can't reload module: %w", err)
	}

	oldCDB := mr.cdbFile
	mr.modMu.Lock()
	mr.cdbFile = cdbFile
	mr.module = module
	mr.modMu.Unlock()

	if oldCDB != nil {
		// todo we can't just close it here
		// there could be any number of opened modules that
		// the simplest way is to rely on mmap finalyzer
		// oldCDB.Close()
	}

	return nil
}

type ctxConfigModuleReloaderKey struct{}

// WithContext returns a new Context that carries value module reloader
func (mr *ModuleReloader) WithContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, ctxConfigModuleReloaderKey{}, mr)
}

// ModuleReloaderFromContext retrieves a config module from context.
func ModuleReloaderFromContext(ctx context.Context) (*ModuleReloader, bool) {
	mr, ok := ctx.Value(ctxConfigModuleReloaderKey{}).(*ModuleReloader)
	return mr, ok
}
