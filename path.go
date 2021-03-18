package onlineconf

import (
	"errors"
	"strings"
)

type ParamPath struct {
	path string
}

var ErrNoLeadingSlash = errors.New("onlineconf path must start with `/`")
var ErrTailSlash = errors.New("onlineconf path must not end with `/`")
var ErrDoubleSlash = errors.New("onlineconf path must not contain doubled `/`")

// IsValid checks if onlinecong path is valid.
// Onlineconf path can't end with `/`.
// Onlineconf path must be prefixed with `/`.
// Single slash (`/`) is a valid onlineconf path.
func IsValidParamPath(path string) error {
	if path == "/" {
		return nil
	}

	if !strings.HasPrefix(path, "/") {
		return ErrNoLeadingSlash
	}

	if strings.HasSuffix(path, "/") {
		return ErrTailSlash
	}

	if strings.Contains(path, "//") {
		return ErrDoubleSlash
	}

	return nil
}

// IsValid checks is path is valid.
// The same as IsValidParamPath but as structure method.
// Always returns nil for ParamPaths that were created with NewParamPath.
func (p *ParamPath) IsValid() error {
	return IsValidParamPath(p.String())
}

// String returns path string
func (p *ParamPath) String() string {
	return p.path
}

// IsRoot returns true if path is '/'
func (p *ParamPath) IsRoot() bool {
	return p.path == "/"
}

// Join creates new path that is join of current path and other one.
func (p *ParamPath) Join(otherPath *ParamPath) *ParamPath {
	if p.IsRoot() {
		return &ParamPath{
			path: otherPath.path,
		}
	}

	if otherPath.IsRoot() {
		return &ParamPath{
			path: p.path,
		}
	}

	return &ParamPath{
		path: p.path + otherPath.path,
	}
}

// NewParamPath creates new ParamsPath struct for valid onlineconf paths.
// Error returned for invalid paths.
func NewParamPath(path string) (*ParamPath, error) {
	if err := IsValidParamPath(path); err != nil {
		return nil, err
	}

	return &ParamPath{path: path}, nil
}

// MustParamPath creates new ParamsPath struct for valid onlineconf paths.
// Panics for invalid paths.
func MustParamPath(path string) *ParamPath {
	paramPath, err := NewParamPath(path)
	if err != nil {
		panic(err)
	}

	return paramPath
}
