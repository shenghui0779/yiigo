package yiigo

import (
	"context"
	"errors"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSteps(t *testing.T) {
	arr := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}
	for index, step := range Steps(len(arr), 6) {
		ids := arr[step.Head:step.Tail]
		log.Printf("step[%d], slice: %d\n", index, ids)
	}
}

func TestMarshalNoEscapeHTML(t *testing.T) {
	data := map[string]string{"url": "https://github.com/shenghui0779/yiigo?id=996&name=yiigo"}

	b, err := MarshalNoEscapeHTML(data)
	assert.Nil(t, err)
	assert.Equal(t, string(b), `{"url":"https://github.com/shenghui0779/yiigo?id=996&name=yiigo"}`)
}

func TestVersionCompare(t *testing.T) {
	ok, err := VersionCompare("1.0.0", "1.0.0")
	assert.Nil(t, err)
	assert.True(t, ok)

	ok, err = VersionCompare("1.0.0", "1.0.1")
	assert.Nil(t, err)
	assert.False(t, ok)

	ok, err = VersionCompare("=1.0.0", "1.0.0")
	assert.Nil(t, err)
	assert.True(t, ok)

	ok, err = VersionCompare("=1.0.0", "1.0.1")
	assert.Nil(t, err)
	assert.False(t, ok)

	ok, err = VersionCompare("!=4.0.4", "4.0.0")
	assert.Nil(t, err)
	assert.True(t, ok)

	ok, err = VersionCompare("!=4.0.4", "4.0.4")
	assert.Nil(t, err)
	assert.False(t, ok)

	ok, err = VersionCompare(">2.0.0", "2.0.1")
	assert.Nil(t, err)
	assert.True(t, ok)

	ok, err = VersionCompare(">2.0.0", "1.0.1")
	assert.Nil(t, err)
	assert.False(t, ok)

	ok, err = VersionCompare(">=1.0.0&<2.0.0", "1.0.2")
	assert.Nil(t, err)
	assert.True(t, ok)

	ok, err = VersionCompare(">=1.0.0&<2.0.0", "2.0.1")
	assert.Nil(t, err)
	assert.False(t, ok)

	ok, err = VersionCompare("<2.0.0|>3.0.0", "1.0.2")
	assert.Nil(t, err)
	assert.True(t, ok)

	ok, err = VersionCompare("<2.0.0|>3.0.0", "3.0.1")
	assert.Nil(t, err)
	assert.True(t, ok)

	ok, err = VersionCompare("<2.0.0|>3.0.0", "2.0.1")
	assert.Nil(t, err)
	assert.False(t, ok)
}

func TestRetry(t *testing.T) {
	now := time.Now()
	err := Retry(context.Background(), func(ctx context.Context) error {
		fmt.Println("Retry...")
		return errors.New("something wrong")
	}, 3, time.Second)
	assert.NotNil(t, err)
	assert.Equal(t, 2, int(time.Since(now).Seconds()))
}

func TestIsUniqueDuplicateError(t *testing.T) {
	errMySQL := errors.New("Duplicate entry 'value' for key 'key_name'")
	assert.True(t, IsUniqueDuplicateError(errMySQL))

	errPgSQL := errors.New(`duplicate key value violates unique constraint "constraint_name"`)
	assert.True(t, IsUniqueDuplicateError(errPgSQL))

	errSQLite := errors.New("UNIQUE constraint failed: table_name.column_name")
	assert.True(t, IsUniqueDuplicateError(errSQLite))
}

func TestExcelColumnIndex(t *testing.T) {
	assert.Equal(t, 0, ExcelColumnIndex("A"))
	assert.Equal(t, 1, ExcelColumnIndex("B"))
	assert.Equal(t, 26, ExcelColumnIndex("AA"))
	assert.Equal(t, 27, ExcelColumnIndex("AB"))
}
