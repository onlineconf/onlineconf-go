package onlineconf

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
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

	mr *ModuleReloader
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

	fmt.Printf("setd cdb: %s\n", f.Name())

	suite.cdbFile = f
	suite.cdbHandle = cdb.New() // create new cdb handle

	suite.prepareTestData()

	mr, err := NewModuleReloader(&ReloaderOptions{FilePath: f.Name()})
	suite.Nilf(err, "Cant init onlineconf module!: %#v", err)
	suite.mr = mr
}

func (suite *OCTestSuite) TearDownTest() {
	err := suite.mr.Close()
	suite.Nilf(err, "Can't close module reloader: %#v", err)

	err = suite.cdbFile.Close()
	suite.Nilf(err, "Can't close cdb file: %#v", err)

	// err = os.Remove(suite.cdbFile.Name())
	suite.Nilf(err, "Can't remove cdb file: %#v", err)
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

func (suite *OCTestSuite) prepareTestData() {

	testRecordsStr, testRecordsInt := generateTestRecords(2)

	suite.testRecordsStr = testRecordsStr
	suite.testRecordsInt = testRecordsInt

	writer := suite.getCDBWriter()
	err := fillTestCDB(writer, suite.testRecordsStr, suite.testRecordsInt)
	suite.Nil(err)
	suite.Require().Nilf(err, "Cant put new value to cdb: %#v", err)
}

func (suite *OCTestSuite) TestInt() {
	module := suite.mr.Module()

	for _, testRec := range suite.testRecordsInt {
		intPath := MustParamPath(string(testRec.key))
		ocInt, err := module.Int(intPath)
		suite.NoErrorf(err, "Cant find key %s in test onlineconf", string(testRec.key))
		testInt, err := strconv.Atoi(string(testRec.val[1:]))
		if err != nil {
			panic(fmt.Errorf("Cant parse test record int: %w", err))
		}
		suite.Equal(ocInt, testInt)

		ocInt, err = module.IntWithDef(intPath, 0)
		suite.NoErrorf(err, "Cant find key %s in test onlineconf", string(testRec.key))
		suite.Equal(ocInt, testInt)
	}

	for i, testRec := range suite.testRecordsInt {
		intPath := MustParamPath(string(testRec.key) + "_not_exists")
		ocInt, err := module.IntWithDef(intPath, i)
		suite.True(IsErrKeyNotFound(err), "non existing path: %s", string(testRec.key))
		suite.Equal(ocInt, i, "Default result was returned")
	}
}

func (suite *OCTestSuite) TestString() {
	module := suite.mr.Module()

	for _, testRec := range suite.testRecordsStr {
		strPath := MustParamPath(string(testRec.key))
		ocStr, err := module.String(strPath)
		suite.NoErrorf(err, "Cant find key %s in test onlineconf", string(testRec.key))
		suite.Equal(string(testRec.val[1:]), ocStr)

		ocStr, err = module.StringWithDef(strPath, "")
		suite.NoErrorf(err, "Cant find key %s in test onlineconf", string(testRec.key))
		suite.Equal(string(testRec.val[1:]), ocStr)

	}

	for i, testRec := range suite.testRecordsStr {
		defaultParamValue := "test_not_exists_" + strconv.Itoa(i)
		strPath := MustParamPath(string(testRec.key) + "_not_exists")
		ocStr, err := module.StringWithDef(strPath, defaultParamValue)
		suite.True(IsErrKeyNotFound(err), "non existing path: %s", string(testRec.key))
		suite.Equal(ocStr, defaultParamValue, "Default result was returned")
	}
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
