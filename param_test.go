package onlineconf

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var testParamInt *ConfigParamInt = MustConfigParamInt("/onlineconf/test/int", 1)
var testParamStrign *ConfigParamString = MustConfigParamString("/onlineconf/test/string", "test_string")
var testParamBool *ConfigParamBool = MustConfigParamBool("/onlineconf/test/bool", true)

func TestConfigParams(t *testing.T) {

	var configParams = []ConfigParam{
		testParamInt,
		testParamStrign,
		testParamBool,
	}

	origLen := len(configParams)

	configPrefix := MustParamPath("/")
	err := ParamsPrefix(configPrefix, configParams)
	assert.NoError(t, err)

	assert.Len(t, configParams, origLen, "configParams length not changed")

	for _, configParam := range configParams {
		confPathStr := configParam.GetPath().String()
		confPrefixStr := configPrefix.String()
		assert.Truef(t, strings.HasPrefix(confPathStr, confPrefixStr), "config %s has prefix %s")
	}
}
