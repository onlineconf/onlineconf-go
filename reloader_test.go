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

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		for {
			err := suite.mr.RunWatcher(ctx)
			if err != nil {
				fmt.Fprintf(os.Stderr, "onlineconf reloader error: %s", err.Error())
				continue
			}
			break
		}

		wg.Done()
	}()

	testPath := string(suite.testRecordsStr[0].key)
	testValue := string(suite.testRecordsStr[0].val[1:])

	module := suite.mr.Module()
	val, err := module.String(MustConfigParamString(testPath, ""))
	suite.Assert().Equal(val, testValue)
	suite.Assert().NoError(err)

	typeByte := "s"
	newTestValue := "updated_ccb_value"
	suite.testRecordsStr[0].val = []byte(typeByte + newTestValue)

	// rewrite cdb data with updated key
	writer := suite.getCDBWriter()
	err = fillTestCDB(writer, suite.testRecordsStr)
	suite.Require().NoError(err)
	err = os.Chmod(suite.cdbFile.Name(), 0644)
	suite.Require().NoError(err)

	var newModule *Module

	maxTries := 10
	for i := 0; i < maxTries; i++ {
		newModule = suite.mr.Module()
		if newModule != module {
			break
		}
		suite.Require().Less(i, maxTries-1, "max tries limit reached")
		time.Sleep(time.Second)
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
