# p3m_test.go — 우주론 N체 시뮬레이션 테스트

`TestP3M` 함수를 통해 P³M 솔버를 이용한 **대규모 구조 형성** 시뮬레이션을 실행합니다.  
적색편이 z=49에서 z=0까지의 우주 진화를 재현합니다.

---

## 추가 구조체

### `Simulator_`

```go
type Simulator_ struct {
    *Simulator        // atom3D.Simulator 임베딩
    z float64         // 현재 적색편이
    a float64         // 스케일 인자 a = 1/(1+z)
}
```

기본 시뮬레이터에 우주론적 팽창 변수를 추가한 확장형.

### `SimulatorP3M_`

```go
type SimulatorP3M_ struct {
    *Simulator_
    p3m    *P3M
    OmegaM float64  // 물질 밀도 파라미터 (default 0.3)
    OmegaL float64  // 암흑에너지 밀도 파라미터 (default 0.7)
}
```

P³M 솔버를 내장한 우주론 시뮬레이터.

---

## `SimulatorP3M_.HubbleParam() float64`

현재 스케일 인자 `a`에서의 무차원 허블 파라미터:

```
H(a) / H₀ = √(Ω_m / a³ + Ω_Λ)
```

---

## `SimulatorP3M_.Step()`

공변(comoving) 좌표계 **음해적(implicit) leapfrog** 적분:

| 변수 | 의미 |
|---|---|
| `x` | 공변 위치 `[-L/2, L/2]` |
| `u = a·ẋ` | 특이 속도 (peculiar velocity) |

**운동 방정식:**
```
du/dt = F_PM / a² - H(a) · u
dx/dt = u / a
da/dt = H(a) · a
```

**수치 구현 (음해적 Hubble 감쇠):**
```
damp    = 1 / (1 + H·dt)
invA    = 1 / a²
f       = F_PM × invA                    // 공변 힘
u_new   = (u + f·dt) × damp             // 음해적 감쇠
x_new   = x + u_new/a · dt
a_new   = a × (1 + H·dt)
z_new   = 1/a_new - 1
```

> **PM만 사용**: `PMForces`만 적용 (PP 보정 없음).  
> 현재 PP 커널이 올바른 `erfc` 기반 Ewald 보정과 달라 `r≈RCut` 근방에서 힘이 ~2배가 되는 문제가 있어 PM 단독 사용 중.

---

## `TestP3M` 시뮬레이션 설정

| 항목 | 값 |
|---|---|
| 파티클 수 | 32³ = 32,768 |
| 박스 크기 | L = 100 Mpc |
| 초기 적색편이 | z₀ = 49 |
| 목표 적색편이 | z = 0 |
| Ω_m | 0.3 |
| Ω_Λ | 0.7 |
| 타임스텝 dt | 0.0005 H₀⁻¹ |
| PM 격자 | 32³ |
| 최대 스텝 수 | 10,000 |
| 저장 간격 | 100 스텝마다 |

### 초기 조건 (IC)

`ic_za.h5` 파일에서 Zel'dovich 근사(ZA) 변위장을 읽어 초기 위치와 속도를 설정합니다:

```
u[i] = a₀ · H₀ · Ψ[i]
```

여기서 Ψ는 가장 가까운 격자점에서의 변위.

### 중력 상수 (우주론 단위, H₀=1)

```
G · M_particle = (3 · Ω_m) / (8π) × L³/N
```

### 시각화

100 스텝마다:
- `snapshots_p3m/` 에 HDF5 스냅샷 저장
- `images_p3m/` 에 PNG 렌더 이미지 저장
- 파티클 색상: **밀도 대비 과밀도** δ 기반 로그 스케일 (저밀도 파랑 ↔ 고밀도 빨강)

---

## 실행

```bash
go test -v -run TestP3M -timeout 0
```

> 프로파일링 서버 (`pprof`)가 `localhost:6061`에서 자동 실행됩니다.
