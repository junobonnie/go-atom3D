package atom3D

import (
	"math"
)

type Tensor struct {
	XX, XY, XZ float64
	YX, YY, YZ float64
	ZX, ZY, ZZ float64
}

func (t Tensor) Abs() float64 {
	return t.XX*t.YY*t.ZZ + t.XY*t.YZ*t.ZX + t.XZ*t.YX*t.ZY -
		t.XX*t.YZ*t.ZY - t.XY*t.YX*t.ZZ - t.XZ*t.YY*t.ZX
}

func (t Tensor) Add(other Tensor) Tensor {
	return Tensor{
		t.XX + other.XX, t.XY + other.XY, t.XZ + other.XZ,
		t.YX + other.YX, t.YY + other.YY, t.YZ + other.YZ,
		t.ZX + other.ZX, t.ZY + other.ZY, t.ZZ + other.ZZ}
}

func (t Tensor) Sub(other Tensor) Tensor {
	return Tensor{
		t.XX - other.XX, t.XY - other.XY, t.XZ - other.XZ,
		t.YX - other.YX, t.YY - other.YY, t.YZ - other.YZ,
		t.ZX - other.ZX, t.ZY - other.ZY, t.ZZ - other.ZZ}
}

func (t Tensor) Mul(num float64) Tensor {
	return Tensor{
		t.XX * num, t.XY * num, t.XZ * num,
		t.YX * num, t.YY * num, t.YZ * num,
		t.ZX * num, t.ZY * num, t.ZZ * num}
}

func (t Tensor) Div(num float64) Tensor {
	return Tensor{
		t.XX / num, t.XY / num, t.XZ / num,
		t.YX / num, t.YY / num, t.YZ / num,
		t.ZX / num, t.ZY / num, t.ZZ / num}
}

func (t Tensor) DotT(other Tensor) Tensor {
	return Tensor{
		t.XX*other.XX + t.XY*other.YX + t.XZ*other.ZX, t.XX*other.XY + t.XY*other.YY + t.XZ*other.ZY, t.XX*other.XZ + t.XY*other.YZ + t.XZ*other.ZZ,
		t.YX*other.XX + t.YY*other.YX + t.YZ*other.ZX, t.YX*other.XY + t.YY*other.YY + t.YZ*other.ZY, t.YX*other.XZ + t.YY*other.YZ + t.YZ*other.ZZ,
		t.ZX*other.XX + t.ZY*other.YX + t.ZZ*other.ZX, t.ZX*other.XY + t.ZY*other.YY + t.ZZ*other.ZY, t.ZX*other.XZ + t.ZY*other.YZ + t.ZZ*other.ZZ}
}

func (t Tensor) DotV(v Vector) Vector {
	return Vector{
		t.XX*v.X + t.XY*v.Y + t.XZ*v.Z,
		t.YX*v.X + t.YY*v.Y + t.YZ*v.Z,
		t.ZX*v.X + t.ZY*v.Y + t.ZZ*v.Z}
}

func (t Tensor) Inv() Tensor {
	det := t.Abs()
	if det == 0 {
		panic("Determinant is zero")
	}
	return Tensor{
		(t.YY*t.ZZ - t.YZ*t.ZY) / det, (t.XZ*t.ZY - t.XY*t.ZZ) / det, (t.XY*t.YZ - t.XZ*t.YY) / det,
		(t.YZ*t.ZX - t.YX*t.ZZ) / det, (t.XX*t.ZZ - t.XZ*t.ZX) / det, (t.XZ*t.YX - t.XX*t.YZ) / det,
		(t.YX*t.ZY - t.YY*t.ZX) / det, (t.XY*t.ZX - t.XX*t.ZY) / det, (t.XX*t.YY - t.XY*t.YX) / det}
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
		t.XX, t.YX, t.ZX,
		t.XY, t.YY, t.ZY,
		t.XZ, t.YZ, t.ZZ}
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
	X, Y, Z float64
}

func (v Vector) Abs() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y + v.Z*v.Z)
}

func (v Vector) Add(other Vector) Vector {
	return Vector{v.X + other.X, v.Y + other.Y, v.Z + other.Z}
}

func (v Vector) Sub(other Vector) Vector {
	return Vector{v.X - other.X, v.Y - other.Y, v.Z - other.Z}
}

func (v Vector) Mul(num float64) Vector {
	return Vector{v.X * num, v.Y * num, v.Z * num}
}

func (v Vector) Div(num float64) Vector {
	return Vector{v.X / num, v.Y / num, v.Z / num}
}

func (v Vector) Dot(other Vector) float64 {
	return v.X*other.X + v.Y*other.Y + v.Z*other.Z
}

func (v Vector) Cross(other Vector) Vector {
	return Vector{
		v.Y*other.Z - v.Z*other.Y,
		v.Z*other.X - v.X*other.Z,
		v.X*other.Y - v.Y*other.X}
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
