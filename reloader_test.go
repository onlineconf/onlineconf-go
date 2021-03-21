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

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		for {
			err := mr.RunWatcher(ctx)
			if err != nil {
				fmt.Fprintf(os.Stderr, "onlineconf reloader error: %s", err.Error())
				continue
			}
			break
		}

		wg.Done()
	}()

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

	err = os.Chmod(file.Name(), 0644)
	suite.Require().NoError(err)

	var newModule *Module

	maxTries := 20
	for i := 0; i < maxTries; i++ {
		newModule = mr.Module()
		if newModule != module {
			break
		}
		limitNotReached := suite.Assert().Less(i, maxTries-1, "max tries limit reached")
		if limitNotReached {
			time.Sleep(time.Second)
		} else {
			// force reload is inotiify watcher
			mr.Reload()
		}
	}

	// old module instance returns old value
	val, err = module.String(MustConfigParamString(testPath, ""))
	suite.Assert().Equal(val, testValue)
	suite.Assert().NoError(err)

	val, err = newModule.String(MustConfigParamString(testPath, ""))
	suite.Assert().Equal(newTestValue, val)
	suite.Assert().NoError(err)

	cancel()
	wg.Wait()
}
