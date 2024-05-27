package xcoord

import (
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
