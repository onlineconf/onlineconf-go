package onlineconf

import (
	"context"
	"fmt"
	"io"
	"sync"

	"errors"

	"github.com/colinmarc/cdb"
)

// ErrInvalidCDB means that cdb is invalid
var ErrInvalidCDB = errors.New("cdb is inconsistent")

// Module is a structure that associated with onlineconf module file.
type Module struct {
	CDB *cdb.CDB

	stringParamsMu sync.RWMutex
	stringParams   map[string]string
	intParamsMu    sync.RWMutex
	intParams      map[string]int
	boolParamsMu   sync.RWMutex
	boolParams     map[string]bool

	notExistingParamsMu sync.RWMutex
	notExistingParams   map[string]struct{}
}

// NewModule
func NewModule(reader io.ReaderAt) (*Module, error) {

	cdbReader, err := cdb.New(reader, nil)
	if err != nil {
		return nil, fmt.Errorf("can't get cdb reader for module: %w", err)
	}

	module := &Module{
		CDB: cdbReader,

		// todo use prev module to bulk allocate memory for params
		boolParams:   make(map[string]bool),
		intParams:    make(map[string]int),
		stringParams: make(map[string]string),

		notExistingParamsMu: sync.RWMutex{},
		notExistingParams:   make(map[string]struct{}),
	}

	return module, nil
}

// NewPredeclaredModule
func NewPredeclaredModule(reader io.ReaderAt, paramsDescriptsion []ConfigParam) (*Module, error) {
	module, err := NewModule(reader)
	if err != nil {
		return nil, err
	}

	// todo fill params

	return module, nil
}

// getCachedNotExist returns true is path not in module.
// false returned in case we don't know either param exists or not exists
func (m *Module) getCachedNotExist(path *ParamPath) bool {
	m.notExistingParamsMu.RLock()
	defer m.notExistingParamsMu.RUnlock()
	_, ok := m.notExistingParams[path.path]
	return ok
}

func (m *Module) setCachedNotExisting(pathParam *ParamPath) {

	if ok := m.getCachedNotExist(pathParam); ok {
		return
	}

	m.notExistingParamsMu.Lock()
	m.notExistingParams[pathParam.path] = struct{}{}
	defer m.notExistingParamsMu.Unlock()

	return
}

func (m *Module) getStringCached(path *ParamPath) (string, bool) {
	m.stringParamsMu.RLock()
	defer m.stringParamsMu.RUnlock()

	param, ok := m.stringParams[path.path]
	return param, ok
}

func (m *Module) setStringCached(path *ParamPath, value string) {
	if _, ok := m.getStringCached(path); ok {
		return
	}

	m.stringParamsMu.Lock()
	defer m.stringParamsMu.Unlock()

	m.stringParams[path.path] = value
	return
}

func (m *Module) getIntCached(path *ParamPath) (int, bool) {
	m.intParamsMu.RLock()
	defer m.intParamsMu.RUnlock()

	param, ok := m.intParams[path.path]
	return param, ok
}

func (m *Module) setIntCached(path *ParamPath, value int) {
	if _, ok := m.getIntCached(path); ok {
		return
	}

	m.intParamsMu.Lock()
	defer m.intParamsMu.Unlock()

	m.intParams[path.path] = value
	return
}

func (m *Module) getBoolCached(path *ParamPath) (bool, bool) {
	m.boolParamsMu.RLock()
	defer m.boolParamsMu.RUnlock()

	param, ok := m.boolParams[path.path]
	return param, ok
}

func (m *Module) setBoolCached(path *ParamPath, value bool) {
	if _, ok := m.getBoolCached(path); ok {
		return
	}

	m.boolParamsMu.Lock()
	defer m.boolParamsMu.Unlock()

	m.boolParams[path.path] = value
	return
}

type ctxConfigModuleKey struct{}

// WithContext returns a new Context that carries value module
func (m *Module) WithContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, ctxConfigModuleKey{}, m)
}

// ModuleFromContext retrieves a config module from context.
func ModuleFromContext(ctx context.Context) (*Module, bool) {
	m, ok := ctx.Value(ctxConfigModuleKey{}).(*Module)
	return m, ok
}
