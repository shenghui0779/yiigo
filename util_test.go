package yiigo

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDate(t *testing.T) {
	assert.Equal(t, "2016-03-19 15:03:19", Date(1458370999))
}

func TestStrToTime(t *testing.T) {
	assert.Equal(t, int64(1562910319), StrToTime("2019-07-12 13:45:19"))
}

func TestWeekAround(t *testing.T) {
	monday, sunday := WeekAround(1562909685)

	assert.Equal(t, "2019-07-08", monday)
	assert.Equal(t, "2019-07-14", sunday)
}

func TestIP2Long(t *testing.T) {
	assert.Equal(t, uint32(3221234342), IP2Long("192.0.34.166"))
}

func TestLong2IP(t *testing.T) {
	assert.Equal(t, "192.0.34.166", Long2IP(uint32(3221234342)))
}

func TestMarshalNoEscapeHTML(t *testing.T) {
	data := map[string]string{"url": "https://github.com/shenghui0779/yiigo?id=996&name=yiigo"}

	b, err := MarshalNoEscapeHTML(data)

	assert.Nil(t, err)
	assert.Equal(t, string(b), `{"url":"https://github.com/shenghui0779/yiigo?id=996&name=yiigo"}`)
}

func TestAddSlashes(t *testing.T) {
	assert.Equal(t, `Is your name O\'Reilly?`, AddSlashes("Is your name O'Reilly?"))
}

func TestStripSlashes(t *testing.T) {
	assert.Equal(t, "Is your name O'Reilly?", StripSlashes(`Is your name O\'Reilly?`))
}

func TestQuoteMeta(t *testing.T) {
	assert.Equal(t, `Hello world\. \(can you hear me\?\)`, QuoteMeta("Hello world. (can you hear me?)"))
}

func TestSliceUniq(t *testing.T) {
	assert.Equal(t, []int{1, 2, 3, 4}, SliceUniq([]int{1, 2, 1, 3, 4, 3}))
	assert.Equal(t, []int64{1, 2, 3, 4}, SliceUniq([]int64{1, 2, 1, 3, 4, 3}))
	assert.Equal(t, []float64{1.01, 2.02, 3.03, 4.04}, SliceUniq([]float64{1.01, 2.02, 1.01, 3.03, 4.04, 3.03}))
	assert.Equal(t, []string{"h", "e", "l", "o"}, SliceUniq([]string{"h", "e", "l", "l", "o"}))
}

func TestSliceRand(t *testing.T) {
	a1 := []int{1, 2, 3, 4, 5, 6}
	ret1 := SliceRand(a1, 2)
	assert.Equal(t, 2, len(ret1))
	assert.NotEqual(t, a1[:2], ret1)

	a2 := []float64{1.01, 2.02, 3.03, 4.04, 5.05, 6.06}
	ret2 := SliceRand(a2, 2)
	assert.Equal(t, 2, len(ret2))
	assert.NotEqual(t, a2[:2], ret2)

	a3 := []string{"h", "e", "l", "l", "o"}
	ret3 := SliceRand(a3, 2)
	assert.Equal(t, 2, len(ret3))
	assert.NotEqual(t, a3[:2], ret3)

	type User struct {
		ID   int64
		Name string
	}

	a4 := []User{
		{
			ID:   1,
			Name: "h",
		},
		{
			ID:   2,
			Name: "e",
		},
		{
			ID:   3,
			Name: "l",
		},
		{
			ID:   4,
			Name: "l",
		},
		{
			ID:   5,
			Name: "o",
		},
	}
	ret4 := SliceRand(a4, 2)
	assert.Equal(t, 2, len(ret4))
	assert.NotEqual(t, a4[:2], ret4)
}

func TestCreateFile(t *testing.T) {
	f, err := CreateFile("app.log")

	if err != nil {
		assert.Fail(t, fmt.Sprintf("Expected nil, but got: %#v", err))

		return
	}

	f.Close()
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
