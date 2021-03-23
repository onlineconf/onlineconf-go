package onlineconf

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"testing"

	"github.com/colinmarc/cdb"
	"github.com/stretchr/testify/suite"
	"golang.org/x/exp/mmap"
)

type testCDBRecord struct {
	key []byte
	val []byte
}
type OCTestSuite struct {
	suite.Suite
	cdbFile *os.File

	testRecordsStr  []testCDBRecord
	testRecordsInt  []testCDBRecord
	testRecordsBool []testCDBRecord

	mr *ModuleReloader
}

func TestOCTestSuite(t *testing.T) {
	suite.Run(t, new(OCTestSuite))
}

func (suite *OCTestSuite) getCDBWriter(file *os.File) *cdb.Writer {
	// initialize writer
	writer, err := cdb.Create(file.Name())
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

	suite.cdbFile = f

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

	writer := suite.getCDBWriter(suite.cdbFile)
	err := fillTestCDB(writer, allTestRecords)
	suite.Require().NoError(err, "Cant put new value to cdb: %#v", err)
}

func (suite *OCTestSuite) TestInt() {
	module := suite.mr.Module()

	for _, testRec := range suite.testRecordsInt {
		intParam := MustConfigParamInt(string(testRec.key), 0)
		ocInt := module.Int(intParam)
		testInt, err := strconv.Atoi(string(testRec.val[1:]))
		suite.Require().NoErrorf(err, "Cant parse test record int: %w", err)
		suite.Equal(ocInt, testInt)

		// cached value expected to be returned
		ocInt = module.Int(intParam)
		suite.Equal(ocInt, testInt)
	}

	for i, testRec := range suite.testRecordsInt {
		intParam := MustConfigParamInt(string(testRec.key)+"_not_exists", i)
		ocInt := module.Int(intParam)
		suite.Equal(ocInt, i, "Default result was returned")

		// cached value expected to be returned
		ocInt = module.Int(intParam)
		suite.Equal(ocInt, i, "Default result was returned")
	}
}

func (suite *OCTestSuite) TestString() {
	module := suite.mr.Module()

	for _, testRec := range suite.testRecordsStr {
		strPath := MustConfigParamString(string(testRec.key), "")
		ocStr := module.String(strPath)
		suite.Equal(string(testRec.val[1:]), ocStr)

		// cached value expected to be returned
		ocStr = module.String(strPath)
		suite.Equal(string(testRec.val[1:]), ocStr)
	}

	for i, testRec := range suite.testRecordsStr {
		defaultParamValue := "test_not_exists_" + strconv.Itoa(i)
		strParam := MustConfigParamString(string(testRec.key)+"_not_exists", defaultParamValue)
		ocStr := module.String(strParam)
		suite.Equal(ocStr, defaultParamValue, "Default result was returned")

		// cached value expected to be returned
		ocStr = module.String(strParam)
		suite.Equal(ocStr, defaultParamValue, "Default result was returned")
	}
}

func (suite *OCTestSuite) TestBool() {
	module := suite.mr.Module()

	// TestOnlineConfTrue
	trueVal := module.Bool(MustConfigParamBool(TestOnlineConfTrue, false))
	suite.Assert().True(trueVal)

	// cached value expected to be returned
	trueVal = module.Bool(MustConfigParamBool(TestOnlineConfTrue, false))
	suite.Assert().True(trueVal)

	// TestOnlineConfFalse
	falseVal := module.Bool(MustConfigParamBool(TestOnlineConfFalse, true))
	suite.Assert().False(falseVal)

	// cached value expected to be returned
	falseVal = module.Bool(MustConfigParamBool(TestOnlineConfTrue, true))
	suite.Assert().True(falseVal)

	// TestOnlineConfEmpty
	falseVal = module.Bool(MustConfigParamBool(TestOnlineConfEmpty, true))
	suite.Assert().False(falseVal)

	// cached value expected to be returned
	falseVal = module.Bool(MustConfigParamBool(TestOnlineConfEmpty, true))
	suite.Assert().False(falseVal)
}

func (suite *OCTestSuite) TestJSON() {
	// todo
}

func (suite *OCTestSuite) TestConcurrent() {
	// todo
}

func BenchmarkMmappedCdb(t *testing.B) {
	f, err := ioutil.TempFile("", "bench_*.cdb")
	if err != nil {
		panic(err)
	}
	f.Close()

	defer os.Remove(f.Name())

	scdWriter, err := cdb.Create(f.Name())
	if err != nil {
		panic(err)
	}

	testKey := "test_key"
	testVal := "test_val"
	fillTestCDB(scdWriter, []testCDBRecord{{key: []byte(testKey), val: []byte(testVal)}})

	cdbFile, err := mmap.Open(f.Name())
	if err != nil {
		panic(err)
	}

	cdbReader, err := cdb.New(cdbFile, nil)
	if err != nil {
		panic(err)
	}

	var gotString string

	t.ResetTimer()

	for i := 0; i < t.N; i++ {
		gotGytes, err := cdbReader.Get([]byte(testKey))
		if err != nil {
			panic(err)
		}
		gotString = string(gotGytes)
		if gotString != testVal {
			panic("cdb returned invalid string")
		}
	}

	_ = gotString

	t.StopTimer()
}

func BenchmarkGoMap(t *testing.B) {

	testKey := "test_key"
	testVal := "test_val"

	goMap := make(map[string]string)

	goMap[testKey] = testVal

	var gotString string

	t.ResetTimer()

	for i := 0; i < t.N; i++ {
		mapString, _ := goMap[testKey]
		if testVal != mapString {
			panic(fmt.Sprintf("map returned invalid val: %s %s", testVal, mapString))
		}
	}

	_ = gotString

	t.StopTimer()
}
