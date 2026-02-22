# vectortools3D.go — 3D 벡터 & 텐서 수학 라이브러리

패키지 `atom3D`의 수치 계산 기반을 이루는 3차원 선형대수 타입과 연산을 제공합니다.

---

## `Vector` 타입

| 필드 | 타입 | 설명 |
|---|---|---|
| `X` | `float64` | x 성분 |
| `Y` | `float64` | y 성분 |
| `Z` | `float64` | z 성분 |

### 메서드

| 메서드 | 반환값 | 설명 |
|---|---|---|
| `Abs()` | `float64` | 크기 (Euclidean norm) `√(X²+Y²+Z²)` |
| `Add(other)` | `Vector` | 벡터 덧셈 |
| `Sub(other)` | `Vector` | 벡터 뺄셈 |
| `Mul(num)` | `Vector` | 스칼라 곱 |
| `Div(num)` | `Vector` | 스칼라 나눗셈 |
| `Dot(other)` | `float64` | 내적 |
| `Cross(other)` | `Vector` | 외적 |

---

## `Tensor` 타입

3×3 실수 행렬. 필드명은 `XX`, `XY`, `XZ`, `YX`, `YY`, `YZ`, `ZX`, `ZY`, `ZZ`.

### 메서드

| 메서드 | 반환값 | 설명 |
|---|---|---|
| `Abs()` | `float64` | 행렬식(determinant) |
| `Add(other)` | `Tensor` | 행렬 덧셈 |
| `Sub(other)` | `Tensor` | 행렬 뺄셈 |
| `Mul(num)` | `Tensor` | 스칼라 곱 |
| `Div(num)` | `Tensor` | 스칼라 나눗셈 |
| `DotT(other)` | `Tensor` | 행렬 곱 (matmul) |
| `DotV(v)` | `Vector` | 행렬-벡터 곱 |
| `Inv()` | `Tensor` | 역행렬 (det=0이면 panic) |
| `Pow(n)` | `Tensor` | 정수 거듭제곱 (n<0이면 역행렬 거듭제곱) |
| `T()` | `Tensor` | 전치(transpose) |

### SO(3) 회전 행렬 생성 함수

```go
SO3_x(angle float64) Tensor  // x축 회전
SO3_y(angle float64) Tensor  // y축 회전
SO3_z(angle float64) Tensor  // z축 회전
```

`angle`은 라디안 단위. 각 함수는 해당 축 방향으로의 능동 회전 행렬을 반환합니다.

---

## 사용 예시

```go
v1 := atom3D.Vector{1, 0, 0}
v2 := atom3D.Vector{0, 1, 0}
cross := v1.Cross(v2)   // {0, 0, 1}
dot   := v1.Dot(v2)     // 0

R := atom3D.SO3_z(math.Pi / 4)   // Z축 45° 회전
v3 := R.DotV(v1)                  // 회전된 벡터
```
