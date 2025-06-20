package onlineconf

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
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
	stringRecords   []testCDBRecord
	intRecords      []testCDBRecord
	boolRecords     []testCDBRecord
	durationRecords []testCDBRecord
	floatRecords    []testCDBRecord
	textStrings     []testCDBRecord
	jsonStrings     []testCDBRecord
	structRecords   []testCDBRecord
	partialStruct   testCDBRecord
	invalidStruct   testCDBRecord
}

type ocTestSuite struct {
	suite.Suite
	cdbFile *os.File

	testRecords testRecords

	module *Module
}

func TestOCTestSuite(t *testing.T) {
	suite.Run(t, new(ocTestSuite))
}

// initialize writer
func (suite *ocTestSuite) getCDBWriter() *cdb.Writer {
	writer, err := cdb.Create(suite.cdbFile.Name())
	suite.Require().Nilf(err, "Can't get CDB writer: %#v", err)

	return writer
}

// generate test records
func generateTestRecords(count int) testRecords {
	ret := testRecords{
		stringRecords:   make([]testCDBRecord, count),
		intRecords:      make([]testCDBRecord, count),
		boolRecords:     make([]testCDBRecord, count),
		durationRecords: make([]testCDBRecord, count),
		floatRecords:    make([]testCDBRecord, count),
		textStrings:     make([]testCDBRecord, count),
		jsonStrings:     make([]testCDBRecord, count),
		structRecords:   make([]testCDBRecord, count),
	}

	for i := 0; i < count; i++ {
		stri := strconv.Itoa(i)
		typeByte := "s"
		ret.stringRecords[i].key = []byte("/test/onlineconf/str" + stri)
		ret.stringRecords[i].val = []byte(typeByte + "val" + stri)

		// log.Printf("key %s val %s", string(testRecordsStr[i].key), string(testRecordsStr[i].val))

		ret.intRecords[i].key = []byte("/test/onlineconf/int" + stri)
		ret.intRecords[i].val = []byte(typeByte + stri)

		boolVal := "s"

		switch i % 3 {
		case 1:
			boolVal = "s0"
		case 2:
			boolVal = "s1"
		}

		ret.boolRecords[i].key = []byte("/test/onlineconf/bool" + stri)
		ret.boolRecords[i].val = []byte(boolVal)

		ret.durationRecords[i].key = []byte("/test/onlineconf/duration" + stri)

		switch i % 3 {
		case 0:
			ret.durationRecords[i].val = []byte("s" + stri + "ms")
		case 1:
			ret.durationRecords[i].val = []byte("s" + stri + "s")
		case 2:
			ret.durationRecords[i].val = []byte("s" + stri)
		}

		ret.floatRecords[i].key = []byte("/test/onlineconf/float" + stri)
		ret.floatRecords[i].val = []byte("s" + strconv.FormatFloat(float64(i)/2, 'f', 2, 64))

		// log.Printf("key %s val %s", string(testRecordsInt[i].key), string(testRecordsInt[i].val))

		list := make([]string, 0, 10)
		for j := 0; j < 10; j++ {
			list = append(list, "value "+strconv.Itoa(i)+":"+strconv.Itoa(j))
		}

		ret.textStrings[i].key = []byte("/test/onlineconf/list" + stri)
		ret.textStrings[i].val = []byte("s" + strings.Join(list, ","))
		ret.textStrings[i].exp = list

		value, _ := json.Marshal(list)
		ret.jsonStrings[i].key = []byte("/test/onlineconf/array" + stri)
		ret.jsonStrings[i].val = append([]byte{'j'}, value...) //nolint:wsl
		ret.jsonStrings[i].exp = list

		data := map[string]string{}
		for j := 0; j < 10; j++ {
			data["key"+strconv.Itoa(j)] = "value " + strconv.Itoa(i) + ":" + strconv.Itoa(j)
		}

		value, _ = json.Marshal(data)
		ret.structRecords[i].key = []byte("/test/onlineconf/struct" + stri)
		ret.structRecords[i].val = append([]byte{'j'}, value...) //nolint:wsl
	}

	ret.partialStruct = testCDBRecord{
		key: []byte("/test/onlineconf/partial-struct"),
		val: []byte(`j{"key1":"value1", "key2":"value2"}`),
	}
	ret.invalidStruct = testCDBRecord{
		key: []byte("/test/onlineconf/invalid-struct"),
		val: []byte(`j{"key1":"value1", "key2":[]}`),
	}

	return ret
}

