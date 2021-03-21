package onlineconf

import (
	"io/ioutil"
	"os"
	"strconv"
	"testing"

	"github.com/colinmarc/cdb"
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

	testRecordsStr  []testCDBRecord
	testRecordsInt  []testCDBRecord
	testRecordsBool []testCDBRecord

	mr *ModuleReloader
}

func TestOCTestSuite(t *testing.T) {
	suite.Run(t, new(OCTestSuite))
}

func (suite *OCTestSuite) getCDBReader() *cdb.CDB {
	return suite.cdbHandle
}

func (suite *OCTestSuite) getCDBWriter() *cdb.Writer {
	// initialize writer
	writer, err := cdb.Create(suite.cdbFile.Name())
	suite.Require().Nilf(err, "Can't get CDB writer: %#v", err)
	return writer
}

const TestOnlineConfTrue = "/test/onlineconf/bool_true"
const TestOnlineConfFalse = "/test/onlineconf/bool_false"
const TestOnlineConfEmpty = "/test/onlineconf/bool_empty"

// generate test records
func generateTestRecords(tesRecsCnt int) ([]testCDBRecord, []testCDBRecord, []testCDBRecord) {
	testRecordsStr := make([]testCDBRecord, tesRecsCnt)
	testRecordsInt := make([]testCDBRecord, tesRecsCnt)

	typeByte := "s"

	for i := 0; i < tesRecsCnt; i++ {
		stri := strconv.Itoa(i)
		testRecordsStr[i].key = []byte("/test/onlineconf/str" + stri)
		testRecordsStr[i].val = []byte(typeByte + "val" + stri)

		testRecordsInt[i].key = []byte("/test/onlineconf/int" + stri)
		testRecordsInt[i].val = []byte(typeByte + stri)

	}

	testRecordsBool := make([]testCDBRecord, 3)
	testRecordsBool[0].key = []byte(TestOnlineConfTrue)
	testRecordsBool[0].val = []byte(typeByte + "1")
	testRecordsBool[1].key = []byte(TestOnlineConfFalse)
	testRecordsBool[1].val = []byte(typeByte + "0")
	testRecordsBool[2].key = []byte(TestOnlineConfEmpty)
	testRecordsBool[2].val = []byte(typeByte + "")

	return testRecordsStr, testRecordsInt, testRecordsBool
}

func (suite *OCTestSuite) SetupTest() {
	f, err := ioutil.TempFile("", "test_*.cdb")
	suite.Require().Nilf(err, "Can't open temporary file: %#v", err)

	suite.T().Logf("setd cdb: %s\n", f.Name())

	suite.cdbFile = f
	suite.cdbHandle = suite.getCDBReader()

	suite.prepareTestData()

	mr, err := NewModuleReloader(&ReloaderOptions{FilePath: f.Name()})
	suite.Nilf(err, "Cant init onlineconf module!: %#v", err)
	suite.mr = mr
}

func (suite *OCTestSuite) TearDownTest() {
	err := suite.cdbFile.Close()
	suite.Nilf(err, "Can't close cdb file: %#v", err)

	err = os.Remove(suite.cdbFile.Name())
	suite.Nilf(err, "Can't remove cdb file: %#v", err)
}

