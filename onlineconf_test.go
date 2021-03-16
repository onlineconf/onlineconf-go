package onlineconf

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"sync"
	"testing"

	"github.com/alldroll/cdb"
	"github.com/stretchr/testify/suite"
)

type testCDBRecord struct {
	key []byte
	val []byte
}
type OCTestSuite struct {
	suite.Suite
	cdbFile   *os.File
	cdbHandle *cdb.CDB

	testRecordsStr []testCDBRecord
	testRecordsInt []testCDBRecord

	module *Module
}

func TestOCTestSuite(t *testing.T) {
	suite.Run(t, new(OCTestSuite))
}

func (suite *OCTestSuite) getCDBReader() cdb.Reader {
	// initialize reader
	reader, err := suite.cdbHandle.GetReader(suite.cdbFile)
	suite.Require().Nilf(err, "Can't get CDB reader: %#v", err)
	return reader
}

func (suite *OCTestSuite) getCDBWriter() cdb.Writer {
	// initialize writer
	writer, err := suite.cdbHandle.GetWriter(suite.cdbFile)
	suite.Require().Nilf(err, "Can't get CDB writer: %#v", err)
	return writer
}

// generate test records
func generateTestRecords(tesRecsCnt int) ([]testCDBRecord, []testCDBRecord) {
	testRecordsStr := make([]testCDBRecord, tesRecsCnt)
	testRecordsInt := make([]testCDBRecord, tesRecsCnt)

	for i := 0; i < tesRecsCnt; i++ {
		stri := strconv.Itoa(i)
		typeByte := "s"
		testRecordsStr[i].key = []byte("/test/onlineconf/str" + stri)
		testRecordsStr[i].val = []byte(typeByte + "val" + stri)

		// log.Printf("key %s val %s", string(testRecordsStr[i].key), string(testRecordsStr[i].val))

		testRecordsInt[i].key = []byte("/test/onlineconf/int" + stri)
		testRecordsInt[i].val = []byte(typeByte + stri)

		// log.Printf("key %s val %s", string(testRecordsInt[i].key), string(testRecordsInt[i].val))

	}
	return testRecordsStr, testRecordsInt
}

func (suite *OCTestSuite) SetupTest() {
	f, err := ioutil.TempFile("", "test_*.cdb")
	suite.Require().Nilf(err, "Can't open temporary file: %#v", err)

	suite.cdbFile = f
	suite.cdbHandle = cdb.New() // create new cdb handle

	testRecordsStr, testRecordsInt := generateTestRecords(2)

	suite.testRecordsStr = testRecordsStr
	suite.testRecordsInt = testRecordsInt

	suite.fillTestCDB()

	suite.module = &Module{name: "testmodule", filename: f.Name()}
	suite.module.reopen()
}

func fillTestCDB(writer cdb.Writer, testRecordsStr, testRecordsInt []testCDBRecord) error {

	allTestRecords := []testCDBRecord{}
	allTestRecords = append(allTestRecords, testRecordsInt...)
	allTestRecords = append(allTestRecords, testRecordsStr...)
	for _, rec := range allTestRecords {
		// log.Printf("putting: key %s val %s", string(rec.key), string(rec.val))
		err := writer.Put(rec.key, rec.val)
		if err != nil {
			return err
		}
	}
	err := writer.Close()
	if err != nil {
		return err
	}
	return nil
}

func (suite *OCTestSuite) fillTestCDB() {

	writer := suite.getCDBWriter()
	err := fillTestCDB(writer, suite.testRecordsStr, suite.testRecordsInt)
	suite.Nil(err)
	suite.Require().Nilf(err, "Cant put new value to cdb: %#v", err)
}

func (suite *OCTestSuite) TestInt() {
	module := suite.module

	for _, testRec := range suite.testRecordsInt {
		ocInt, ok := module.GetIntIfExists(string(testRec.key))
		suite.True(ok, "Cant find key %s in test onlineconf", string(testRec.key))
		testInt, err := strconv.Atoi(string(testRec.val[1:]))
		if err != nil {
			panic(fmt.Errorf("Cant parse test record int: %w", err))
		}
		suite.Equal(ocInt, testInt)

		ocInt, ok = module.GetIntIfExists(string(testRec.key))
		suite.True(ok, "Cant find key %s in test onlineconf", string(testRec.key))
		suite.Equal(ocInt, testInt)
	}

	for _, testRec := range suite.testRecordsInt {
		_, ok := module.GetIntIfExists(string(testRec.key) + "_not_exists")
		suite.False(ok, "Cant find key %s in test onlineconf", string(testRec.key))
	}
}

func (suite *OCTestSuite) TestString() {
	module := suite.module

	for _, testRec := range suite.testRecordsStr {
		ocStr, ok := module.GetStringIfExists(string(testRec.key))
		suite.True(ok, "Cant find key %s in test onlineconf", string(testRec.key))
		suite.Equal(string(testRec.val[1:]), ocStr)

		ocStr, ok = module.GetStringIfExists(string(testRec.key))
		suite.True(ok, "Cant find key %s in test onlineconf", string(testRec.key))
		suite.Equal(string(testRec.val[1:]), ocStr)

	}

	for _, testRec := range suite.testRecordsStr {
		_, ok := module.GetStringIfExists(string(testRec.key) + "_not_exists")
		suite.False(ok, "Cant find key %s in test onlineconf", string(testRec.key))
	}
}

func (suite *OCTestSuite) TestReload() {
	// todo
}

func (suite *OCTestSuite) TestConcurrent() {

	module := suite.module

	wg := &sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 10; i++ {
				for _, rec := range suite.testRecordsInt {
					value, ok := module.GetIntIfExists(string(rec.key))
					suite.True(ok, "Test recor existst")
					testInt, err := strconv.Atoi(string(rec.val[1:]))
					if err != nil {
						panic(fmt.Errorf("Cant parse test record int: %w", err))
					}
					suite.Equal(value, testInt)
				}
			}
		}()
	}

	wg.Wait()
}
