package onlineconf

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testCacheStruct struct {
	a int
	b string
}

type testAnotherCacheStruct struct {
	c int
	d bool
}

func TestValueCache(t *testing.T) {
	tests := []struct {
		path     string
		value    interface{}
		newValue interface{}
	}{
		{
			"/path/to/int32",
			int32(123123),
			int32(0),
		},
		{
			"/path/to/struct",
			testCacheStruct{a: 123, b: "test"},
			testCacheStruct{},
		},
		{
			"/path/to/struct",
			testAnotherCacheStruct{c: -1, d: true},
			testAnotherCacheStruct{},
		},
	}

	var cache valueCache

	cache.init()

	for _, tt := range tests {
		cache.set(tt.path, reflect.ValueOf(tt.value))
	}

	for _, tt := range tests {
		val := reflect.New(reflect.TypeOf(tt.newValue)).Elem()

		if !cache.get(tt.path, val) {
			t.Errorf("%s: get failed", tt.path)
		}

		assert.Equal(t, tt.value, val.Interface(), "missing value for %s", tt.path)
	}
}

func TestCacheImmutability(t *testing.T) {
	var (
		cache     valueCache
		fromCache testCacheStruct
	)

	cache.init()

	orig := testCacheStruct{a: 123, b: "test"}

	cache.set("/path/to", reflect.ValueOf(orig))
	cache.get("/path/to", reflect.ValueOf(&fromCache).Elem())
	assert.Equal(t, orig, fromCache, "cached value differs from original")

	fromCache.b = "foobar"

	cache.get("/path/to", reflect.ValueOf(&fromCache).Elem())
	assert.Equal(t, orig, fromCache, "cached value was modified")
}

func TestSyncCache(t *testing.T) {
	sc := syncCache[string]{}
	key := "test"
	want := "foobar"

	got, ch, ok := sc.load(key)
	if ok {
		t.Fatal("cache isn't empty on start")
	}

	if got != "" {
		t.Fatalf(`first load(): got %q, want ""`, got)
	}

	if ch == nil {
		t.Fatal("first load(): ch must not be nil")
	}

	started := make(chan struct{})
	done := make(chan struct{})

	go func() {
		close(started) // doesn't strictly guarantee that the second load() will lock

		got, ch, ok := sc.load(key)
		if !ok {
			t.Error("second load() isn't ok")
		}

		if ch != nil {
			t.Error("second load(): ch must be nil")
		}

		if got != want {
			t.Errorf("second load() = %q, want %q", got, want)
		}

		close(done)
	}()

	<-started
	sc.store(key, ch, want)
	<-done
}
