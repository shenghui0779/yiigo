package yiigo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValues(t *testing.T) {
	v1 := Values{}

	v1.Set("bar", "baz")
	v1.Set("foo", "quux")

	assert.Equal(t, "bar=baz&foo=quux", v1.Encode("=", "&"))
	assert.Equal(t, "bar:baz#foo:quux", v1.Encode(":", "#"))

	v2 := Values{}

	v2.Set("bar", "baz@666")
	v2.Set("foo", "quux%666")

	assert.Equal(t, "bar=baz%40666&foo=quux%25666", v2.EncodeEscape("=", "&"))
	assert.Equal(t, "bar:baz%40666#foo:quux%25666", v2.EncodeEscape(":", "#"))
}
