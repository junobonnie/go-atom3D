# atom3D.go — 핵심 시뮬레이터 (`Simulator`)

`atom3D` 패키지의 기본 N체 시뮬레이터를 정의합니다. 시간 적분, 경계 조건, 그리드 기반 이웃 탐색, HDF5 스냅샷 저장/읽기 기능을 포함합니다.

---

## `Simulator` 구조체

```go
type Simulator struct {
    Dt         float64   // 타임스텝 크기
    T          float64   // 현재 물리 시간
    Count      int       // 누적 스텝 수
    N          int       // 파티클 수
    Id         []int     // 파티클 ID 배열
    Pos        []Vector  // 위치 배열 [N]
    Vel        []Vector  // 속도 배열 [N]
    Gravity    Vector    // 외부 균일 중력 가속도
    RegionSize float64   // 그리드 탐색을 위한 시뮬레이션 영역 크기
    GridSize   float64   // 이웃 탐색 그리드 셀 크기
    Grid       [][]int   // 격자 → 파티클 인덱스 매핑
}
```

### 생성자

```go
func NewSimulator(Dt float64, Id []int, Pos, Vel []Vector, Gravity Vector) *Simulator
```

`T`, `Count`, `RegionSize`, `GridSize` 는 0으로 초기화됩니다.

---

## 메서드

### `Step()`

1스텝 Euler 적분 (`v += g*dt`, `x += v_new*dt`).  
균일 외부 중력(`Gravity`)만 적용; P³M/SIDM 힘은 포함되지 않습니다.

```
Count++
T = Count * Dt
v_new = v + Gravity * Dt
x_new = x + v_new * Dt
```

### `SolidBoundary(length float64)`

길이 `length`인 정육면체 박스의 **반사 경계** 처리.  
충돌 시간을 역산해 파티클을 정확히 벽에서 반사합니다 (여러 번 반사 가능).

### `PeriodicBoundary(length float64)`

**주기 경계** 적용. `[-L/2, L/2]` 범위를 유지합니다.

### `MakeGrid()`

`RegionSize` / `GridSize`로 3D 격자를 생성하고 각 셀에 파티클을 할당합니다.  
`GetNearAtoms()` 호출 전에 반드시 실행해야 합니다.

### `GetNearAtoms(atom_index int, is_periodic ...bool) []int`

지정 파티클의 **이웃 격자 셀** 내 파티클 인덱스 반환.  
`is_periodic=true`이면 주기 경계를 고려해 경계 셀의 이웃도 포함합니다.  
자기 자신(`atom_index`)은 결과에서 제외됩니다.

### `PeriodicDisplacement(atom_index, another_atom_index int) Vector`

주기 경계 조건 하에서 두 파티클 사이의 **최소 이미지** 변위 벡터를 반환합니다.

### `Save(directory string)`

현재 스냅샷을 HDF5 파일로 저장합니다.  
파일 경로: `<directory>/snapshot_<Count:010d>.hdf5`

### `Load(filename string)`

HDF5 스냅샷을 읽어 시뮬레이터 상태를 복원합니다.

---

## 패키지 수준 함수

### `Read(filename string)`

```go
func Read(filename string) (dt, t float64, count, N int, gravity Vector, id []int, pos, vel []Vector)
```

HDF5 파일에서 모든 상태를 읽어 반환합니다 (`Load`의 내부 구현).

### `Mod(a, b int) int`

항상 양수인 나머지 연산 (`a % b`, 결과가 음수면 `b`를 더함).

---

## 사용 예시

```go
pos := []atom3D.Vector{{0, 0, 0}, {1, 0, 0}}
vel := []atom3D.Vector{{0.1, 0, 0}, {-0.1, 0, 0}}
sim := atom3D.NewSimulator(0.01, []int{0,1}, pos, vel, atom3D.Vector{0, 0, -9.8})

for i := 0; i < 100; i++ {
    sim.Step()
    sim.SolidBoundary(10.0)
}
sim.Save("snapshots")
```
