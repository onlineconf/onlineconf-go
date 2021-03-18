package onlineconf

import (
	"errors"
	"fmt"
	"strconv"
)

// ErrBrokenCDB returned for empty onlineconf values.
// Valid onlineconf module value must contain at least
// one byte for parameter type e. g. `s` for string params and `j` for JSON params
var ErrBrokenCDB = errors.New("cdb record is not valid for onlineconf module")

type ErrKeyNotFound struct {
	path string
}

func (err *ErrKeyNotFound) Error() string {
	return fmt.Sprintf("key `%s` not found in onlineconf", err.path)
}

func IsErrKeyNotFound(err error) bool {
	_, ok := err.(*ErrKeyNotFound)
	return ok
}

func newErrKeyNotFound(path string) *ErrKeyNotFound {
	return &ErrKeyNotFound{
		path: path,
	}
}

func (m *Module) readBytes(paramPath *ParamPath) ([]byte, error) {
	path := paramPath.path
	stringData, err := m.CDB.Get([]byte(path))
	fmt.Printf("reading key: %s, data: %s, err: %#v\n", path, string(stringData), err)
	if stringData == nil && err == nil {
		return nil, newErrKeyNotFound(path)
	}

	if err != nil {
		return nil, fmt.Errorf("onlineconf module readSregin: %w", err)
	}

	if len(stringData) <= 1 {
		return nil, ErrBrokenCDB
	}

	return stringData[1:], nil
}

// String returns value of a named parameter from the module.
// It returns the boolean true if the parameter exists and is a string.
// In the other case it returns the boolean false and an empty string.
func (m *Module) String(paramPath *ParamPath) (string, error) {
	param, ok := m.getStringCached(paramPath)
	if ok {
		return param, nil
	}

	paramStr, err := m.readString(paramPath)
	if IsErrKeyNotFound(err) {
		m.setCachedNotExisting(paramPath)
		return "", err
	}

	if err != nil {
		return "", err
	}

	m.setStringCached(paramPath, paramStr)

	return paramStr, nil
}

func (m *Module) readString(paramPath *ParamPath) (string, error) {
	readBytes, err := m.readBytes(paramPath)
	if err != nil {
		return "", err
	}

	return string(readBytes), nil
}

// StringWithDef returns value of a named parameter from the module.
// It returns the boolean true if the parameter exists and is a string.
// In the other case it returns the boolean false and an empty string.
func (m *Module) StringWithDef(paramPath *ParamPath, defaultValue string) (string, error) {
	param, err := m.String(paramPath)
	if IsErrKeyNotFound(err) {
		return defaultValue, err
	}
	return param, err
}

// Int returns value of a named parameter from the module.
// It returns the boolean true if the parameter exists and is an int.
// In the other case it returns the boolean false and zero.
func (m *Module) Int(paramPath *ParamPath) (int, error) {
	param, ok := m.getIntCached(paramPath)
	if ok {
		return param, nil
	}

	paramStr, err := m.String(paramPath)
	if err != nil {
		return 0, err
	}

	paramInt, err := strconv.Atoi(paramStr)
	if err != nil {
		return paramInt, err
	}

	m.setIntCached(paramPath, paramInt)

	return paramInt, nil
}

// IntWithDef returns value of a named parameter from the module.
// It returns the boolean true if the parameter exists and is an int.
// In the other case it returns the boolean false and zero.
func (m *Module) IntWithDef(path *ParamPath, defaultValue int) (int, error) {
	param, err := m.Int(path)
	if IsErrKeyNotFound(err) {
		return defaultValue, err
	}
	return param, err
}

// Bool returns bool interpretation of param.
// If length of string parameter with same path is greater than 0,
// returns true. In other case false.
func (m *Module) Bool(paramPath *ParamPath) (bool, error) {
	param, ok := m.getBoolCached(paramPath)
	if ok {
		return param, nil
	}

	paramStr, err := m.String(paramPath)
	if err != nil {
		return false, err
	}

	paramBool := paramStr == "1"
	m.setBoolCached(paramPath, paramBool)

	return paramBool, nil
}

// BoolWithDef the same as Bool but is no such parameter? it returns default value
func (m *Module) BoolWithDef(path *ParamPath, defaultValue bool) (bool, error) {
	param, err := m.Bool(path)
	if IsErrKeyNotFound(err) {
		return defaultValue, err
	}
	return param, nil
}
