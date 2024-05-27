package xcoord

import (
	"fmt"
	"math"
)

// Location 地理位置(经纬度)
type Location struct {
	lng float64
	lat float64
}

// Longtitude 返回经度
func (l *Location) Longtitude() float64 {
	return l.lng
}

// Latitude 返回维度
func (l *Location) Latitude() float64 {
	return l.lat
}

// String 实现 Stringer 接口
func (l *Location) String() string {
	return fmt.Sprintf("(lng: %v, lat: %v)", l.lng, l.lat)
}

// Distance 根据经纬度计算距离(单位：m)
func (l *Location) Distance(t *Location) float64 {
	R := 6378137.0 // 地球半径
	rad := math.Pi / 180.0

	lng1 := l.lng * rad
	lat1 := l.lat * rad

	lng2 := t.lng * rad
	lat2 := t.lat * rad

	theta := lng2 - lng1

	dist := math.Sin(lat1)*math.Sin(lat2) + math.Cos(lat1)*math.Cos(lat2)*math.Cos(theta)

	return math.Acos(dist) * R
}

// Azimuth 根据经纬度计算方位角(0 ～ 360)
func (l *Location) Azimuth(t *Location) float64 {
	if t.lng == l.lng && t.lat == l.lat {
		return 0
	}

	if t.lng == l.lng {
		if t.lat > l.lat {
			return 0
		}

		return 180
	}

	if t.lat == l.lat {
		if t.lng > l.lng {
			return 90
		}

		return 270
	}

	rad := math.Pi / 180.0

	a := (90 - t.lat) * rad
	b := (90 - l.lat) * rad

	AOC_BOC := (t.lng - l.lng) * rad

	cosc := math.Cos(a)*math.Cos(b) + math.Sin(a)*math.Sin(b)*math.Cos(AOC_BOC)
	sinc := math.Sqrt(1 - cosc*cosc)

	sinA := math.Sin(a) * math.Sin(AOC_BOC) / sinc
	if sinA > 1 {
		sinA = 1
	}
	if sinA < -1 {
		sinA = -1
	}

	angle := math.Asin(sinA) / math.Pi * 180
	if t.lat < l.lat {
		return 180 - angle
	}
	if t.lng < l.lng {
		return 360 + angle
	}

	return angle
}

// NewLocation 生成一个Location
func NewLocation(lng, lat float64) *Location {
	return &Location{
		lng: lng,
		lat: lat,
	}
}
