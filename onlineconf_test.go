package onlineconf

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"testing"
	"unsafe"

	"github.com/colinmarc/cdb"
	"github.com/stretchr/testify/suite"
)

type testCDBRecord struct {
	key []byte
	val []byte
	exp interface{}
}

type testRecords struct {
	stringRecords []testCDBRecord
	intRecords    []testCDBRecord
	textStrings   []testCDBRecord
	jsonStrings   []testCDBRecord
	structRecords []testCDBRecord
	partialStruct testCDBRecord
	invalidStruct testCDBRecord
}

type OCTestSuite struct {
	suite.Suite
	cdbFile   *os.File
	cdbReader *cdb.CDB
	cdbWriter *cdb.Writer

	testRecords testRecords

	module *Module
}

func TestOCTestSuite(t *testing.T) {
	suite.Run(t, new(OCTestSuite))
}

func (suite *OCTestSuite) getCDBReader() *cdb.CDB {
	// initialize reader
	return suite.cdbReader
}

func (suite *OCTestSuite) getCDBWriter() *cdb.Writer {
	// initialize writer
	writer, err := cdb.Create(suite.cdbFile.Name())
	suite.Require().Nilf(err, "Can't get CDB writer: %#v", err)
	return writer
}

// generate test records
func generateTestRecords(count int) testRecords {
	testRecords := testRecords{
		stringRecords: make([]testCDBRecord, count),
		intRecords:    make([]testCDBRecord, count),
		textStrings:   make([]testCDBRecord, count),
		jsonStrings:   make([]testCDBRecord, count),
		structRecords: make([]testCDBRecord, count),
	}

	for i := 0; i < count; i++ {
		stri := strconv.Itoa(i)
		typeByte := "s"
		testRecords.stringRecords[i].key = []byte("/test/onlineconf/str" + stri)
		testRecords.stringRecords[i].val = []byte(typeByte + "val" + stri)

		// log.Printf("key %s val %s", string(testRecordsStr[i].key), string(testRecordsStr[i].val))

		testRecords.intRecords[i].key = []byte("/test/onlineconf/int" + stri)
		testRecords.intRecords[i].val = []byte(typeByte + stri)

		// log.Printf("key %s val %s", string(testRecordsInt[i].key), string(testRecordsInt[i].val))

		list := make([]string, 0, 10)
		for j := 0; j < 10; j++ {
			list = append(list, "value "+strconv.Itoa(i)+":"+strconv.Itoa(j))
		}
		testRecords.textStrings[i].key = []byte("/test/onlineconf/list" + stri)
		testRecords.textStrings[i].val = []byte("s" + strings.Join(list, ","))
		testRecords.textStrings[i].exp = list

		value, _ := json.Marshal(list)
		testRecords.jsonStrings[i].key = []byte("/test/onlineconf/array" + stri)
		testRecords.jsonStrings[i].val = append([]byte{'j'}, value...)
		testRecords.jsonStrings[i].exp = list

		data := map[string]string{}
		for j := 0; j < 10; j++ {
			data["key"+strconv.Itoa(j)] = "value " + strconv.Itoa(i) + ":" + strconv.Itoa(j)
		}
		value, _ = json.Marshal(data)
		testRecords.structRecords[i].key = []byte("/test/onlineconf/struct" + stri)
		testRecords.structRecords[i].val = append([]byte{'j'}, value...)
	}

	testRecords.partialStruct = testCDBRecord{
		key: []byte("/test/onlineconf/partial-struct"),
		val: []byte(`j{"key1":"value1", "key2":"value2"}`),
	}
	testRecords.invalidStruct = testCDBRecord{
		key: []byte("/test/onlineconf/invalid-struct"),
		val: []byte(`j{"key1":"value1", "key2":[]}`),
	}

	return testRecords
}

func (suite *OCTestSuite) SetupTest() {
	f, err := ioutil.TempFile("", "test_*.cdb")
	suite.Require().Nilf(err, "Can't open temporary file: %#v", err)

	suite.cdbFile = f

	suite.prepareTestData()

	suite.cdbReader, err = cdb.New(f, nil) // create new cdb handle
	suite.Require().NoError(err)

	suite.module = &Module{name: "testmodule", filename: f.Name()}
	suite.module.reopen()
}

func fillTestCDB(writer *cdb.Writer, testRecords testRecords) error {

	allTestRecords := []testCDBRecord{}
	allTestRecords = append(allTestRecords, testRecords.stringRecords...)
	allTestRecords = append(allTestRecords, testRecords.intRecords...)
	allTestRecords = append(allTestRecords, testRecords.textStrings...)
	allTestRecords = append(allTestRecords, testRecords.jsonStrings...)
	allTestRecords = append(allTestRecords, testRecords.structRecords...)
	allTestRecords = append(allTestRecords, testRecords.partialStruct)
	allTestRecords = append(allTestRecords, testRecords.invalidStruct)
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

	suite.testRecords = generateTestRecords(2)

	writer := suite.getCDBWriter()
	err := fillTestCDB(writer, suite.testRecords)
	suite.Nil(err)
	suite.Require().Nilf(err, "Cant put new value to cdb: %#v", err)
}

