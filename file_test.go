package yiigo

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateFile(t *testing.T) {
	f, err := CreateFile("app.log")
	if err != nil {
		assert.Fail(t, fmt.Sprintf("Expected nil, but got: %#v", err))
		return
	}
	_ = f.Close()
}

func TestOpenFile(t *testing.T) {
	f, err := OpenFile("app.log")
	if err != nil {
		assert.Fail(t, fmt.Sprintf("Expected nil, but got: %#v", err))
		return
	}
	_ = f.Close()
}
