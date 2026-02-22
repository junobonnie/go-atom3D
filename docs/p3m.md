# p3m.go — P³M 중력 솔버

Particle-Particle Particle-Mesh (P³M) 방식으로 주기 경계 하의 장거리 중력을 효율적으로 계산하는 솔버입니다.

---

## 이론적 배경

중력 포텐셜 ∇²Φ = 4πGρ 를 Ewald 분리법으로 두 부분으로 나눕니다:

| 구성 요소 | 방법 | 범위 |
|---|---|---|
| **PM (장거리)** | FFT + Ewald Green 함수 | 전체 박스 |
| **PP (단거리 보정)** | 실공간 직접합 | `r < RCut` |

```
F_total = F_PM + F_PP
```

**PM Green 함수:**  
`Φ̃(k) = -4πG · exp(-k²/4α²) / k²`

**PP 보정 힘:**  
`F = G · [erf(αr) - (2αr/√π)·e^{-α²r²}] / r³ · d`

> 참고: Hockney & Eastwood, *Computer Simulation Using Particles*, 1988.

---

## `P3M` 구조체

```go
type P3M struct {
    Ng    int     // PM 격자 해상도 (차원당 셀 수, 2의 거듭제곱 권장)
    L     float64 // 주기 박스 크기
    G     float64 // 중력 상수
    Alpha float64 // Ewald 분리 파라미터 [1/length]
    RCut  float64 // PP 컷오프 반경 ≈ 2.5 × (L/Ng)
}
```

### 생성자

```go
func NewP3M(ng int, L, G float64) *P3M
```

자동으로 `RCut = 2.5 × dx`, `Alpha = 3.0 / RCut` 을 설정합니다.  
`sim.GridSize <= RCut` 조건 만족 시 이웃 탐색 누락이 없습니다.

---

## 공개 메서드

### `AssignDensity(pos []Vector) []float64`

**CIC(Cloud-In-Cell)** 보간으로 파티클 위치를 격자 밀도장 ρ[Ng³] 에 사상합니다.

- 좌표 변환: `gx = (x/L + 0.5)*Ng - 0.5`
- 각 파티클은 인접 2³=8 셀에 가중치 `w = (1-tx)(1-ty)(1-tz)` 등으로 분산.

### `SolvePotential(rho []float64) []float64`

밀도장 → 포텐셜 Φ 계산:
1. ρ → 복소 배열 변환
2. 3D FFT
3. k-공간에서 Ewald Green 함수 × CIC 창함수 역보정 곱셈
4. 역FFT + 정규화(1/Ng³)

`k=0` 모드는 0으로 설정 (중력 포텐셜의 기준값 = 0).

### `PMForces(pos []Vector) []Vector`

장거리 PM 가속도 계산 파이프라인:
1. `AssignDensity` → ρ
2. `SolvePotential` → Φ
3. 중앙 유한차분 `-∇Φ` → 격자 힘 (fxG, fyG, fzG)
4. CIC 역보간 → 파티클 힘

### `PPCorrections(sim *Simulator) []Vector`

단거리 Ewald 보정 힘 (`r < RCut` 내 직접합):

- **4 Worker goroutine** 병렬 처리 (`sync.WaitGroup`)
- `sim.GetNearAtoms(i, true)` 로 후보 이웃 탐색
- 보정 커널: `ppForce(d, r)` (아래 참조)

> **사전 조건**: `sim.MakeGrid()` 호출 필수, `sim.GridSize <= RCut`

### `ComputeForces(sim *Simulator) []Vector`

```go
total[i] = PMForces[i] + PPCorrections[i]
```

P³M 전체 힘 = PM 장거리 + PP 단거리 보정.

---

## 내부 함수

| 함수 | 설명 |
|---|---|
| `wrap3D(ix, iy, iz int)` | 주기 경계 적용 3D → 1D 인덱스 변환 |
| `cicW(t float64, d int)` | CIC 가중치: d=0 → `1-t`, d=1 → `t` |
| `psinc(x float64)` | `sin(x)/x` (x≈0이면 1.0) |
| `fft3D(data, ng, inverse)` | x→y→z 방향 순차 1D FFT로 3D FFT 구현 |
| `ppForce(d Vector, r float64)` | Ewald 단거리 보정 힘 벡터 |

---

## 사용 예시

```go
p3m := atom3D.NewP3M(32, 100.0, G_cosmo)

sim.RegionSize = 100.0
sim.GridSize   = p3m.RCut   // 이웃 탐색 셀 크기 ≤ RCut

sim.MakeGrid()
forces := p3m.ComputeForces(sim)  // []Vector, 각 파티클 가속도
```
