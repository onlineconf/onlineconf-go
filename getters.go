package onlineconf

import "fmt"

// String returns value of a named parameter from the module.
// It returns the boolean true if the parameter exists and is a string.
// In the other case it returns the boolean false and an empty string.
func (m *Module) String(path string) (string, bool) {
	param, ok := m.StringParams[path]
	return param, ok
}

// StringWithDef returns value of a named parameter from the module.
// It returns the boolean true if the parameter exists and is a string.
// In the other case it returns the boolean false and an empty string.
func (m *Module) StringWithDef(path string, defaultValue string) (string, bool) {
	param, ok := m.StringParams[path]
	if !ok {
		return defaultValue, ok
	}
	return param, ok
}

// MustString returns value of a named parameter from the module.
// It panics if no such parameter or this parameter is not a string.
func (m *Module) MustString(path string) string {
	param, ok := m.StringParams[path]
	if !ok {
		panic(fmt.Errorf("Missing required parameter in onlineconf or cant parse it %s", path))
	}
	return param
}

// Int returns value of a named parameter from the module.
// It returns the boolean true if the parameter exists and is an int.
// In the other case it returns the boolean false and zero.
func (m *Module) Int(path string) (int, bool) {
	param, ok := m.IntParams[path]
	return param, ok
}

// IntWithDef returns value of a named parameter from the module.
// It returns the boolean true if the parameter exists and is an int.
// In the other case it returns the boolean false and zero.
func (m *Module) IntWithDef(path string, defaultValue int) (int, bool) {
	param, ok := m.IntParams[path]
	if !ok {
		return defaultValue, ok
	}
	return param, ok
}

// MustInt returns value of a named parameter from the module.
// It panics if no such parameter or this parameter is not an int
func (m *Module) MustInt(path string) int {
	param, ok := m.IntParams[path]
	if !ok {
		panic(fmt.Errorf("Missing required parameter in onlineconf or cant parse it %s", path))
	}
	return param
}

// Bool returns bool interpretation of param.
// If length of string parameter with same path is greater than 0,
// returns true. In other case false.
func (m *Module) Bool(path string) (bool, bool) {
	param, ok := m.String(path)
	if !ok {
		return false, false
	}
	return (len(param) > 0), ok
}

// BoolWithDef the same as Bool but is no such parameter? it returns default value
func (m *Module) BoolWithDef(path string, defaultValue bool) (bool, bool) {
	param, ok := m.Bool(path)
	if !ok {
		return defaultValue, false
	}
	return param, ok
}

// MustBool returns value of a named parameter from the module.
// It panics if no such parameter or this parameter is not a string.
func (m *Module) MustBool(path string) bool {
	param, ok := m.Bool(path)
	if !ok {
		panic(fmt.Errorf("Missing required parameter in onlineconf or cant parse it %s", path))
	}
	return param
}

// MapStringInterface
// Interfaces will not be copied!
func (m *Module) MapStringInterface(path string) (map[string]interface{}, bool) {
	param, ok := m.MapStringInterfaceParams[path]
	if ok {
		clone := make(map[string]interface{}, len(param))
		for k, v := range param {
			clone[k] = v
		}
		return clone, ok
	}
	return param, ok
}

// MapStringInterfaceWithDef default valur will not be copied!
func (m *Module) MapStringInterfaceWithDef(path string, defaultValue map[string]interface{}) (map[string]interface{}, bool) {
	param, ok := m.MapStringInterface(path)
	if !ok {
		return defaultValue, ok
	}
	return param, ok
}

func (m *Module) MustMapStringInterface(path string) map[string]interface{} {
	param, ok := m.MapStringInterface(path)
	if !ok {
		panic(fmt.Errorf("Missing required parameter in onlineconf or cant parse it %s", path))
	}
	return param
}

