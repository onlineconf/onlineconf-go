package onlineconf

type ConfigParam interface {
	GetPath() *ParamPath
	SetPath(*ParamPath) error
}

type ConfigParamInt struct {
	Path     *ParamPath
	Default  int
	Required bool
}

var _ ConfigParam = (*ConfigParamInt)(nil) // compile time interface check

func (param *ConfigParamInt) GetPath() *ParamPath {
	return param.Path
}

func (param *ConfigParamInt) SetPath(newPath *ParamPath) error {
	if err := newPath.IsValid(); err != nil {
		return err
	}
	param.Path = newPath
	return nil
}

type ConfigParamString struct {
	Path     *ParamPath
	Default  string
	Required bool
}

var _ ConfigParam = (*ConfigParamString)(nil) // compile time interface check

func (param *ConfigParamString) GetPath() *ParamPath {
	return param.Path
}

func (param *ConfigParamString) SetPath(newPath *ParamPath) error {
	if err := newPath.IsValid(); err != nil {
		return err
	}
	param.Path = newPath
	return nil
}

type ConfigParamBool struct {
	Path     *ParamPath
	Default  bool
	Required bool
}

var _ ConfigParam = (*ConfigParamBool)(nil) // compile time interface check

func (param *ConfigParamBool) GetPath() *ParamPath {
	return param.Path
}

func (param *ConfigParamBool) SetPath(newPath *ParamPath) error {
	if err := newPath.IsValid(); err != nil {
		return err
	}
	param.Path = newPath
	return nil
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