func (suite *ocTestSuite) SetupTest() {
	f, err := os.CreateTemp("", "test_*.cdb")
	suite.Require().Nilf(err, "Can't open temporary file: %#v", err)

	suite.cdbFile = f

	suite.prepareTestData()

	suite.module, err = OpenModule(f.Name())
	suite.NoError(err, "OpenModule() failed")
}

func (suite *ocTestSuite) TearDownTest() {
	os.Remove(suite.module.filename)
}

func fillTestCDB(writer *cdb.Writer, testRecords testRecords) error {
	allTestRecords := []testCDBRecord{}
	allTestRecords = append(allTestRecords, testRecords.stringRecords...)
	allTestRecords = append(allTestRecords, testRecords.intRecords...)
	allTestRecords = append(allTestRecords, testRecords.boolRecords...)
	allTestRecords = append(allTestRecords, testRecords.durationRecords...)
	allTestRecords = append(allTestRecords, testRecords.floatRecords...)
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

	if err := writer.Close(); err != nil {
		return err
	}

	return nil
}

func (suite *ocTestSuite) prepareTestData() {
	suite.testRecords = generateTestRecords(5)

	writer := suite.getCDBWriter()
	err := fillTestCDB(writer, suite.testRecords)
	suite.Nil(err)
	suite.Require().Nilf(err, "Cant put new value to cdb: %#v", err)
}

func (suite *ocTestSuite) TestModCache() {
	mod2, err := OpenModule(suite.module.filename)
	suite.NoError(err, "OpenModule() failed")
	suite.Equal(suite.module, mod2, "module should be cached")
}

func (suite *ocTestSuite) TestInt() {
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

		ocInt = module.GetInt(string(testRec.key), 0xDEADC0DE)
		suite.Equal(ocInt, testInt)
	}

	for _, testRec := range suite.testRecords.intRecords {
		_, ok := module.GetIntIfExists(string(testRec.key) + "_not_exists")
		suite.False(ok, "Cant find key %s in test onlineconf", string(testRec.key))

		got := module.GetInt(string(testRec.key)+"_not_exists", 0xDEADC0DE)
		suite.Equal(got, 0xDEADC0DE, "GetInt(%s_not_exists) should return a default value", testRec.key)
	}
}

func (suite *ocTestSuite) TestBrokenInts() {
	defer log.SetOutput(os.Stderr)

	for _, tr := range suite.testRecords.textStrings {
		buf := &bytes.Buffer{}
		log.SetOutput(buf)

		_, ok := suite.module.GetIntIfExists(string(tr.key))
		suite.False(ok, "GetIntIfExists(%s) shouldn't return ok", tr.key)
		suite.Contains(buf.String(), "invalid syntax")
	}
}

func (suite *ocTestSuite) TestBool() {
	for i, tr := range suite.testRecords.boolRecords {
		b, ok := suite.module.GetBoolIfExists(string(tr.key))
		suite.True(ok, "Can't find %s key in the test onlineconf", tr.key)

		bb := suite.module.GetBool(string(tr.key), false)

		switch i % 3 {
		case 0, 1:
			suite.False(b, "%s is true, but false is expected", tr.key)
			suite.False(bb, "%s is true, but false is expected", tr.key)
		case 2:
			suite.True(b, "%s is false, but true is expected", tr.key)
			suite.True(bb, "%s is false, but true is expected", tr.key)
		}
	}

	b := suite.module.GetBool("/not/existent", true)
	suite.True(b)
}

