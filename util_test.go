package yiigo

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

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

type ctxKey string

func TestDetachContext(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	key := ctxKey("key")
	ctx = context.WithValue(ctx, key, "value")
	dctx := DetachContext(ctx)

	// Detached context has the same values.
	got, ok := dctx.Value(key).(string)
	if !ok || got != "value" {
		t.Errorf("Value: got (%v, %t), want 'value', true", got, ok)
	}

	// Detached context doesn't time out.
	time.Sleep(500 * time.Millisecond)
	if err := ctx.Err(); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("original context Err: got %v, want DeadlineExceeded", err)
	}
	if err := dctx.Err(); err != nil {
		t.Errorf("detached context Err: got %v, want nil", err)
	}
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
