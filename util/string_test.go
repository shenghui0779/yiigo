package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddSlashes(t *testing.T) {
	assert.Equal(t, `Is your name O\'Reilly?`, AddSlashes("Is your name O'Reilly?"))
}

func TestStripSlashes(t *testing.T) {
	assert.Equal(t, "Is your name O'Reilly?", StripSlashes(`Is your name O\'Reilly?`))
}

func TestQuoteMeta(t *testing.T) {
	assert.Equal(t, `Hello world\. \(can you hear me\?\)`, QuoteMeta("Hello world. (can you hear me?)"))
}
