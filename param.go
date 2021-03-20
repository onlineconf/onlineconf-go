package onlineconf

type ConfigParam interface {
	GetPath() *ParamPath
}

type ConfigParamInt struct {
	path    *ParamPath
	Default int
}

var _ ConfigParam = (*ConfigParamInt)(nil) // compile time interface check

func NewConfigParamInt(path string, defaultValue int) (*ConfigParamInt, error) {
	validPath, err := NewParamPath(path)
	if err != nil {
		return nil, err
	}

	confParam := &ConfigParamInt{path: validPath, Default: defaultValue}
	return confParam, nil
}

func MustConfigParamInt(path string, defaultValue int) *ConfigParamInt {
	configParam, err := NewConfigParamInt(path, defaultValue)
	if err != nil {
		panic(err)
	}

	return configParam
}

func (param *ConfigParamInt) GetPath() *ParamPath {
	return param.path
}

type ConfigParamString struct {
	Path    *ParamPath
	Default string
}

var _ ConfigParam = (*ConfigParamString)(nil) // compile time interface check

func NewConfigParamString(path string, defaultValue string) (*ConfigParamString, error) {
	validPath, err := NewParamPath(path)
	if err != nil {
		return nil, err
	}

	confParam := &ConfigParamString{Path: validPath, Default: defaultValue}
	return confParam, nil
}

func MustConfigParamString(path string, defaultValue string) *ConfigParamString {
	configParam, err := NewConfigParamString(path, defaultValue)
	if err != nil {
		panic(err)
	}

	return configParam
}

func (param *ConfigParamString) GetPath() *ParamPath {
	return param.Path
}

type ConfigParamBool struct {
	Path     *ParamPath
	Default  bool
	Required bool
}

var _ ConfigParam = (*ConfigParamBool)(nil) // compile time interface check

func NewConfigParamBool(path string, defaultValue bool) (*ConfigParamBool, error) {
	validPath, err := NewParamPath(path)
	if err != nil {
		return nil, err
	}

	confParam := &ConfigParamBool{Path: validPath, Default: defaultValue}
	return confParam, nil
}

func MustConfigParamBool(path string, defaultValue bool) *ConfigParamBool {
	configParam, err := NewConfigParamBool(path, defaultValue)
	if err != nil {
		panic(err)
	}

	return configParam
}

func (param *ConfigParamBool) GetPath() *ParamPath {
	return param.Path
}

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
