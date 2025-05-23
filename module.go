package onlineconf

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"path"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/colinmarc/cdb"
	"github.com/my-mail-ru/exp/mmap"
)

// CDB path defaults
const (
	DefaultOnlineConfPath = "/usr/local/etc/onlineconf"
	DefaultOnlineConfExt  = ".cdb"
)

// Errors returned by Get*Err methods
var (
	ErrNotFound          = errors.New("onlineconf: key not found")
	ErrFormatIsNotString = errors.New("format is not a string")
	ErrFormatIsNotJSON   = errors.New("format is not JSON")
)

// Module represents a CDB configuration database.
type Module struct {
	mutex         sync.RWMutex
	name          string // CDB relative file name, used in error messages. not a name passed to OpenModule
	filename      string // full CDB file path
	cache         valueCache
	mmappedFile   *mmap.ReaderAt
	cdb           *cdb.CDB
	subscriptions map[subscriptionKey]subscription
}

var modCache syncCache[*Module]

// OpenModule opens a CDB configuration database.
//
// If the name argument doesn't contain a filesystem path separator (usually '/'),
// it is treated as relative to [DefaultOnlineConfPath] (/usr/local/etc/onlineconf),
// in the other case - as an absolute or a relative path.
// If no extension is specified, the default `.cdb` extension is appended.
// There's no way to specify a file without an extension.
//
// Calling OpenModule multiple times with the same file argument (regardless of the
// actual mode used, e.g., relative or absolute) will always return the same value.
//
// The file opened is tracked for changes using the [fsnotify] library and is reloaded when a change is detected.
// See [Module.Subscribe], [Module.SubscribeChan], [Module.SubscribeSubtree] and [Module.SubscribeChanSubtree] methods
// for a description of high-level value change notification mechanism.
//
// Currently, there's no way to "close" a [Module].
func OpenModule(name string) (*Module, error) {
	cached, inProgressByName, ok := modCache.load(name)
	if ok {
		return cached, nil
	}

	filename, err := modFileName(name)
	if err != nil {
		return nil, err
	}

	var inProgressByFileName chan<- struct{}

	if filename != name {
		cached, inProgressByFileName, ok = modCache.load(filename)
		if ok {
			modCache.store(name, inProgressByName, cached) // re-cache by relative/short name if already cached by fully qualified name
			return cached, nil
		}
	}

	module := &Module{
		name:     filepath.Base(filename),
		filename: filename,
	}

	if err := module.reopen(); err != nil {
		return nil, err
	}

	modCache.store(name, inProgressByName, module)

	if filename != name {
		modCache.store(filename, inProgressByFileName, module)
	}

	return module, initWatcher(filepath.Dir(filename))
}

func modFileName(name string) (string, error) {
	if !strings.ContainsRune(name, filepath.Separator) {
		name = filepath.Join(DefaultOnlineConfPath, name)
	}

	if filepath.Ext(name) == "" {
		name += DefaultOnlineConfExt
	}

	filename, err := filepath.Abs(name)
	if err != nil {
		return "", fmt.Errorf("OpenModule(%s): error getting absolute path: %w", name, err)
	}

	filename, err = filepath.EvalSymlinks(filename)
	if err != nil {
		return "", fmt.Errorf("OpenModule(%s): error resolving symlinks: %w", name, err)
	}

	return filename, nil
}

func (m *Module) reopen() error {
	log.Printf("onlineconf: reopen %s", m.filename)
	m.mutex.Lock()
	defer m.mutex.Unlock()

	oldMmappedFile := m.mmappedFile

	mmappedFile, err := mmap.Open(m.filename)
	if err != nil {
		return fmt.Errorf("mmap.Open(%s): %w", m.filename, err)
	}

	cdb, err := cdb.New(mmappedFile, nil)
	if err != nil {
		mmappedFile.Close()
		return fmt.Errorf("cdb.New(%s): %w", m.filename, err)
	}

	m.cdb = cdb

	if oldMmappedFile != nil {
		oldMmappedFile.Close()
	}

	m.cache.init()
	m.processSubscriptions()

	return nil
}

