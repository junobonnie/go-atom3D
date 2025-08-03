package atom3D

import (
	"fmt"
	"math"
	"testing"
)

func TestVectortools3D(t *testing.T) {
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
