package yiigo

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

func TestXYBL(t *testing.T) {
	zgct := NewZtGeoCoordTransform(-360, GK)

	// 真值 (x: 440000, y: 4400000)
	p := zgct.BL2XY(NewLocation(116.300105669, 39.731939769))

	assert.Equal(t, 440000.0, math.Round(p.X()))
	assert.Equal(t, 4400000.0, math.Round(p.Y()))

	// 真值 (lng: 116.300105669, lat: 39.731939769)
	l := zgct.XY2BL(p)

	assert.Equal(t, "116.300105669", fmt.Sprintf("%.9f", l.Longtitude()))
	assert.Equal(t, "39.731939769", fmt.Sprintf("%.9f", l.Latitude()))
}
