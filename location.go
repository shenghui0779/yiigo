package yiigo

import (
	"fmt"
	"math"
)

// ProjType 投影类型
type ProjType int

const (
	UTM    ProjType = 0 // UTM投影
	GaussK ProjType = 1 // 高斯-克吕格(Gauss-Kruger)投影
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

// Point 直角坐标系点
type Point struct {
	x  float64
	y  float64
	ml int
}

// X 返回 X 坐标
func (p *Point) X() float64 {
	return p.x
}

// Y 返回 Y 坐标
func (p *Point) Y() float64 {
	return p.y
}

// String 实现 Stringer 接口
func (p *Point) String() string {
	return fmt.Sprintf("(x: %v, y: %v)", p.x, p.y)
}

// NewPoint 生成一个直角坐标系的点
func NewPoint(x, y float64) *Point {
	p := &Point{
		x: x,
		y: y,
	}

	return p
}

// Polar 极坐标系点
type Polar struct {
	rho float64
	rad float64
}

// Rad 返回极角(θ)弧度
func (p *Polar) Rad() float64 {
	return p.rad
}

// Angle 返回极角(θ)角度
func (p *Polar) Theta() float64 {
	return p.rad / math.Pi * 180
}

// Dist 返回极径(ρ)
func (p *Polar) Rho() float64 {
	return p.rho
}

// XY 转化为直角坐标系点
func (p *Polar) XY() *Point {
	return NewPoint(p.rho*math.Cos(p.rad), p.rho*math.Sin(p.rad))
}

// String 实现 Stringer 接口
func (p *Polar) String() string {
	return fmt.Sprintf("(ρ: %v, θ: %v)", p.rho, p.rad/math.Pi*180)
}

// NewPolar 生成一个极坐标点
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

// NewPolarFromXY 由直角坐标系点生成一个极坐标点
func NewPolarFromXY(x, y float64) *Polar {
	return &Polar{
		rho: math.Sqrt(math.Pow(x, 2) + math.Pow(y, 2)),
		rad: math.Atan2(y, x),
	}
}

// EllipsoidParameter 椭球体参数
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

// NewWGS84Parameter 生成 WGS84 椭球体参数
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
// [翻译自C++代码](https://www.cnblogs.com/xingzhensun/p/11377963.html)
type ZtGeoCoordTransform struct {
	ep *EllipsoidParameter
	ml int
	pt ProjType
}

// NewZtGeoCoordTransform 返回经纬度与大地平面直角坐标系间的转换器
func NewZtGeoCoordTransform(options ...ZGCTOption) *ZtGeoCoordTransform {
	zgct := &ZtGeoCoordTransform{
		ep: NewWGS84Parameter(),
		ml: -360,
		pt: GaussK,
	}

	for _, f := range options {
		f(zgct)
	}

	return zgct
}

// BL2XY 经纬度转大地平面直角坐标系点
func (zgct *ZtGeoCoordTransform) BL2XY(loc *Location) *Point {
	ml := zgct.ml

	if ml < -180 {
		ml = int((loc.lng+1.5)/3) * 3
	}

	lat := loc.lat * 0.0174532925199432957692
	dL := (loc.lng - float64(ml)) * 0.0174532925199432957692

	X := zgct.ep.A0*lat - zgct.ep.A2*math.Sin(2*lat)/2 + zgct.ep.A4*math.Sin(4*lat)/4 - zgct.ep.A6*math.Sin(6*lat)/6

	tn := math.Tan(lat)
	tn2 := tn * tn
	tn4 := tn2 * tn2

	j2 := (1/math.Pow(1-zgct.ep.F, 2) - 1) * math.Pow(math.Cos(lat), 2)
	n := zgct.ep.A / math.Sqrt(1.0-zgct.ep.E2*math.Sin(lat)*math.Sin(lat))

	var temp [6]float64

	temp[0] = n * math.Sin(lat) * math.Cos(lat) * dL * dL / 2
	temp[1] = n * math.Sin(lat) * math.Pow(math.Cos(lat), 3) * (5 - tn2 + 9*j2 + 4*j2*j2) * math.Pow(dL, 4) / 24
	temp[2] = n * math.Sin(lat) * math.Pow(math.Cos(lat), 5) * (61 - 58*tn2 + tn4) * math.Pow(dL, 6) / 720
	temp[3] = n * math.Cos(lat) * dL
	temp[4] = n * math.Pow(math.Cos(lat), 3) * (1 - tn2 + j2) * math.Pow(dL, 3) / 6
	temp[5] = n * math.Pow(math.Cos(lat), 5) * (5 - 18*tn2 + tn4 + 14*j2 - 58*tn2*j2) * math.Pow(dL, 5) / 120

	px := temp[3] + temp[4] + temp[5]
	py := X + temp[0] + temp[1] + temp[2]

	switch zgct.pt {
	case GaussK:
		px += 500000
	case UTM:
		px = px*0.9996 + 500000
		py = py * 0.9996
	}

	return &Point{
		x:  px,
		y:  py,
		ml: ml,
	}
}

// XY2BL 大地平面直角坐标系点转经纬度
func (zgct *ZtGeoCoordTransform) XY2BL(p *Point) *Location {
	x := p.x - 500000
	y := p.y

	if zgct.pt == UTM {
		x = x / 0.9996
		y = y / 0.9996
	}

	var (
		bf0       = y / zgct.ep.A0
		bf        float64
		threshold = 1.0
	)

	for threshold > 0.00000001 {
		y0 := -zgct.ep.A2*math.Sin(2*bf0)/2 + zgct.ep.A4*math.Sin(4*bf0)/4 - zgct.ep.A6*math.Sin(6*bf0)/6
		bf = (y - y0) / zgct.ep.A0

		threshold = bf - bf0
		bf0 = bf
	}

	t := math.Tan(bf)
	j2 := zgct.ep.EP2 * math.Pow(math.Cos(bf), 2)

	v := math.Sqrt(1 - zgct.ep.E2*math.Sin(bf)*math.Sin(bf))
	n := zgct.ep.A / v
	m := zgct.ep.A * (1 - zgct.ep.E2) / math.Pow(v, 3)

	temp0 := t * x * x / (2 * m * n)
	temp1 := t * (5 + 3*t*t + j2 - 9*j2*t*t) * math.Pow(x, 4) / (24 * m * math.Pow(n, 3))
	temp2 := t * (61 + 90*t*t + 45*math.Pow(t, 4)) * math.Pow(x, 6) / (720 * math.Pow(n, 5) * m)

	lat := (bf - temp0 + temp1 - temp2) * 57.29577951308232

	temp0 = x / (n * math.Cos(bf))
	temp1 = (1 + 2*t*t + j2) * math.Pow(x, 3) / (6 * math.Pow(n, 3) * math.Cos(bf))
	temp2 = (5 + 28*t*t + 6*j2 + 24*math.Pow(t, 4) + 8*t*t*j2) * math.Pow(x, 5) / (120 * math.Pow(n, 5) * math.Cos(bf))

	lng := (temp0-temp1+temp2)*57.29577951308232 + float64(p.ml)

	return &Location{
		lng: lng,
		lat: lat,
	}
}

// ZGCTOption 经纬度与大地平面直角坐标系间的转换选项
type ZGCTOption func(zgct *ZtGeoCoordTransform)

// WithMerLine 设置子午线值
func WithMerLine(ml int) ZGCTOption {
	return func(zgct *ZtGeoCoordTransform) {
		zgct.ml = ml
	}
}

// WithProjType 设置投影类型
func WithProjType(pt ProjType) ZGCTOption {
	return func(zgct *ZtGeoCoordTransform) {
		zgct.pt = pt
	}
}

// WithBaseLoc 设置基准点(以基准点子午线值建立坐标系)
func WithBaseLoc(loc *Location) ZGCTOption {
	return func(zgct *ZtGeoCoordTransform) {
		zgct.ml = int((loc.lng+1.5)/3) * 3
	}
}
