//    formatter_test.go
//    ~~~~~~~~~
//    This module implements the Formatter tests.
//    :authors: Konstantin Bokarius.
//    :copyright: (c) 2015 by Fanout, Inc.
//    :license: MIT, see LICENSE for more details.

package pubcontrol

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type FormatTestStruct struct {
	value string
}

func (format *FormatTestStruct) Name() string {
	return "test-format"
}
func (format *FormatTestStruct) Export() interface{} {
	return format.value
}

func TestFormatter(t *testing.T) {
	fmt := &FormatTestStruct{value: "value"}
	assert.Equal(t, fmt.Name(), "test-format")
	assert.Equal(t, fmt.Export(), "value")
}
