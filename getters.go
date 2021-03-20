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

func (m *Module) string(configParam ConfigParam) (string, error) {
	paramPath := configParam.GetPath()
	param, ok := m.getStringCached(paramPath)
	if ok {
		return param, nil
	}

	if m.getCachedNotExist(paramPath) {
		return "", newErrKeyNotFound(paramPath.path)
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

// String returns value of a named parameter from the module.
func (m *Module) String(configParam *ConfigParamString) (string, error) {
	paramStr, err := m.string(configParam)

	if err != nil {
		return configParam.defaultValue, err
	}

	return paramStr, nil
}

func (m *Module) readString(paramPath *ParamPath) (string, error) {
	readBytes, err := m.readBytes(paramPath)
	if err != nil {
		return "", err
	}

	return string(readBytes), nil
}

// Int returns value of a named parameter from the module.
func (m *Module) Int(configParam *ConfigParamInt) (int, error) {
	paramPath := configParam.path
	param, ok := m.getIntCached(paramPath)
	if ok {
		return param, nil
	}

	paramStr, err := m.string(configParam)
	if err != nil {
		return configParam.defaultValue, err
	}

	paramInt, err := strconv.Atoi(paramStr)
	if err != nil {
		return configParam.defaultValue, err
	}

	m.setIntCached(paramPath, paramInt)

	return paramInt, nil
}

// Bool returns bool interpretation of param.
func (m *Module) Bool(configParam *ConfigParamBool) (bool, error) {
	paramPath := configParam.path
	param, ok := m.getBoolCached(paramPath)
	if ok {
		return param, nil
	}

	paramStr, err := m.string(configParam)
	if err != nil {
		return configParam.defaultValue, err
	}

	paramBool := len(paramStr) != 0 && paramStr != "0"
	m.setBoolCached(paramPath, paramBool)

	return paramBool, nil
}
