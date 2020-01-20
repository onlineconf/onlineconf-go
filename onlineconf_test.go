package onlineconf

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/alldroll/cdb"
	"github.com/stretchr/testify/assert"
)

func TestMmapError(t *testing.T) {
	assert := assert.New(t)
	_, err := loadModuleFromFile("")
	assert.NotNil(err)
}

func createBrokenCDB() (*os.File, error) {
	f, err := ioutil.TempFile("", "test_*.cdb")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	writer, err := cdb.New().GetWriter(f)
	if err != nil {
		return nil, err
	}

	err = writer.Put([]byte("/test/key"), []byte(""))
	defer writer.Close()
	if err != nil {
		return nil, err
	}

	return f, nil
}

func TestInvalidCDB(t *testing.T) {
	assert := assert.New(t)

	f, err := createBrokenCDB()
	assert.Nilf(err, "error while creating broken cdb: %#v", err)

	_, err = loadModuleFromFile(f.Name())
	// log.Printf("loadModuleFromFile err: %#v", err)
	assert.NotNil(err)
}