func (suite *ocTestSuite) TestDuration() {
	defer log.SetOutput(os.Stderr)

	for i, tr := range suite.testRecords.durationRecords {
		want := time.Duration(i) * time.Second
		if i%3 == 0 {
			want = time.Duration(i) * time.Millisecond
		}

		got := suite.module.GetDuration(string(tr.key), 0)
		suite.Equal(want, got)

		buf := &bytes.Buffer{}
		log.SetOutput(buf)

		got = suite.module.GetDuration(string(suite.testRecords.stringRecords[i].key), time.Duration(456))
		suite.Equal(time.Duration(456), got)
		suite.Contains(buf.String(), "invalid duration")
	}

	got := suite.module.GetDuration("/not/existent", 123*time.Second)
	suite.Equal(123*time.Second, got)
}

func (suite *ocTestSuite) TestFloat() {
	defer log.SetOutput(os.Stderr)

	for i, tr := range suite.testRecords.floatRecords {
		want := float64(i) / 2

		got := suite.module.GetFloat(string(tr.key), 0)
		suite.Equal(want, got)

		buf := &bytes.Buffer{}
		log.SetOutput(buf)

		got = suite.module.GetFloat(string(suite.testRecords.stringRecords[i].key), 2.7)
		suite.Equal(2.7, got)
		suite.Contains(buf.String(), "invalid syntax")
	}

	got := suite.module.GetFloat("/not/existent", 3.14)
	suite.Equal(3.14, got)
}

func (suite *ocTestSuite) TestString() {
	module := suite.module

	for _, testRec := range suite.testRecords.stringRecords {
		ocStr, ok := module.GetStringIfExists(string(testRec.key))
		suite.True(ok, "Cant find key %s in test onlineconf", string(testRec.key))
		suite.Equal(string(testRec.val[1:]), ocStr)

		ocStr, ok = module.GetStringIfExists(string(testRec.key))
		suite.True(ok, "Cant find key %s in test onlineconf", string(testRec.key))
		suite.Equal(string(testRec.val[1:]), ocStr)

		ocStr = module.GetString(string(testRec.key), "not found")
		suite.NotEqual(ocStr, "not found")
	}

	for _, testRec := range suite.testRecords.stringRecords {
		_, ok := module.GetStringIfExists(string(testRec.key) + "_not_exists")
		suite.False(ok, "Found missing key %s", string(testRec.key))

		got := module.GetString(string(testRec.key)+"_not_exists", "not found")
		suite.Equal(got, "not found")
	}
}

func (suite *ocTestSuite) TestTypeMismatch() {
	defer log.SetOutput(os.Stderr)

	for _, tr := range suite.testRecords.structRecords {
		_, err := suite.module.GetStringErr(string(tr.key))
		suite.ErrorIs(err, ErrFormatIsNotString)

		buf := &bytes.Buffer{}
		log.SetOutput(buf)

		_, ok := suite.module.GetStringIfExists(string(tr.key))
		suite.False(ok, "%s should not be treated as a string", tr.key)
		suite.Contains(buf.String(), ErrFormatIsNotString.Error(), "ErrFormatIsNotString should be logged")

		buf.Reset()

		_, ok = suite.module.GetBoolIfExists(string(tr.key))
		suite.False(ok, "%s should not be treated as a bool", tr.key)
		suite.Contains(buf.String(), ErrFormatIsNotString.Error(), "ErrFormatIsNotString should be logged")
	}
}

type testStruct struct {
	Key0 string
	Key1 string
	Key2 string
}

func (suite *ocTestSuite) TestStrings() {
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

	buf := &bytes.Buffer{}
	log.SetOutput(buf)

	defer log.SetOutput(os.Stderr)

	strs = module.GetStrings(string(suite.testRecords.invalidStruct.key), defaultValue)
	suite.Equal(defaultValue, strs, "default value must be returned if value can't be unmarshaled")
	suite.Contains(buf.String(), "failed to unmarshal JSON", "unmarshal error isn't logged")
}

func (suite *ocTestSuite) testGetStruct(module *Module, key []byte, present bool, valPtr interface{}) {
	ok, err := module.GetStruct(string(key), valPtr)

	suite.Equal(present, ok, "GetStruct(%s) key presence check failed")
	suite.NoError(err, "GetStruct(%s) failed", key)
}