func (suite *OCTestSuite) TestInt() {
	module := suite.module

	for _, testRec := range suite.testRecords.intRecords {
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

	for _, testRec := range suite.testRecords.intRecords {
		_, ok := module.GetIntIfExists(string(testRec.key) + "_not_exists")
		suite.False(ok, "Cant find key %s in test onlineconf", string(testRec.key))
	}
}

func (suite *OCTestSuite) TestString() {
	module := suite.module

	for _, testRec := range suite.testRecords.stringRecords {
		ocStr, ok := module.GetStringIfExists(string(testRec.key))
		suite.True(ok, "Cant find key %s in test onlineconf", string(testRec.key))
		suite.Equal(string(testRec.val[1:]), ocStr)

		ocStr, ok = module.GetStringIfExists(string(testRec.key))
		suite.True(ok, "Cant find key %s in test onlineconf", string(testRec.key))
		suite.Equal(string(testRec.val[1:]), ocStr)

	}

	for _, testRec := range suite.testRecords.stringRecords {
		_, ok := module.GetStringIfExists(string(testRec.key) + "_not_exists")
		suite.False(ok, "Cant find key %s in test onlineconf", string(testRec.key))
	}
}

type testStruct struct {
	Key0 string
	Key1 string
	Key2 string
}

func (suite *OCTestSuite) TestStings() {
	module := suite.module

	for i := 0; i < 2; i++ {
		for _, testRec := range suite.testRecords.textStrings {
			strs := module.GetStrings(string(testRec.key), nil)
			suite.Equal(testRec.exp, strs, "unexpected []string value")
		}
		for _, testRec := range suite.testRecords.jsonStrings {
			strs := module.GetStrings(string(testRec.key), nil)
			suite.Equal(testRec.exp, strs, "unexpected []string value")
		}
	}

	defaultValue := []string{"default", "value"}

	strs := module.GetStrings("/not-exists", defaultValue)
	suite.Equal(defaultValue, strs, "default value must be returned if key is not exists")

	strs = module.GetStrings(string(suite.testRecords.invalidStruct.key), defaultValue)
	suite.Equal(defaultValue, strs, "default value must be returned if value can't be unmarshaled")
}

func (suite *OCTestSuite) TestStruct() {
	module := suite.module

	var v0 testStruct
	module.GetStruct(string(suite.testRecords.structRecords[0].key), &v0)

	for i := 0; i < 2; i++ {
		for _, testRec := range suite.testRecords.structRecords {
			var refValue map[string]string
			json.Unmarshal(testRec.val[1:], &refValue)
			var mapValue map[string]string
			ok, err := module.GetStruct(string(testRec.key), &mapValue)
			suite.True(ok, "Cant find key %s in test onlineconf", string(testRec.key))
			suite.Nil(err, "Error is not nil")
			suite.Equal(refValue, mapValue)
			var structValue testStruct
			ok, err = module.GetStruct(string(testRec.key), &structValue)
			suite.True(ok, "Cant find key %s in test onlineconf", string(testRec.key))
			suite.Nil(err, "Error is not nil")
			suite.Equal(testStruct{refValue["key0"], refValue["key1"], refValue["key2"]}, structValue)
		}
	}

	for _, testRec := range suite.testRecords.structRecords {
		var value map[string]string
		ok, err := module.GetStruct(string(testRec.key)+"_not_exists", &value)
		suite.False(ok, "Cant find key %s in test onlineconf", string(testRec.key))
		suite.Nil(err, "Error is not nil")
	}

	{
		var v1, v2 *testStruct
		module.GetStruct(string(suite.testRecords.structRecords[0].key), &v1)
		module.GetStruct(string(suite.testRecords.structRecords[0].key), &v2)
		suite.True(v1 == v2, "Pointers must be equal")
	}

	{
		var v1, v2 testStruct
		module.GetStruct(string(suite.testRecords.structRecords[0].key), &v1)
		module.GetStruct(string(suite.testRecords.structRecords[0].key), &v2)
		h1 := (*reflect.StringHeader)(unsafe.Pointer(&v1.Key0))
		h2 := (*reflect.StringHeader)(unsafe.Pointer(&v2.Key0))
		suite.True(h1.Data == h2.Data, "Underlying memory must be the same")
	}

	v0.Key0 = "xxx"
	var v1 testStruct
	module.GetStruct(string(suite.testRecords.structRecords[0].key), &v1)
	suite.NotEqual("xxx", v1.Key0, "Cached value should not be bound to value returned from the first call")

	orig := testStruct{Key0: "default0", Key1: "default1", Key2: "default2"}
	{
		val := orig
		module.GetStruct(string(suite.testRecords.partialStruct.key), &val)
		suite.Equal(testStruct{Key1: "value1", Key2: "value2"}, val, "Value must not be merged with default")
	}
	{
		val := orig
		module.GetStruct(string(suite.testRecords.invalidStruct.key), &val)
		suite.Equal(orig, val, "Default value must not be touched on error")
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
				for _, rec := range suite.testRecords.intRecords {
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
