# render.go — 3D 렌더러

소프트웨어 래스터라이저를 이용한 3D → 2D 투영 렌더링 모듈.  
투시 투영(perspective projection) + SO(3) 회전 + 깊이 정렬을 지원합니다.

---

## `Render` 구조체

```go
type Render struct {
    Width       float64  // 뷰포트 가로 (논리 단위)
    Height      float64  // 뷰포트 세로 (논리 단위)
    Depth       float64  // 원근감 깊이 (카메라 거리)
    Angle       Vector   // 오일러 각 (X, Y, Z, 라디안)
    FocusFactor float64  // 확대/축소 비율 (default 1.0)
}
```

실제 픽셀 해상도는 `10 × Width` × `10 × Height`.

### 생성자

```go
func NewRender(width, height, depth float64) *Render
```

`Angle = {0,0,0}`, `FocusFactor = 1.0` 으로 초기화.

---

## 투영 수식

카메라는 **+Y 방향**을 바라보며, Z축이 화면 위쪽입니다.

```
render_pos = R(Angle) · pos                  // 3D 회전
ratio      = FocusFactor × Depth / (render_pos.Y + Depth)  // 투시 비율
screen_x   = 5W + 10 × ratio × render_pos.X
screen_y   = 5H - 10 × ratio × render_pos.Z
```

---

## 메서드

| 메서드 | 설명 |
|---|---|
| `Figure() *gg.Context` | 픽셀 캔버스 생성 |
| `Background(dc, rgba)` | 배경색 채우기 |
| `DrawAtom(dc, pos, radius, rgba)` | 원형 파티클 그리기 |
| `DrawText(dc, pos, text, font_size, font, rgba)` | 3D 위치에 텍스트 |
| `DrawPlaneText(dc, x, y, text, font_size, font, rgba)` | 2D 화면 고정 텍스트 |
| `DrawLine(dc, pos1, pos2, width, rgba)` | 3D 선분 그리기 |
| `DrawCube(dc, pos, length, width, rgba)` | 와이어프레임 정육면체 |
| `DrawAxis(dc, length, width, font_size, font)` | XYZ 축 (RGB 색상) |
| `GetSortedIndices(pos []Vector) []int` | **화가 알고리즘** 깊이 정렬 인덱스 반환 |
| `Save(dc, directory, count)` | PNG 저장 (`<dir>/render_<count:010d>.png`) |

### `GetSortedIndices`

뷰 방향 벡터와의 내적을 기준으로 **먼 것부터 가까운 순서**로 인덱스를 정렬합니다 (화가 알고리즘).

---

## `RenderSO3(angle Vector) Tensor`

`SO3_x(angle.X) · SO3_y(angle.Y) · SO3_z(angle.Z)` 합성 회전 행렬을 반환합니다.

---

## 사용 예시

```go
render := atom3D.NewRender(100, 100, 300)
render.Angle = atom3D.Vector{0, math.Pi / 6, math.Pi / 4}

dc := render.Figure()
render.Background(dc, []float64{0, 0, 0, 1})

indices := render.GetSortedIndices(sim.Pos)
for _, i := range indices {
    render.DrawAtom(dc, sim.Pos[i], 1.0, []float64{1, 0.5, 0, 0.8})
}
render.DrawAxis(dc, 5, 2, 2, "Arial.ttf")
render.Save(dc, "images", sim.Count)
```
