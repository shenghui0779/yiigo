package yiigo

import (
	"fmt"
	"math"
)

// Location geographic location
type Location struct {
	lng float64
	lat float64
}

// Longtitude returns longtitude
func (l *Location) Longtitude() float64 {
	return l.lng
}

// Latitude returns latitude
func (l *Location) Latitude() float64 {
	return l.lat
}

// String implements Stringer interface for print.
func (l *Location) String() string {
	return fmt.Sprintf("(lng: %.16f, lat: %.16f)", l.lng, l.lat)
}

// Distance calculates distance in meters with target location.
func (l *Location) Distance(t *Location) float64 {
	R := 6378137.0 // radius of the earth
	rad := math.Pi / 180.0

	lng1 := l.lng * rad
	lat1 := l.lat * rad

	lng2 := t.Longtitude() * rad
	lat2 := t.Latitude() * rad

	theta := lng2 - lng1

	dist := math.Sin(lat1)*math.Sin(lat2) + math.Cos(lat1)*math.Cos(lat2)*math.Cos(theta)

	return math.Acos(dist) * R
}

// Azimuth calculates azimuth angle with target location.
func (l *Location) Azimuth(t *Location) float64 {
	if t.Longtitude() == l.lng && t.Latitude() == l.lat {
		return 0
	}

	if t.Longtitude() == l.lng {
		if t.Latitude() > l.lat {
			return 0
		}

		return 180
	}

	if t.Latitude() == l.lat {
		if t.Longtitude() > l.lng {
			return 90
		}

		return 270
	}

	rad := math.Pi / 180.0

	a := (90 - t.Latitude()) * rad
	b := (90 - l.lat) * rad

	AOC_BOC := (t.Longtitude() - l.lng) * rad

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

	if t.Latitude() < l.lat {
		return 180 - angle
	}

	if t.Longtitude() < l.lng {
		return 360 + angle
	}

	return angle
}

// NewLocation returns a new location.
func NewLocation(lng, lat float64) *Location {
	return &Location{
		lng: lng,
		lat: lat,
	}
}

// Point coordinate point
type Point struct {
	x  float64
	y  float64
	ml float64
}

// X returns x
func (p *Point) X() float64 {
	return p.x
}

// Y returns y
func (p *Point) Y() float64 {
	return p.y
}

// MeridianLine returns meridian line for conversion between point and location.
func (p *Point) MeridianLine() float64 {
	return p.ml
}

// String implements Stringer interface for print.
func (p *Point) String() string {
	return fmt.Sprintf("(x: %.16f, y: %.16f)", p.x, p.y)
}

// PointOption point option
type PointOption func(p *Point)

// WithMeridianLine specifies the meridian line for point.
func WithMeridianLine(ml float64) PointOption {
	return func(p *Point) {
		p.ml = ml
	}
}

// NewPoint returns a new point.
func NewPoint(x, y float64, options ...PointOption) *Point {
	p := &Point{
		x: x,
		y: y,
	}

	for _, f := range options {
		f(p)
	}

	return p
}

// Polar polar coordinate point
type Polar struct {
	rho float64
	rad float64
}

// Rad returns radian of theta(θ)
func (p *Polar) Rad() float64 {
	return p.rad
}

// Angle returns theta(θ)
func (p *Polar) Theta() float64 {
	return p.rad / math.Pi * 180
}

// Dist returns pho(ρ)
func (p *Polar) Rho() float64 {
	return p.rho
}

// XY returns the point of X&Y coordinate.
func (p *Polar) XY(options ...PointOption) *Point {
	return NewPoint(p.rho*math.Cos(p.rad), p.rho*math.Sin(p.rad), options...)
}

// String implements Stringer interface for polar print.
func (p *Polar) String() string {
	return fmt.Sprintf("(ρ: %.16f, θ: %.16f)", p.rho, p.rad/math.Pi*180)
}

// NewPolar returns a new polar point.
func NewPolar(rho, theta float64) *Polar {
	rad := theta / 180 * math.Pi

	if math.Abs(rad) > math.Pi {
		rad = math.Atan2(rho*math.Sin(rad), rho*math.Cos(rad))
	}

	return &Polar{
		rho: rho,
		rad: rad,
	}
}

// NewPolarFromXY returns a new polar point from X&Y.
func NewPolarFromXY(x, y float64) *Polar {
	return &Polar{
		rho: math.Sqrt(math.Pow(x, 2) + math.Pow(y, 2)),
		rad: math.Atan2(y, x),
	}
}

// EllipsoidParameter params for ellipsoid.
type EllipsoidParameter struct {
	A   float64
	B   float64
	F   float64
	E2  float64
	EP2 float64
	C   float64
	A0  float64
	A2  float64
	A4  float64
	A6  float64
}

// NewWGS84Parameter params for WGS84.
func NewWGS84Parameter() *EllipsoidParameter {
	ep := &EllipsoidParameter{
		A:  6378137.0,
		E2: 0.00669437999013,
	}

	ep.B = math.Sqrt(ep.A * ep.A * (1 - ep.E2))
	ep.EP2 = (ep.A*ep.A - ep.B*ep.B) / (ep.B * ep.B)
	ep.F = (ep.A - ep.B) / ep.A

	// f0 := 1 / 298.257223563;
	// f1 := 1 / ep.F;

	ep.C = ep.A / (1 - ep.F)

	m0 := ep.A * (1 - ep.E2)
	m2 := 1.5 * ep.E2 * m0
	m4 := 1.25 * ep.E2 * m2
	m6 := 7 * ep.E2 * m4 / 6
	m8 := 9 * ep.E2 * m6 / 8

	ep.A0 = m0 + m2/2 + 3*m4/8 + 5*m6/16 + 35*m8/128
	ep.A2 = m2/2 + m4/2 + 15*m6/32 + 7*m8/16
	ep.A4 = m4/8 + 3*m6/16 + 7*m8/32
	ep.A6 = m6/32 + m8/16

	return ep
}

