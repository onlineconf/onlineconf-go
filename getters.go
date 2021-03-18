package onlineconf

import "fmt"

// String returns value of a named parameter from the module.
// It returns the boolean true if the parameter exists and is a string.
// In the other case it returns the boolean false and an empty string.
func (m *Module) String(path string) (string, bool) {
	param, ok := m.stringParams[path]
	return param, ok
}

// StringWithDef returns value of a named parameter from the module.
// It returns the boolean true if the parameter exists and is a string.
// In the other case it returns the boolean false and an empty string.
func (m *Module) StringWithDef(path string, defaultValue string) (string, bool) {
	param, ok := m.stringParams[path]
	if !ok {
		return defaultValue, ok
	}
	return param, ok
}


// Int returns value of a named parameter from the module.
// It returns the boolean true if the parameter exists and is an int.
// In the other case it returns the boolean false and zero.
func (m *Module) Int(path string) (int, bool) {
	param, ok := m.intParams[path]
	return param, ok
}

// IntWithDef returns value of a named parameter from the module.
// It returns the boolean true if the parameter exists and is an int.
// In the other case it returns the boolean false and zero.
func (m *Module) IntWithDef(path string, defaultValue int) (int, bool) {
	param, ok := m.intParams[path]
	if !ok {
		return defaultValue, ok
	}
	return param, ok
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