// MapIntInt
//
//
func (m *Module) MapIntInt(path string) (map[int]int, bool) {
	param, ok := m.MapIntIntParams[path]
	if ok {
		clone := make(map[int]int, len(param))
		for k, v := range param {
			clone[k] = v
		}
		return clone, ok
	}
	return param, ok
}

// MapIntIntWithDef default valur will not be copied!
func (m *Module) MapIntIntWithDef(path string, defaultValue map[int]int) (map[int]int, bool) {
	param, ok := m.MapIntInt(path)
	if !ok {
		return defaultValue, ok
	}
	return param, ok
}

func (m *Module) MustMapIntInt(path string) map[int]int {
	param, ok := m.MapIntInt(path)
	if !ok {
		panic(fmt.Errorf("Missing required parameter in onlineconf or cant parse it %s", path))
	}
	return param
}

// MapIntString
//
//
func (m *Module) MapIntString(path string) (map[int]string, bool) {
	param, ok := m.MapIntStringParams[path]
	if ok {
		clone := make(map[int]string, len(param))
		for k, v := range param {
			clone[k] = v
		}
		return clone, ok
	}
	return param, ok
}

// MapIntStringWithDef default valur will not be copied!
func (m *Module) MapIntStringWithDef(path string, defaultValue map[int]string) (map[int]string, bool) {
	param, ok := m.MapIntString(path)
	if !ok {
		return defaultValue, ok
	}
	return param, ok
}

func (m *Module) MustMapIntString(path string) map[int]string {
	param, ok := m.MapIntString(path)
	if !ok {
		panic(fmt.Errorf("Missing required parameter in onlineconf or cant parse it %s", path))
	}
	return param
}

// MapStringInt
//
//
func (m *Module) MapStringInt(path string) (map[string]int, bool) {
	param, ok := m.MapStringIntParams[path]
	if ok {
		clone := make(map[string]int, len(param))
		for k, v := range param {
			clone[k] = v
		}
		return clone, ok
	}
	return param, ok
}

// MapStringIntWithDef default valur will not be copied!
func (m *Module) MapStringIntWithDef(path string, defaultValue map[string]int) (map[string]int, bool) {
	param, ok := m.MapStringInt(path)
	if !ok {
		return defaultValue, ok
	}
	return param, ok
}

//
func (m *Module) MustMapStringInt(path string) map[string]int {
	param, ok := m.MapStringInt(path)
	if !ok {
		panic(fmt.Errorf("Missing required parameter in onlineconf or cant parse it %s", path))
	}
	return param
}

// MapStringString
//
//
func (m *Module) MapStringString(path string) (map[string]string, bool) {
	param, ok := m.MapStringStringParams[path]
	if ok {
		clone := make(map[string]string, len(param))
		for k, v := range param {
			clone[k] = v
		}
		return clone, ok
	}
	return param, ok
}

// MapStringStringWithDef default valur will not be copied!
func (m *Module) MapStringStringWithDef(path string, defaultValue map[string]string) (map[string]string, bool) {
	param, ok := m.MapStringString(path)
	if !ok {
		return defaultValue, ok
	}
	return param, ok
}

func (m *Module) MustMapStringString(path string) map[string]string {
	param, ok := m.MapStringString(path)
	if !ok {
		panic(fmt.Errorf("Missing required parameter in onlineconf or cant parse it %s", path))
	}
	return param
}

// RawJSON returns raw json string
func (m *Module) RawJSON(path string) (string, bool) {
	param, ok := m.RawJSONParams[path]
	if ok {
		return param, ok
	}
	return param, ok
}

// RawJSONWithDef default valur will not be copied!
func (m *Module) RawJSONWithDef(path string, defaultValue string) (string, bool) {
	param, ok := m.RawJSON(path)
	if !ok {
		return defaultValue, ok
	}
	return param, ok
}

func (m *Module) MustRawJSON(path string) string {
	param, ok := m.RawJSON(path)
	if !ok {
		panic(fmt.Errorf("Missing required parameter in onlineconf or cant parse it %s", path))
	}
	return param
}