func fillTestCDB(writer *cdb.Writer, testRecords []testCDBRecord) error {

	for _, rec := range testRecords {
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

func (suite *OCTestSuite) prepareTestData() {

	testRecordsStr, testRecordsInt, testRecordsBool := generateTestRecords(2)

	suite.testRecordsStr = testRecordsStr
	suite.testRecordsInt = testRecordsInt
	suite.testRecordsBool = testRecordsBool

	allTestRecords := []testCDBRecord{}
	allTestRecords = append(allTestRecords, testRecordsInt...)
	allTestRecords = append(allTestRecords, testRecordsStr...)
	allTestRecords = append(allTestRecords, testRecordsBool...)

	writer := suite.getCDBWriter()
	err := fillTestCDB(writer, allTestRecords)
	suite.Nil(err)
	suite.Require().Nilf(err, "Cant put new value to cdb: %#v", err)
}

func (suite *OCTestSuite) TestInt() {
	module := suite.mr.Module()

	for _, testRec := range suite.testRecordsInt {
		intParam := MustConfigParamInt(string(testRec.key), 0)
		ocInt, err := module.Int(intParam)
		suite.NoErrorf(err, "Cant find key %s in test onlineconf", string(testRec.key))
		testInt, err := strconv.Atoi(string(testRec.val[1:]))
		suite.Require().NoErrorf(err, "Cant parse test record int: %w", err)
		suite.Equal(ocInt, testInt)

		// cached value expected to be returned
		ocInt, err = module.Int(intParam)
		suite.NoErrorf(err, "Cant find key %s in test onlineconf", string(testRec.key))
		suite.Equal(ocInt, testInt)
	}

	for i, testRec := range suite.testRecordsInt {
		intParam := MustConfigParamInt(string(testRec.key)+"_not_exists", i)
		ocInt, err := module.Int(intParam)
		suite.True(IsErrKeyNotFound(err), "non existing path: %s", string(testRec.key))
		suite.Equal(ocInt, i, "Default result was returned")

		// cached value expected to be returned
		ocInt, err = module.Int(intParam)
		suite.True(IsErrKeyNotFound(err), "non existing path: %s", string(testRec.key))
		suite.Equal(ocInt, i, "Default result was returned")
	}
}

func (suite *OCTestSuite) TestString() {
	module := suite.mr.Module()

	for _, testRec := range suite.testRecordsStr {
		strPath := MustConfigParamString(string(testRec.key), "")
		ocStr, err := module.String(strPath)
		suite.NoErrorf(err, "Cant find key %s in test onlineconf", string(testRec.key))
		suite.Equal(string(testRec.val[1:]), ocStr)

		// cached value expected to be returned
		ocStr, err = module.String(strPath)
		suite.NoErrorf(err, "Cant find key %s in test onlineconf", string(testRec.key))
		suite.Equal(string(testRec.val[1:]), ocStr)
	}

	for i, testRec := range suite.testRecordsStr {
		defaultParamValue := "test_not_exists_" + strconv.Itoa(i)
		strParam := MustConfigParamString(string(testRec.key)+"_not_exists", defaultParamValue)
		ocStr, err := module.String(strParam)
		suite.True(IsErrKeyNotFound(err), "non existing path: %s", string(testRec.key))
		suite.Equal(ocStr, defaultParamValue, "Default result was returned")

		// cached value expected to be returned
		ocStr, err = module.String(strParam)
		suite.True(IsErrKeyNotFound(err), "non existing path: %s", string(testRec.key))
		suite.Equal(ocStr, defaultParamValue, "Default result was returned")
	}
}

func (suite *OCTestSuite) TestBool() {
	module := suite.mr.Module()

	// TestOnlineConfTrue
	trueVal, err := module.Bool(MustConfigParamBool(TestOnlineConfTrue, false))
	suite.Assert().NoError(err)
	suite.Assert().True(trueVal)

	// cached value expected to be returned
	trueVal, err = module.Bool(MustConfigParamBool(TestOnlineConfTrue, false))
	suite.Assert().NoError(err)
	suite.Assert().True(trueVal)

	// TestOnlineConfFalse
	falseVal, err := module.Bool(MustConfigParamBool(TestOnlineConfFalse, true))
	suite.Assert().NoError(err)
	suite.Assert().False(falseVal)

	// cached value expected to be returned
	falseVal, err = module.Bool(MustConfigParamBool(TestOnlineConfTrue, true))
	suite.Assert().NoError(err)
	suite.Assert().True(falseVal)

	// TestOnlineConfEmpty
	falseVal, err = module.Bool(MustConfigParamBool(TestOnlineConfEmpty, true))
	suite.Assert().NoError(err)
	suite.Assert().False(falseVal)

	// cached value expected to be returned
	falseVal, err = module.Bool(MustConfigParamBool(TestOnlineConfEmpty, true))
	suite.Assert().NoError(err)
	suite.Assert().False(falseVal)
}

func (suite *OCTestSuite) TestReload() {
	// todo
}

func (suite *OCTestSuite) TestUnknownParamType() {
	// todo
}

func (suite *OCTestSuite) TestJSON() {
	// todo
}

func (suite *OCTestSuite) TestConcurrent() {
	// todo
}
