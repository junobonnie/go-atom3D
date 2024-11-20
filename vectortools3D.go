package atom3D

import (
	"math"
)

type Tensor struct {
	xx, xy, xz float64
	yx, yy, yz float64
	zx, zy, zz float64
}

func (t Tensor) Abs() float64 {
	return t.xx*t.yy*t.zz + t.xy*t.yz*t.zx + t.xz*t.yx*t.zy -
		t.xx*t.yz*t.zy - t.xy*t.yx*t.zz - t.xz*t.yy*t.zx
}

func (t Tensor) Add(other Tensor) Tensor {
	return Tensor{
		t.xx + other.xx, t.xy + other.xy, t.xz + other.xz,
		t.yx + other.yx, t.yy + other.yy, t.yz + other.yz,
		t.zx + other.zx, t.zy + other.zy, t.zz + other.zz}
}

func (t Tensor) Sub(other Tensor) Tensor {
	return Tensor{
		t.xx - other.xx, t.xy - other.xy, t.xz - other.xz,
		t.yx - other.yx, t.yy - other.yy, t.yz - other.yz,
		t.zx - other.zx, t.zy - other.zy, t.zz - other.zz}
}

func (t Tensor) Mul(num float64) Tensor {
	return Tensor{
		t.xx * num, t.xy * num, t.xz * num,
		t.yx * num, t.yy * num, t.yz * num,
		t.zx * num, t.zy * num, t.zz * num}
}

func (t Tensor) Div(num float64) Tensor {
	return Tensor{
		t.xx / num, t.xy / num, t.xz / num,
		t.yx / num, t.yy / num, t.yz / num,
		t.zx / num, t.zy / num, t.zz / num}
}

func (t Tensor) DotT(other Tensor) Tensor {
	return Tensor{
		t.xx*other.xx + t.xy*other.yx + t.xz*other.zx, t.xx*other.xy + t.xy*other.yy + t.xz*other.zy, t.xx*other.xz + t.xy*other.yz + t.xz*other.zz,
		t.yx*other.xx + t.yy*other.yx + t.yz*other.zx, t.yx*other.xy + t.yy*other.yy + t.yz*other.zy, t.yx*other.xz + t.yy*other.yz + t.yz*other.zz,
		t.zx*other.xx + t.zy*other.yx + t.zz*other.zx, t.zx*other.xy + t.zy*other.yy + t.zz*other.zy, t.zx*other.xz + t.zy*other.yz + t.zz*other.zz}
}

func (t Tensor) DotV(v Vector) Vector {
	return Vector{
		t.xx*v.x + t.xy*v.y + t.xz*v.z,
		t.yx*v.x + t.yy*v.y + t.yz*v.z,
		t.zx*v.x + t.zy*v.y + t.zz*v.z}
}

func (t Tensor) Inv() Tensor {
	det := t.Abs()
	if det == 0 {
		panic("Determinant is zero")
	}
	return Tensor{
		(t.yy*t.zz - t.yz*t.zy) / det, (t.xz*t.zy - t.xy*t.zz) / det, (t.xy*t.yz - t.xz*t.yy) / det,
		(t.yz*t.zx - t.yx*t.zz) / det, (t.xx*t.zz - t.xz*t.zx) / det, (t.xz*t.yx - t.xx*t.yz) / det,
		(t.yx*t.zy - t.yy*t.zx) / det, (t.xy*t.zx - t.xx*t.zy) / det, (t.xx*t.yy - t.xy*t.yx) / det}
}

func (t Tensor) Pow(num int) Tensor {
	result := Tensor{
		1, 0, 0,
		0, 1, 0,
		0, 0, 1}

	var pow_t Tensor
	if num >= 0 {
		pow_t = t
	} else {
		pow_t = t.Inv()
	}

	for i := 0; i < int(math.Abs(float64(num))); i++ {
		result = result.DotT(pow_t)
	}
	return result

}

func (t Tensor) T() Tensor {
	return Tensor{
		t.xx, t.yx, t.zx,
		t.xy, t.yy, t.zy,
		t.xz, t.yz, t.zz}
}

func SO3_z(angle float64) Tensor {
	return Tensor{
		math.Cos(angle), -math.Sin(angle), 0,
		math.Sin(angle), math.Cos(angle), 0,
		0, 0, 1}
}

func SO3_y(angle float64) Tensor {
	return Tensor{
		math.Cos(angle), 0, math.Sin(angle),
		0, 1, 0,
		-math.Sin(angle), 0, math.Cos(angle)}
}

func SO3_x(angle float64) Tensor {
	return Tensor{
		1, 0, 0,
		0, math.Cos(angle), -math.Sin(angle),
		0, math.Sin(angle), math.Cos(angle)}
}

type Vector struct {
	x, y, z float64
}

func (v Vector) Abs() float64 {
	return math.Sqrt(v.x*v.x + v.y*v.y + v.z*v.z)
}

func (v Vector) Add(other Vector) Vector {
	return Vector{v.x + other.x, v.y + other.y, v.z + other.z}
}

func (v Vector) Sub(other Vector) Vector {
	return Vector{v.x - other.x, v.y - other.y, v.z - other.z}
}

func (v Vector) Mul(num float64) Vector {
	return Vector{v.x * num, v.y * num, v.z * num}
}

func (v Vector) Div(num float64) Vector {
	return Vector{v.x / num, v.y / num, v.z / num}
}

func (v Vector) Dot(other Vector) float64 {
	return v.x*other.x + v.y*other.y + v.z*other.z
}

func (v Vector) Cross(other Vector) Vector {
	return Vector{
		v.y*other.z - v.z*other.y,
		v.z*other.x - v.x*other.z,
		v.x*other.y - v.y*other.x}
}

/*
func main() {
	t1 := Tensor{1, 2, 3, 4, 5, 6, 7, 8, 10}
	t2 := SO3_z(math.Pi / 4)
	t3 := Tensor{2, 3, 4, 5, 6, 7, 8, 9, 10}
	fmt.Println(t1.Abs())
	fmt.Println(t1.Add(t2))
	fmt.Println(t1.Sub(t2))
	fmt.Println(t1.Mul(2))
	fmt.Println(t1.Div(2))
	fmt.Println(t1.Pow(3))
	fmt.Println(t1.Pow(-1))
	fmt.Println(t1 == t3)
	fmt.Println(t1 != t3)
	fmt.Println(t1)
	fmt.Println(t1.Inv())
	fmt.Println(t1.T())
	fmt.Println(t1.DotT(t2))

	v1 := Vector{3, 4, 5}
	v2 := Vector{5, 4, 3}
	fmt.Println(t1.DotV(v1))
	fmt.Println(v1.Abs())
	fmt.Println(v1.Add(v2))
	fmt.Println(v1.Sub(v2))
	fmt.Println(v1.Mul(2))
	fmt.Println(v1.Div(2))
	fmt.Println(v1 == v2)
	fmt.Println(v1 != v2)
	fmt.Println(v1)
	fmt.Println(v1.Dot(v2))
	fmt.Println(v1.Cross(v2))
}
*/