// get never returns ErrNotFound.
// the first returned value is type:
//
//	0   - not found
//	's' - any text value including numbers, strings, and bools (since onlineconf UI doesn't support strict typing)
//	'j' - JSON or YAML (which is converted to JSON in the updater)
func (m *Module) get(path string) (byte, []byte, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	data, err := m.getRaw(path)
	if len(data) == 0 {
		return 0, nil, err
	}

	return data[0], data[1:], nil
}

func (m *Module) getRaw(path string) ([]byte, error) {
	data, err := m.cdb.Get(s2b(path))
	if err != nil {
		return nil, fmt.Errorf("cdb.Get(%s:%s): %w", m.filename, path, err)
	}

	return data, nil
}

// Path returns its argument.
func (*Module) Path(path string) string {
	return path
}

// GetStringErr reads a string value of a named parameter from the module.
//
// If no such value exists, [ErrNotFound] is returned.
// If the value is not a string, [ErrFormatIsNotString] is returned.
func (m *Module) GetStringErr(path string) (string, error) {
	switch format, data, err := m.get(path); {
	case err != nil:
		return "", err
	case format == 0:
		return "", ErrNotFound
	case format == 's':
		return b2s(data), nil
	default:
		return "", fmt.Errorf("%s:%s: %w", m.name, path, ErrFormatIsNotString)
	}
}

// GetStringIfExists reads a string value of a named parameter from the module.
//
// It returns the boolean true if the parameter exists and is a string.
// In the other case, it returns the boolean false and an empty string.
//
// CDB errors and format mismatches are logged.
func (m *Module) GetStringIfExists(path string) (string, bool) {
	switch format, data, err := m.get(path); {
	case err != nil:
		log.Print(err)
		return "", false
	case format == 0:
		return "", false
	case format == 's':
		return b2s(data), true
	default:
		log.Printf("%s:%s: %v", m.name, path, ErrFormatIsNotString)
		return "", false
	}
}

// GetString reads a string value of a named parameter from the module.
//
// It returns this value if the parameter exists and is a string.
// In all other cases (no value found, format is not a string, CDB error)
// the default value specified in the second argument is returned.
//
// CDB errors and format mismatches are logged.
func (m *Module) GetString(path string, dfl string) string {
	if val, ok := m.GetStringIfExists(path); ok {
		return val
	}

	return dfl
}

// GetIntErr reads an int value of a named parameter from the module.
//
// If no such value exists, [ErrNotFound] is returned.
// If the value is not a string, [ErrFormatIsNotString] is returned.
// If the value doesn't represent a valid int (see [strconv.Atoi]), a wrapped parsing error is returned.
func (m *Module) GetIntErr(path string) (int, error) {
	str, err := m.GetStringErr(path)
	if err != nil {
		return 0, err
	}

	i, err := strconv.Atoi(str)
	if err != nil {
		return 0, fmt.Errorf("%s:%s: %w", m.name, path, err)
	}

	return i, nil
}

// GetIntIfExists reads an integer value of a named parameter from the module.
//
// It returns this value and the boolean true if the parameter exists and is
// a valid string representation of an integer value.
// In the other case, it returns the boolean false and 0.
// In all other cases (no value found, format is not a string, CDB error, [strconv.Atoi] error)
// the default value specified in the second argument is returned.
//
// CDB errors, format mismatches, and parsing errors are logged.
func (m *Module) GetIntIfExists(path string) (int, bool) {
	i, err := m.GetIntErr(path)
	if err != nil {
		if err != ErrNotFound { // ErrNotFound is returned unwrapped
			log.Print(err)
		}

		return 0, false
	}

	return i, true
}

// GetInt reads an integer value of a named parameter from the module.
// It returns this value if the parameter exists and represents a valid int value (see [strconv.Atoi]).
//
// CDB errors, format mismatches, and parsing errors are logged.
func (m *Module) GetInt(path string, dfl int) int {
	if val, ok := m.GetIntIfExists(path); ok {
		return val
	}

	return dfl
}

// GetBoolErr reads a boolean value of a named parameter from the module.
//
// false is returned when the value exists and is empty of "0", true is
// returned when the value exists but is neither empty nor "0".
func (m *Module) GetBoolErr(path string) (bool, error) {
	str, err := m.GetStringErr(path)
	if err != nil {
		return false, err
	}

	if len(str) == 0 || str == "0" { // preserve compatibility with mature perl projects. please do not add new values.
		return false, nil
	}

	return true, nil
}

