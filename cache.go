package onlineconf

import (
	"reflect"
	"sync"
)

// valueCache stores deserialized representations for every distinct type used when getting values of the path.
type valueCache struct {
	sync.RWMutex
	cache map[string][]reflect.Value // map[string]map[reflect.Type] should be inefficient when very few (usually 1) types are cached
}

func (cache *valueCache) init() {
	cache.Lock()
	defer cache.Unlock()

	cache.cache = make(map[string][]reflect.Value)
}

func (cache *valueCache) get(path string, val reflect.Value) bool {
	cache.RLock()
	defer cache.RUnlock()

	typ := val.Type()
	for _, cached := range cache.cache[path] {
		if cached.Type() == typ {
			val.Set(cached)

			return true
		}
	}

	return false
}

func (cache *valueCache) set(path string, val reflect.Value) {
	cache.Lock()
	defer cache.Unlock()

	typ := val.Type()
	values := cache.cache[path]

	for i, cached := range values {
		if cached.Type() == typ {
			values[i].Set(val)
			return
		}
	}

	cache.cache[path] = append(values, reflect.ValueOf(val.Interface())) // store a shallow copy in the cache
}

// syncCache is a sync.Map with cache stampede protection.
// cache invalidation isn't implemented.
type syncCache[T any] struct {
	m sync.Map
}

func (sc *syncCache[T]) load(key any) (T, chan<- struct{}, bool) {
	ch := make(chan struct{})

	value, loaded := sc.m.LoadOrStore(key, ch)
	if !loaded { // no value and no pending operation
		var zero T
		return zero, ch, false
	}

	if cached, ok := value.(T); ok {
		return cached, nil, true
	}

	<-value.(chan struct{})
	value, _ = sc.m.Load(key)

	return value.(T), nil, true
}

func (sc *syncCache[T]) loadOnly(key any) (T, bool) {
	value, ok := sc.m.Load(key)
	if !ok {
		var zero T
		return zero, false
	}

	if cached, ok := value.(T); ok {
		return cached, true
	}

	<-value.(chan struct{})
	value, _ = sc.m.Load(key)

	return value.(T), true
}

func (sc *syncCache[T]) store(key any, ch chan<- struct{}, value T) {
	sc.m.Store(key, value)
	close(ch)
}
