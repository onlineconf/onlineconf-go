package onlineconf

import (
	"context"
	"fmt"
	"io"

	"errors"

	"github.com/alldroll/cdb"
)

// ErrInvalidCDB means that cdb is invalid
var ErrInvalidCDB = errors.New("cdb is inconsistent")

// Module is a structure that associated with onlineconf module file.
type Module struct {
	CDB cdb.Reader

	stringParams map[string]string
	intParams    map[string]int
	boolParams   map[string]bool

	notExistingParams map[string]struct{}
}

// NewModule
func NewModule(reader io.ReaderAt) (*Module, error) {

	cdbReader, err := cdb.New().GetReader(reader)
	if err != nil {
		return nil, fmt.Errorf("Cant cant cdb reader for module: %w", err)
	}

	module := &Module{
		CDB: cdbReader,

		// todo use prev module to bulk allocate memory for params
		boolParams:   make(map[string]bool),
		intParams:    make(map[string]int),
		stringParams: make(map[string]string),

		notExistingParams: make(map[string]struct{}),
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
