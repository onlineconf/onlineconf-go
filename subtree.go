package onlineconf

import (
	"path"
	"time"
)

// Subtree represents a subset of a [Module] containing all parameters with a specified prefix.
// [Module.Subtree] and [Subtree.Subtree] methods use [path.Clean] and drop a trailing slash
// in a prefix. Get* methods construct paths using simple string concatenation and require
// a leading slash.
//
// Path separators other than a slash aren't supported by this type.
type Subtree struct {
	mod    *Module
	prefix string
}

// Path returns a full path in a module using the subtree prefix.
func (s *Subtree) Path(path string) string {
	return s.prefix + path
}

// GetStringErr calls [Module.GetStringErr] using the subtree prefix.
func (s *Subtree) GetStringErr(path string) (string, error) {
	return s.mod.GetStringErr(s.prefix + path)
}

// GetStringIfExists calls [Module.GetStringIfExists] using the subtree prefix.
func (s *Subtree) GetStringIfExists(path string) (string, bool) {
	return s.mod.GetStringIfExists(s.prefix + path)
}

// GetString calls [Module.GetString] using the subtree prefix.
func (s *Subtree) GetString(path string, dfl string) string {
	return s.mod.GetString(s.prefix+path, dfl)
}

// GetIntErr calls [Module.GetIntErr] using the subtree prefix.
func (s *Subtree) GetIntErr(path string) (int, error) {
	return s.mod.GetIntErr(s.prefix + path)
}

// GetIntIfExists calls [Module.GetIntIfExists] using the subtree prefix.
func (s *Subtree) GetIntIfExists(path string) (int, bool) {
	return s.mod.GetIntIfExists(s.prefix + path)
}

// GetInt calls [Module.GetInt] using the subtree prefix.
func (s *Subtree) GetInt(path string, dfl int) int {
	return s.mod.GetInt(s.prefix+path, dfl)
}

// GetBoolErr calls [Module.GetBoolErr] using the subtree prefix.
func (s *Subtree) GetBoolErr(path string) (bool, error) {
	return s.mod.GetBoolErr(s.prefix + path)
}

// GetBoolIfExists calls [Module.GetBoolIfExists] using the subtree prefix.
func (s *Subtree) GetBoolIfExists(path string) (bool, bool) {
	return s.mod.GetBoolIfExists(s.prefix + path)
}

// GetBool calls [Module.GetBool] using the subtree prefix.
func (s *Subtree) GetBool(path string, dfl bool) bool {
	return s.mod.GetBool(s.prefix+path, dfl)
}

// GetDurationErr calls [Module.GetDurationErr] using the subtree prefix.
func (s *Subtree) GetDurationErr(path string) (time.Duration, error) {
	return s.mod.GetDurationErr(s.prefix + path)
}

// GetDurationIsExists calls [Module.GetDurationIsExists] using the subtree prefix.
func (s *Subtree) GetDurationIsExists(path string) (time.Duration, bool) {
	return s.mod.GetDurationIsExists(s.prefix + path)
}

// GetDuration calls [Module.GetDuration] using the subtree prefix.
func (s *Subtree) GetDuration(path string, dfl time.Duration) time.Duration {
	return s.mod.GetDuration(s.prefix+path, dfl)
}

// GetFloatErr calls [Module.GetFloatErr] using the subtree prefix.
func (s *Subtree) GetFloatErr(path string) (float64, error) {
	return s.mod.GetFloatErr(s.prefix + path)
}

// GetFloatIfExists calls [Module.GetFloatIfExists] using the subtree prefix.
func (s *Subtree) GetFloatIfExists(path string) (float64, bool) {
	return s.mod.GetFloatIfExists(s.prefix + path)
}

// GetFloat calls [Module.GetFloat] using the subtree prefix.
func (s *Subtree) GetFloat(path string, dfl float64) float64 {
	return s.mod.GetFloat(s.prefix+path, dfl)
}

// GetStrings calls [Module.GetStrings] using the subtree prefix.
func (s *Subtree) GetStrings(path string, dfl []string) []string {
	return s.mod.GetStrings(s.prefix+path, dfl)
}

// GetStruct calls [Module.GetStruct] using the subtree prefix.
func (s *Subtree) GetStruct(path string, valuePtr interface{}) (bool, error) {
	return s.mod.GetStruct(s.prefix+path, valuePtr)
}

// SubscribeChan calls [Module.SubscribeChan] using the subtree prefix.
func (s *Subtree) SubscribeChan(path string, ch chan<- struct{}) error {
	return s.mod.SubscribeChan(s.prefix+path, ch)
}

// Subscribe calls [Module.Subscribe] using the subtree prefix.
func (s *Subtree) Subscribe(path string) (chan struct{}, error) {
	return s.mod.Subscribe(s.prefix + path)
}

// SubscribeChanSubtree calls [Module.SubscribeChanSubtree] using the subtree prefix.
func (s *Subtree) SubscribeChanSubtree(path string, ch chan<- struct{}) error {
	return s.mod.SubscribeChanSubtree(s.prefix+path, ch)
}

// SubscribeSubtree calls [Module.SubscribeSubtree] using the subtree prefix.
func (s *Subtree) SubscribeSubtree(path string) (chan struct{}, error) {
	return s.mod.SubscribeSubtree(s.prefix + path)
}

// UnsubscribeChan calls [Module.UnsubscribeChan] using the subtree prefix.
func (s *Subtree) UnsubscribeChan(path string, ch chan<- struct{}) {
	s.mod.UnsubscribeChan(s.prefix+path, ch)
}

// UnsubscribeChanSubtree calls [Module.UnsubscribeChanSubtree] using the subtree prefix.
func (s *Subtree) UnsubscribeChanSubtree(path string, ch chan<- struct{}) {
	s.mod.UnsubscribeChanSubtree(s.prefix+path, ch)
}

// Unsubscribe calls [Module.Unsubscribe] using the subtree prefix.
func (s *Subtree) Unsubscribe(path string) {
	s.mod.Unsubscribe(s.prefix + path)
}

// UnsubscribeSubtree calls [Module.UnsubscribeSubtree] using the subtree prefix.
func (s *Subtree) UnsubscribeSubtree(path string) {
	s.mod.UnsubscribeSubtree(s.prefix + path)
}

// Subtree returns a subtree of a subtree. Prefixes are concatenated using [path.Join].
func (s *Subtree) Subtree(prefix string) *Subtree {
	return &Subtree{
		mod:    s.mod,
		prefix: cleanPrefix(path.Join(s.prefix, prefix)),
	}
}

func cleanPrefix(prefix string) string {
	if prefix == "/" || prefix == "." {
		return ""
	}

	return prefix
}

func cleanPath(p string) string {
	p = path.Clean(p)
	if p == "." {
		return "/"
	}

	return p
}

// OpenSubtree is a helper function that calls [OpenModule] and [Module.Subtree].
func OpenSubtree(moduleName, prefix string) (*Subtree, error) {
	mod, err := OpenModule(moduleName)
	if err != nil {
		return nil, err
	}

	return mod.Subtree(prefix), nil
}
