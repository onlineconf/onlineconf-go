package onlineconf

// ConfigParam is the interface that must be implemented by all
// config parameters structures.
//
// Different config parameters usually have different default values.
type ConfigParam interface {
	GetPath() *ParamPath
	SetPath(*ParamPath) error
}

// ConfigParamInt implements ConfigParam interface with integer parameters.
type ConfigParamInt struct {
	path         *ParamPath
	defaultValue int
}

var _ ConfigParam = (*ConfigParamInt)(nil) // compile time interface check

// NewConfigParamInt returns a ConfigParamInt with parameter path `path` with default value `defaultValue`
func NewConfigParamInt(path string, defaultValue int) (*ConfigParamInt, error) {
	validPath, err := NewParamPath(path)
	if err != nil {
		return nil, err
	}

	confParam := &ConfigParamInt{path: validPath, defaultValue: defaultValue}
	return confParam, nil
}

// MustConfigParamInt returns a ConfigParamInt with parameter path `path` with default value `defaultValue`.
// Panics on error.
func MustConfigParamInt(path string, defaultValue int) *ConfigParamInt {
	configParam, err := NewConfigParamInt(path, defaultValue)
	if err != nil {
		panic(err)
	}

	return configParam
}

// GetPath returns parameter path of ConfigParamInt
func (param *ConfigParamInt) GetPath() *ParamPath {
	return param.path
}

// SetPath sets parameter path for ConfigParamInt
func (param *ConfigParamInt) SetPath(newPath *ParamPath) error {
	if err := newPath.IsValid(); err != nil {
		return err
	}
	param.path = newPath
	return nil
}

// ConfigParamString implements ConfigParam interface with string parameters.
type ConfigParamString struct {
	path         *ParamPath
	defaultValue string
}

var _ ConfigParam = (*ConfigParamString)(nil) // compile time interface check

// NewConfigParamString returns a ConfigParamString with parameter path `path` with default value `defaultValue`
func NewConfigParamString(path string, defaultValue string) (*ConfigParamString, error) {
	validPath, err := NewParamPath(path)
	if err != nil {
		return nil, err
	}

	confParam := &ConfigParamString{path: validPath, defaultValue: defaultValue}
	return confParam, nil
}

// MustConfigParamString returns a ConfigParamString with parameter path `path` with default value `defaultValue`
// Panics on error.
func MustConfigParamString(path string, defaultValue string) *ConfigParamString {
	configParam, err := NewConfigParamString(path, defaultValue)
	if err != nil {
		panic(err)
	}

	return configParam
}

// GetPath returns parameter path of ConfigParamString
func (param *ConfigParamString) GetPath() *ParamPath {
	return param.path
}

// SetPath sets parameter path for ConfigParamString
func (param *ConfigParamString) SetPath(newPath *ParamPath) error {
	if err := newPath.IsValid(); err != nil {
		return err
	}
	param.path = newPath
	return nil
}

// ConfigParamBool implements ConfigParam interface with bool parameters.
type ConfigParamBool struct {
	path         *ParamPath
	defaultValue bool
}

var _ ConfigParam = (*ConfigParamBool)(nil) // compile time interface check

// NewConfigParamBool returns a ConfigParamBool with parameter path `path` with default value `defaultValue`
func NewConfigParamBool(path string, defaultValue bool) (*ConfigParamBool, error) {
	validPath, err := NewParamPath(path)
	if err != nil {
		return nil, err
	}

	confParam := &ConfigParamBool{path: validPath, defaultValue: defaultValue}
	return confParam, nil
}

// MustConfigParamBool returns a ConfigParamBool with parameter path `path` with default value `defaultValue`
// Panics on error.
func MustConfigParamBool(path string, defaultValue bool) *ConfigParamBool {
	configParam, err := NewConfigParamBool(path, defaultValue)
	if err != nil {
		panic(err)
	}

	return configParam
}

// GetPath returns parameter path of ConfigParamBool
func (param *ConfigParamBool) GetPath() *ParamPath {
	return param.path
}

// SetPath sets parameter path for ConfigParamBool
func (param *ConfigParamBool) SetPath(newPath *ParamPath) error {
	if err := newPath.IsValid(); err != nil {
		return err
	}
	param.path = newPath
	return nil
}

// ParamsPrefix inplace appends prefix `prefix` for slice of ConfigParams `confParams`
func ParamsPrefix(prefix *ParamPath, confParams []ConfigParam) error {
	for _, confParam := range confParams {
		newPath := prefix.Join(confParam.GetPath())
		err := confParam.SetPath(newPath)
		if err != nil {
			return err
		}
	}

	return nil
}