func (suite *ocTestSuite) TestStruct() {
	module := suite.module
	v0 := testStruct{}

	suite.testGetStruct(module, suite.testRecords.structRecords[0].key, true, &v0)

	for i := 0; i < 2; i++ {
		for _, testRec := range suite.testRecords.structRecords {
			var refValue, mapValue map[string]string

			suite.NoError(json.Unmarshal(testRec.val[1:], &refValue), "json.Unmarshal(%s) failed", testRec.val[1:])
			suite.testGetStruct(module, testRec.key, true, &mapValue)
			suite.Equal(refValue, mapValue)

			var structValue testStruct

			suite.testGetStruct(module, testRec.key, true, &structValue)
			suite.Equal(testStruct{refValue["key0"], refValue["key1"], refValue["key2"]}, structValue)
		}
	}

	for _, testRec := range suite.testRecords.structRecords {
		var value map[string]string

		suite.testGetStruct(module, append(append([]byte(nil), testRec.key...), "_not_exists"...), false, &value)
	}

	{
		var v1, v2 *testStruct

		suite.testGetStruct(module, suite.testRecords.structRecords[0].key, true, &v1)
		suite.testGetStruct(module, suite.testRecords.structRecords[0].key, true, &v2)
		suite.True(v1 == v2, "Pointers must be equal")
	}

	{
		var v1, v2 testStruct

		suite.testGetStruct(module, suite.testRecords.structRecords[0].key, true, &v1)
		suite.testGetStruct(module, suite.testRecords.structRecords[0].key, true, &v2)

		d1 := unsafe.StringData(v1.Key0)
		d2 := unsafe.StringData(v2.Key0)

		suite.True(d1 == d2, "Underlying memory must be the same")
	}

	v0.Key0 = "xxx"

	suite.testGetStruct(module, suite.testRecords.structRecords[0].key, true, &testStruct{})

	orig := testStruct{Key0: "default0", Key1: "default1", Key2: "default2"}
	{
		val := orig
		suite.testGetStruct(module, suite.testRecords.partialStruct.key, true, &val)
		suite.Equal(testStruct{Key1: "value1", Key2: "value2"}, val, "Value must not be merged with default")
	}
	{
		val := orig
		ok, err := module.GetStruct(string(suite.testRecords.invalidStruct.key), &val)
		suite.False(ok, "GetStruct(%s) should return an error", suite.testRecords.invalidStruct.key)
		suite.Error(err, "GetStruct(%s) should return an error", suite.testRecords.invalidStruct.key)
		suite.Equal(orig, val, "Default value must not be touched on error")
	}
}

func (suite *ocTestSuite) TestReload() {
	// todo
}

func (suite *ocTestSuite) TestConcurrent() {
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

// please do not copy this to your projects! (especially to libraries)
type onlineconfIface interface {
	Path(path string) string
	GetStringErr(path string) (string, error)
	GetStringIfExists(path string) (string, bool)
	GetString(path string, dfl string) string
	GetIntErr(path string) (int, error)
	GetIntIfExists(path string) (int, bool)
	GetInt(path string, dfl int) int
	GetBoolErr(path string) (bool, error)
	GetBoolIfExists(path string) (bool, bool)
	GetBool(path string, dfl bool) bool
	GetDurationErr(path string) (time.Duration, error)
	GetDurationIsExists(path string) (time.Duration, bool)
	GetDuration(path string, dfl time.Duration) time.Duration
	GetFloatErr(path string) (float64, error)
	GetFloatIfExists(path string) (float64, bool)
	GetFloat(path string, dfl float64) float64
	GetStrings(path string, dfl []string) []string
	GetStruct(path string, valuePtr interface{}) (bool, error)
	Subtree(prefix string) *Subtree
	SubscribeChan(path string, ch chan<- struct{}) error
	Subscribe(path string) (chan struct{}, error)
	SubscribeChanSubtree(path string, ch chan<- struct{}) error
	SubscribeSubtree(path string) (chan struct{}, error)
	UnsubscribeChan(path string, ch chan<- struct{})
	UnsubscribeChanSubtree(path string, ch chan<- struct{})
	Unsubscribe(path string)
	UnsubscribeSubtree(path string)
}

var (
	_ onlineconfIface = &Module{}
	_ onlineconfIface = &Subtree{}
)