// ZtGeoCoordTransform 经纬度与大地平面直角坐标系间的转换
type ZtGeoCoordTransform struct {
	ep           *EllipsoidParameter
	meridianLine float64
	projType     rune
}

// NewZtGeoCoordTransform 返回经纬度与大地平面直角坐标系间的转换器
// eg: zgct := NewZtGeoCoordTransform(-360, 'g', NewWGS84Parameter())
func NewZtGeoCoordTransform(ml float64, pt rune, ep *EllipsoidParameter) *ZtGeoCoordTransform {
	return &ZtGeoCoordTransform{
		ep:           ep,
		meridianLine: ml,
		projType:     pt,
	}
}

// BL2XY 经纬度转大地平面直角坐标系点
func (zt *ZtGeoCoordTransform) BL2XY(loc *Location) *Point {
	meridianLine := zt.meridianLine

	if meridianLine < -180 {
		meridianLine = float64(int((loc.Longtitude()+1.5)/3) * 3)
	}

	lat := loc.Latitude() * 0.0174532925199432957692
	dL := (loc.Longtitude() - meridianLine) * 0.0174532925199432957692

	X := zt.ep.A0*lat - zt.ep.A2*math.Sin(2*lat)/2 + zt.ep.A4*math.Sin(4*lat)/4 - zt.ep.A6*math.Sin(6*lat)/6

	tn := math.Tan(lat)
	tn2 := tn * tn
	tn4 := tn2 * tn2

	j2 := (1/math.Pow(1-zt.ep.F, 2) - 1) * math.Pow(math.Cos(lat), 2)
	n := zt.ep.A / math.Sqrt(1.0-zt.ep.E2*math.Sin(lat)*math.Sin(lat))

	var temp [6]float64

	temp[0] = n * math.Sin(lat) * math.Cos(lat) * dL * dL / 2
	temp[1] = n * math.Sin(lat) * math.Pow(math.Cos(lat), 3) * (5 - tn2 + 9*j2 + 4*j2*j2) * math.Pow(dL, 4) / 24
	temp[2] = n * math.Sin(lat) * math.Pow(math.Cos(lat), 5) * (61 - 58*tn2 + tn4) * math.Pow(dL, 6) / 720
	temp[3] = n * math.Cos(lat) * dL
	temp[4] = n * math.Pow(math.Cos(lat), 3) * (1 - tn2 + j2) * math.Pow(dL, 3) / 6
	temp[5] = n * math.Pow(math.Cos(lat), 5) * (5 - 18*tn2 + tn4 + 14*j2 - 58*tn2*j2) * math.Pow(dL, 5) / 120

	px := temp[3] + temp[4] + temp[5]
	py := X + temp[0] + temp[1] + temp[2]

	switch zt.projType {
	case 'g':
		px += 500000
	case 'u':
		px = px*0.9996 + 500000
		py = py * 0.9996
	}

	return NewPoint(px, py, WithMeridianLine(meridianLine))
}

// XY2BL 大地平面直角坐标系点转经纬度
func (zt *ZtGeoCoordTransform) XY2BL(p *Point) *Location {
	x := p.X() - 500000
	y := p.Y()

	if zt.projType == 'u' {
		x = x / 0.9996
		y = y / 0.9996
	}

	var (
		bf0        = y / zt.ep.A0
		bf         float64
		threshould = 1.0
	)

	for threshould > 0.00000001 {
		y0 := -zt.ep.A2*math.Sin(2*bf0)/2 + zt.ep.A4*math.Sin(4*bf0)/4 - zt.ep.A6*math.Sin(6*bf0)/6
		bf = (y - y0) / zt.ep.A0

		threshould = bf - bf0
		bf0 = bf
	}

	t := math.Tan(bf)
	j2 := zt.ep.EP2 * math.Pow(math.Cos(bf), 2)

	v := math.Sqrt(1 - zt.ep.E2*math.Sin(bf)*math.Sin(bf))
	n := zt.ep.A / v
	m := zt.ep.A * (1 - zt.ep.E2) / math.Pow(v, 3)

	temp0 := t * x * x / (2 * m * n)
	temp1 := t * (5 + 3*t*t + j2 - 9*j2*t*t) * math.Pow(x, 4) / (24 * m * math.Pow(n, 3))
	temp2 := t * (61 + 90*t*t + 45*math.Pow(t, 4)) * math.Pow(x, 6) / (720 * math.Pow(n, 5) * m)

	lat := (bf - temp0 + temp1 - temp2) * 57.29577951308232

	temp0 = x / (n * math.Cos(bf))
	temp1 = (1 + 2*t*t + j2) * math.Pow(x, 3) / (6 * math.Pow(n, 3) * math.Cos(bf))
	temp2 = (5 + 28*t*t + 6*j2 + 24*math.Pow(t, 4) + 8*t*t*j2) * math.Pow(x, 5) / (120 * math.Pow(n, 5) * math.Cos(bf))

	lng := (temp0-temp1+temp2)*57.29577951308232 + p.MeridianLine()

	return NewLocation(lng, lat)
}
