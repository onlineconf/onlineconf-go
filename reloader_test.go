package onlineconf

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/colinmarc/cdb"
	"github.com/stretchr/testify/assert"
)

func TestInvalidReloader(t *testing.T) {
	assert := assert.New(t)

	_, err := NewModuleReloader(&ReloaderOptions{Name: "NoSuchModule"})
	// log.Printf("reloader err: %#v", err)
	assert.NotNil(err)
}

func createCDB() (*os.File, error) {
	f, err := ioutil.TempFile("", "test_*.cdb")
	if err != nil {
		return nil, err
	}
	writer, err := cdb.NewWriter(f, nil)
	if err != nil {
		return nil, err
	}

	testRecordsStr, testRecordsInt := generateTestRecords(100000)
	err = fillTestCDB(writer, testRecordsStr, testRecordsInt)
	return f, err
}

func BenchmarkModuleReload(t *testing.B) {
	f, err := createCDB()
	if err != nil {
		panic(err)
	}

	mr, err := NewModuleReloader(&ReloaderOptions{FilePath: f.Name()})
	if err != nil {
		panic(err)
	}

	t.ResetTimer()

	for i := 0; i < t.N; i++ {
		mr.Reload()
	}

	t.StopTimer()

	return
}
