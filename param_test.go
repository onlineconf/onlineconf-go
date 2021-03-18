package onlineconf

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigParams(t *testing.T) {

	var configParams = []ConfigParam{
		&ConfigParamInt{MustParamPath("/onlineconf/test/int"), 1, false},
		&ConfigParamString{MustParamPath("/onlineconf/test/string"), "test_string", false},
		&ConfigParamBool{MustParamPath("/onlineconf/test/bool"), true, false},
	}

	origLen := len(configParams)

	configPrefix := MustParamPath("/")
	ParamsPrefix(configPrefix, configParams)

	assert.Len(t, configParams, origLen, "configParams length not changed")

	for _, configParam := range configParams {
		confPathStr := configParam.GetPath().String()
		confPrefixStr := configPrefix.String()
		assert.Truef(t, strings.HasPrefix(confPathStr, confPrefixStr), "config %s has prefix %s")
	}
}
