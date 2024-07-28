package xcoord

import "fmt"

// Point 直角坐标系点
type Point struct {
	x  float64
	y  float64
	ml int
}

// X 返回x坐标
func (p *Point) X() float64 {
	return p.x
}

// Y 返回y坐标
func (p *Point) Y() float64 {
	return p.y
}

// MLine 返回用于大地平面直角坐标系间转经纬度的子午线值
func (p *Point) MLine() int {
	return p.ml
}

// String 实现 Stringer 接口
func (p *Point) String() string {
	return fmt.Sprintf("(x: %v, y: %v)", p.x, p.y)
}

// NewPoint 生成一个直角坐标系的点；
// 可选参数 `ml` 是用于大地平面直角坐标系间转经纬度的子午线值
func NewPoint(x, y float64, ml ...int) *Point {
	p := &Point{
		x: x,
		y: y,
	}
	if len(ml) != 0 {
		p.ml = ml[0]
	}
	return p
}