// GetBoolIfExists reads an integer value of a named parameter from the module.
//
// (false, true) is returned when the value exists and is empty of "0", (true, true) is
// returned when the value exists but is neither empty nor "0".
// If there's no such value, or it isn't a string, (false, false) is returned.
//
// CDB errors and format mismatches are logged.
func (m *Module) GetBoolIfExists(path string) (bool, bool) {
	b, err := m.GetBoolErr(path)
	if err != nil {
		if err != ErrNotFound {
			log.Print(err)
		}

		return false, false
	}

	return b, true
}

// GetBool reads a boolean value of a named parameter from the module.
//
// Calls [Module.GetBoolIfExists] internally. The default value `dfl` is returned when
// [Module.GetBoolIfExists] returns (false, false).
func (m *Module) GetBool(path string, dfl bool) bool {
	if val, ok := m.GetBoolIfExists(path); ok {
		return val
	}

	return dfl
}

// GetDurationErr reads a [time.Duration] value of a named parameter from the module.
//
// Calls [Module.GetStringErr] internally.
// The duration value is parsed using [time.ParseDuration].
// For compatibility with other implementations, when no unit suffix is specified,
// a value is treated as a duration in seconds.
func (m *Module) GetDurationErr(path string) (time.Duration, error) {
	str, err := m.GetStringErr(path)
	if err != nil {
		return 0, err
	}

	d, err := parseDuration(str)
	if err != nil {
		return 0, fmt.Errorf("%s:%s: %w", m.name, path, err)
	}

	return d, nil
}

func parseDuration(s string) (time.Duration, error) {
	if strings.ContainsAny(s, "hms") { // fractions of second contain 's' too
		return time.ParseDuration(s)
	}

	return time.ParseDuration(s + "s")
}

// GetDurationIsExists reads a [time.Duration] value of a named parameter from the module.
//
// Calls [Module.GetDurationErr] internally. In the case of an error (0, false) is returned.
// Errors other than [ErrNotFound] are logged.
func (m *Module) GetDurationIsExists(path string) (time.Duration, bool) {
	d, err := m.GetDurationErr(path)
	if err != nil {
		if err != ErrNotFound {
			log.Print(err)
		}

		return 0, false
	}

	return d, true
}

// GetDuration reads a [time.Duration] value of a named parameter from the module.
// Calls [Module.GetDurationIsExists] internally. The default value `dfl` is returned when
// [Module.GetDurationIsExists] returns (0, false).
func (m *Module) GetDuration(path string, dfl time.Duration) time.Duration {
	if val, ok := m.GetDurationIsExists(path); ok {
		return val
	}

	return dfl
}

// GetFloatErr reads a float64 value of a named parameter from the module.
// If no such value exists, [ErrNotFound] is returned.
// If the value is not a string, [ErrFormatIsNotString] is returned.
// If the value doesn't represent a valid float64 (see [strconv.ParseFloat]),
// a wrapped parsing error is returned.
func (m *Module) GetFloatErr(path string) (float64, error) {
	str, err := m.GetStringErr(path)
	if err != nil {
		return 0, err
	}

	f, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return 0, fmt.Errorf("%s:%s: %w", m.name, path, err)
	}

	return f, nil
}

// GetFloatIfExists reads a float64 value of a named parameter from the module.
// It returns this value and the boolean true if the parameter exists and is a valid
// string representation of a float64 value.
// In the other case, it returns the boolean false and 0.
// In all other cases (no value found, format is not a string, CDB error, [strconv.Atoi] error)
// the default value specified in the second argument is returned.
//
// CDB errors, format mismatches, and parsing errors are logged.
func (m *Module) GetFloatIfExists(path string) (float64, bool) {
	f, err := m.GetFloatErr(path)
	if err != nil {
		if err != ErrNotFound {
			log.Print(err)
		}

		return 0, false
	}

	return f, true
}

// GetFloat reads a float64 value of a named parameter from the module.
// Calls [Module.GetFloatIfExists] internally. The default value `dfl` is returned
// when [Module.GetFloatIfExists] returns (0, false).
func (m *Module) GetFloat(path string, dfl float64) float64 {
	if val, ok := m.GetFloatIfExists(path); ok {
		return val
	}

	return dfl
}

