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

	StringParams map[string]string
	IntParams    map[string]int
	BoolParams   map[string]bool

	ModuleConfig []ConfigParam
}

// NewModule parses cdb file and copies all content to local maps.
// Module returned by this method will never be updated
func NewModule(reader io.ReaderAt) (*Module, error) {

	cdbReader, err := cdb.New().GetReader(reader)
	if err != nil {
		return nil, fmt.Errorf("Cant cant cdb reader for module: %w", err)
	}

	module := &Module{
		CDB: cdbReader,
	}

	return module, nil
}

// NewModule parses cdb file and copies all content to local maps.
// Module returned by this method will never be updated
func NewPredeclaredModule(reader io.ReaderAt, paramsDescriptsion []ConfigParam) (*Module, error) {
	module, err := NewModule(reader)
	if err != nil {
		return nil, err
	}

	// todo fill params

	return module, nil
}

func (m *Module) fillParams(cdb cdb.Reader) error {
	cdbIter, err := cdb.Iterator()
	if err != nil {
		return fmt.Errorf("can't get cdb iterator: %w", err)
	}

	for {
		record := cdbIter.Record()
		if record == nil {
			break
		}

		key, err := record.KeyBytes()
		if err != nil {
			return fmt.Errorf("can't read cdb key: %w", err)
		}

		val, err := record.ValueBytes()
		if err != nil {
			return fmt.Errorf("can't read cdb value: %w", err)
		}

		if len(val) < 1 {
			return fmt.Errorf("onlineconf value must contain at least 1 byte: `typeByte|ParamData`")
		}

		// log.Printf("oc parsing: %s %s", string(key), string(val))

		// val's first byte defines datatype of config value
		// onlineconf currently knows 's' and 'j' data types
		paramTypeByte := val[0]
		keyStr := string(key)
		// valStr := string(val[1:])
		// такого треша, конечно же, не было бы, если бы в онлайнконфе
		// была бы типизация
		if paramTypeByte == 's' { // params type string
			// m.parseSimpleParams(keyStr, valStr)
		} else if paramTypeByte == 'j' { // params type JSON
			// err := m.parseJSONParams(keyStr, valStr)
			// if err != nil {
			// 	return err
			// }
		} else {
			return fmt.Errorf("unknown paramTypeByte: %#v for key %s", paramTypeByte, keyStr)
		}

		if !cdbIter.HasNext() {
			break
		}

		_, err = cdbIter.Next()
		if err != nil {
			return fmt.Errorf("can't get next cdb record: %w", err)
		}
	}

	return nil
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
