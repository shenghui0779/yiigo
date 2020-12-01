package yiigo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDate(t *testing.T) {
	assert.Equal(t, "2016-03-19 15:03:19", Date(1458370999))
}

func TestStrToTime(t *testing.T) {
	assert.Equal(t, int64(1562910319), StrToTime("2019-07-12 13:45:19"))
}

func TestIP2Long(t *testing.T) {
	assert.Equal(t, uint32(3221234342), IP2Long("192.0.34.166"))
}

func TestLong2IP(t *testing.T) {
	assert.Equal(t, "192.0.34.166", Long2IP(uint32(3221234342)))
}

func TestVersionCompare(t *testing.T) {
	assert.True(t, VersionCompare("1.0.0", "1.0.0"))
	assert.False(t, VersionCompare("1.0.0", "1.0.1"))

	assert.True(t, VersionCompare("=1.0.0", "1.0.0"))
	assert.False(t, VersionCompare("=1.0.0", "1.0.1"))

	assert.True(t, VersionCompare("!=4.0.4", "4.0.0"))
	assert.False(t, VersionCompare("!=4.0.4", "4.0.4"))

	assert.True(t, VersionCompare(">2.0.0", "2.0.1"))
	assert.False(t, VersionCompare(">2.0.0", "1.0.1"))

	assert.True(t, VersionCompare(">=1.0.0&<2.0.0", "1.0.2"))
	assert.False(t, VersionCompare(">=1.0.0&<2.0.0", "2.0.1"))

	assert.True(t, VersionCompare("<2.0.0|>3.0.0", "1.0.2"))
	assert.True(t, VersionCompare("<2.0.0|>3.0.0", "3.0.1"))
	assert.False(t, VersionCompare("<2.0.0|>3.0.0", "2.0.1"))
}
