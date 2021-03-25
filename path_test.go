package onlineconf

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type paramPathTestCase struct {
	path        string
	expectedErr error
	description string
}

func TestParamPathHappyCases(t *testing.T) {

	happyTestCases := []paramPathTestCase{
		{"/", nil, "root is valid path"},
		{"/onlineconf/test/int", nil, "long valid path"},
	}

	for _, testCase := range happyTestCases {
		rootPath, err := NewParamPath(testCase.path)
		assert.Nil(t, err, "NewParamPath check error: "+testCase.description)
		require.NotNil(t, rootPath, "rootPath must be not nil for happy cases")
		assert.Equal(t, rootPath.String(), testCase.path, "NewParamPath check string path: "+testCase.description)
	}
}

func TestParamPathUnhappyCases(t *testing.T) {

	unhappyTestCases := []paramPathTestCase{
		{"//", ErrTailSlash, "Bad root path"},
		{"/onlineconf/test//int", ErrDoubleSlash, "Double slash is prohibited"},
		{"onlineconf/test/int", ErrNoLeadingSlash, ""},
		{"/onlineconf/test/int/", ErrTailSlash, ""},
	}

	for _, testCase := range unhappyTestCases {
		rootPath, err := NewParamPath(testCase.path)
		assert.Equal(t, err, testCase.expectedErr, fmt.Sprintf("NewParamPath check path `%s` description : %s", testCase.path, testCase.description))
		assert.Nil(t, rootPath, "rootPath must be nil for unhappy cases")
	}
}

func TestParamPathJoin(t *testing.T) {
	rootPath := MustParamPath("/")
	otherPath := MustParamPath("/onlineconf/test")
	pathSuffix := MustParamPath("/path/to/subconfig")

	require.True(t, rootPath.IsRoot())

	rootJoinedWith := rootPath.Join(otherPath)
	assert.NoError(t, rootJoinedWith.IsValid(), "root path joined  with other one is valid")

	joinedWithRoot := otherPath.Join(rootPath)
	assert.NoError(t, joinedWithRoot.IsValid(), "other path joined with root path is valid")

	suffixedOtherPath := otherPath.Join(pathSuffix)
	assert.NoError(t, suffixedOtherPath.IsValid(), "suffixed other path is valid")
}
