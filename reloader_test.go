package onlineconf

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"testing"
	"time"

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

	testRecordsStr, testRecordsInt, testRecordsBool := generateTestRecords(2)

	allTestRecords := []testCDBRecord{}
	allTestRecords = append(allTestRecords, testRecordsInt...)
	allTestRecords = append(allTestRecords, testRecordsStr...)
	allTestRecords = append(allTestRecords, testRecordsBool...)

	err = fillTestCDB(writer, allTestRecords)
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
		_ = mr.Reload()
	}

	t.StopTimer()
}

func (suite *OCTestSuite) TestReload() {
	file, err := ioutil.TempFile("", "test_reloader_*.cdb")
	suite.Require().Nilf(err, "Can't open temporary file: %#v", err)

	// generate test data
	typeByte := "s"
	testPath := "/test/onlineconf/reloader_test"
	testValue := "original_value"

	writer, err := cdb.Create(file.Name())
	suite.Require().NoError(err, "Can't get CDB writer: %#v", err)
	err = fillTestCDB(writer, []testCDBRecord{{key: []byte(testPath), val: []byte(typeByte + testValue)}})
	suite.Require().NoError(err)

	// create reloader
	mr, err := NewModuleReloader(&ReloaderOptions{FilePath: file.Name()})
	suite.Require().NoError(err)

	module := mr.Module()
	val, err := module.String(MustConfigParamString(testPath, ""))
	suite.Assert().Equal(val, testValue)
	suite.Assert().NoError(err)

	// rewrite cdb data with updated key
	newTestValue := "updated_cdb_value"
	writer, err = cdb.Create(file.Name())
	suite.Require().Nilf(err, "Can't get CDB writer: %#v", err)
	err = fillTestCDB(writer, []testCDBRecord{{key: []byte(testPath), val: []byte(typeByte + newTestValue)}})
	suite.Require().NoError(err)

	err = mr.Reload()
	suite.Require().NoError(err)
	newModule := mr.Module()

	// old module instance returns old value
	val, err = module.String(MustConfigParamString(testPath, ""))
	suite.Assert().Equal(val, testValue)
	suite.Assert().NoError(err)

	val, err = newModule.String(MustConfigParamString(testPath, ""))
	suite.Assert().Equal(newTestValue, val)
	suite.Assert().NoError(err)
}

// Simple reloader with no error handling.
func ExampleModuleReloader_RunWatcher_simple() {
	reloader := MustReloader("TREE")
	go reloader.RunWatcher(context.TODO())

	oldModule := reloader.Module()

	// let `/test/onlineconf/int0` == 123
	oldModule.Int(MustConfigParamInt("/test/onlineconf/int0", 0))

	// time.Sleep(time.Minute)

	// now reloader updates let `/test/onlineconf/int0` == 12345
	// After reloader reloads config new module must be created
	updatedModule := reloader.Module()

	// oldModule value is still 123
	oldModule.Int(MustConfigParamInt("/test/onlineconf/int0", 0))

	// updatedModule value is 12345
	updatedModule.Int(MustConfigParamInt("/test/onlineconf/int0", 0))
}

// More sophisticated way with reload retries, error logging and graceful shutdown.
func ExampleModuleReloader_RunWatcher_with_retries() {
	reloader := MustReloader("TREE")
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		for {
			err := reloader.RunWatcher(ctx)
			if err != nil {
				fmt.Fprintf(os.Stderr, "onlineconf reloader error: %s", err.Error())
				time.Sleep(time.Second)
				continue
			}
			break
		}

		wg.Done()
	}()

	module := reloader.Module()
	// retrieve config params from module
	module.Int(MustConfigParamInt("/test/onlineconf/int0", 0))

	// ...

	cancel()
	wg.Wait()

}
