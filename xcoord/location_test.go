package xcoord

import (
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDistance(t *testing.T) {
	loc1 := NewLocation(118.63173312, 31.94530239)
	loc2 := NewLocation(118.63343344, 31.94382162)

	// 期望值来自在线计算工具
	assert.Equal(t, 230.0, math.Round(loc1.Distance(loc2)))
}

func TestAzimuth(t *testing.T) {
	loc1 := NewLocation(118.63173312, 31.94530239)
	loc2 := NewLocation(118.63343344, 31.94382162)

	// 期望值来自在线计算工具
	assert.Equal(t, "135.7", fmt.Sprintf("%.1f", loc1.Azimuth(loc2)))
}
