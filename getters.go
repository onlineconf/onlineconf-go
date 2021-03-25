package onlineconf

import (
	"errors"
	"fmt"
	"strconv"
)

// ErrBrokenCDB returned for invalid onlineconf parameters.
// Valid onlineconf module value must contain at least
// one byte for parameter type e. g. `s` for string params and `j` for JSON params.
var ErrBrokenCDB = errors.New("cdb record is not valid for onlineconf module")

// ErrKeyNotFound returned for parameter that was not defined in onlineconf module.
type ErrKeyNotFound struct {
	path string
}

func (err *ErrKeyNotFound) Error() string {
	return fmt.Sprintf("key `%s` not found in onlineconf", err.path)
}

// IsErrKeyNotFound checks if error is ErrKeyNotFound
func IsErrKeyNotFound(err error) bool {
	_, ok := err.(*ErrKeyNotFound)
	return ok
}

func newErrKeyNotFound(path string) *ErrKeyNotFound {
	return &ErrKeyNotFound{
		path: path,
	}
}

// ReadBytes returns raw bytes that was read ftom CDB file. No caching. Error returned if any.
func (m *Module) ReadBytes(paramPath *ParamPath) ([]byte, error) {
	path := paramPath.path
	stringData, err := m.CDB.Get([]byte(path))

	if err != nil {
		return nil, fmt.Errorf("onlineconf module readSregin: %w", err)
	}

	if stringData == nil && err == nil {
		return nil, newErrKeyNotFound(path)
	}

	if len(stringData) < 1 {
		return nil, ErrBrokenCDB
	}

	return stringData[1:], nil
}

func (m *Module) readString(paramPath *ParamPath) (string, error) {
	readBytes, err := m.ReadBytes(paramPath)
	if err != nil {
		return "", err
	}

	return string(readBytes), nil
}

// string tries to retrieve config value from cached parameters map.
// Lookup CDB file in case parameter was not cached and saves it to cache.
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
// Module getters returns no error due to reading from mmapeed file will never return error.
// In case of reading error on mmapped file SIGSEGV will be sent to the process.
func (m *Module) String(configParam *ConfigParamString) string {
	paramStr, err := m.string(configParam)

	if err != nil {
		return configParam.defaultValue
	}

	return paramStr
}

// Int returns value of a named parameter from the module.
// Module getters returns no error due to reading from mmapeed file will never return error.
// In case of reading error on mmapped file SIGSEGV will be sent to the process.
func (m *Module) Int(configParam *ConfigParamInt) int {
	paramPath := configParam.path
	param, ok := m.getIntCached(paramPath)
	if ok {
		return param
	}

	paramStr, err := m.string(configParam)
	if err != nil {
		return configParam.defaultValue
	}

	paramInt, err := strconv.Atoi(paramStr)
	if err != nil {
		return configParam.defaultValue
	}

	m.setIntCached(paramPath, paramInt)

	return paramInt
}

// Bool returns bool interpretation of param.
// Module getters returns no error due to reading from mmapeed file will never return error.
// In case of reading error on mmapped file SIGSEGV will be sent to the process.
func (m *Module) Bool(configParam *ConfigParamBool) bool {
	paramPath := configParam.path
	param, ok := m.getBoolCached(paramPath)
	if ok {
		return param
	}

	paramStr, err := m.string(configParam)
	if err != nil {
		return configParam.defaultValue
	}

	paramBool := len(paramStr) != 0 && paramStr != "0"
	m.setBoolCached(paramPath, paramBool)

	return paramBool
}