// GetStringsErr reads a []string value of a named parameter from the module.
// It returns this value if the parameter exists and is a comma-separated
// string or a JSON array.
// In all other cases it returns a default value provided in the second
// argument and an error.
//
// Strings returned are cached internally until the configuration is updated.
func (m *Module) GetStringsErr(path string, dfl []string) ([]string, error) {
	var ret []string

	rv := reflect.ValueOf(&ret).Elem()
	if m.cache.get(path, rv) {
		return ret, nil
	}

	format, data, err := m.get(path)
	if err != nil {
		return dfl, err
	}

	switch format {
	case 0:
		return dfl, ErrNotFound
	case 's':
		items := strings.Split(b2s(data), ",")
		ret = make([]string, 0, len(items))

		for _, item := range items {
			if trimmed := strings.TrimSpace(item); trimmed != "" {
				ret = append(ret, trimmed)
			}
		}

		m.cache.set(path, rv)

		return ret, nil
	case 'j':
		if err := json.Unmarshal(data, &ret); err != nil {
			return dfl, fmt.Errorf("%s:%s: failed to unmarshal JSON: %w", m.name, path, err)
		}

		m.cache.set(path, rv)

		return ret, nil

	default:
		return dfl, fmt.Errorf("%s:%s: unexpected format '%c'", m.name, path, format)
	}
}

// getStringsRaw is used internally for recursive subscriptions.
// value cache isn't used. Never returns ErrNotFound.
func (m *Module) getStringsRaw(path string) ([]string, error) {
	data, err := m.getRaw(path)
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, nil
	}

	if data[0] != 'j' {
		return nil, fmt.Errorf("%s:%s: %w", m.name, path, ErrFormatIsNotJSON)
	}

	var ret []string

	if err := json.Unmarshal(data[1:], &ret); err != nil {
		return nil, fmt.Errorf("%s:%s: failed to unmarshal JSON: %w", m.name, path, err)
	}

	return ret, nil
}

// GetStrings reads a []string value of a named parameter from the module.
// Calls [Module.GetStringsErr] internally. All errors but [ErrNotFound] are logged.
func (m *Module) GetStrings(path string, dfl []string) []string {
	ret, err := m.GetStringsErr(path, dfl)
	if err != nil {
		if err != ErrNotFound {
			log.Print(err)
		}

		return dfl
	}

	return ret
}

// GetStruct reads a structured value of a named parameter from the module.
// valuePtr should contain a pointer to a variable of a type compatible
// with the JSON value of the parameter specified.
//
// A value is unmarshaled from a JSON representation using [json.Unmarshal] and is cached internally
// until the configuration is updated. The cached value is a shallow-copy of *valuePtr, so be careful
// with pointers/slices, since values pointed/contained are shared.
//
// In the case of an unmarshal error or if the parameter does not exist, the value isn't clobbered,
// so you can place the default value in a variable pointed to by the valuePtr argument and ignore
// the bool value returned.
//
// Never returns ErrNotFound.
func (m *Module) GetStruct(path string, valuePtr interface{}) (bool, error) {
	rv := reflect.ValueOf(valuePtr)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return false, fmt.Errorf("%s:%s: GetStruct accepts a non-nil pointer", m.name, path)
	}

	rv = rv.Elem()
	if m.cache.get(path, rv) {
		return true, nil
	}

	format, data, err := m.get(path)
	if err != nil {
		return false, err
	}

	switch format {
	case 0:
		return false, nil
	case 'j':
		val := reflect.New(rv.Type()) // ensure that the default value isn't clobbered by a partially failed unmarshal
		if err := json.Unmarshal(data, val.Interface()); err != nil {
			return false, fmt.Errorf("%s:%s: failed to unmarshal JSON: %w", m.name, path, err)
		}

		rv.Set(val.Elem())
		m.cache.set(path, rv)

		return true, nil
	default:
		return false, fmt.Errorf("%s:%s: %w", m.name, path, ErrFormatIsNotJSON)
	}
}

// Subtree returns a [Subtree] of the module rooted at the specified prefix.
// The prefix is normalized using [path.Clean].
func (m *Module) Subtree(prefix string) *Subtree {
	return &Subtree{
		mod:    m,
		prefix: cleanPrefix(path.Clean(prefix)),
	}
}
