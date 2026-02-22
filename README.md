# go-atom3D

**go-atom3D**는 Go로 작성된 3D N체 물리 시뮬레이션 라이브러리입니다.  
범용 입자 시뮬레이터 코어에 P³M(Particle-Particle Particle-Mesh) 중력 솔버, 3D 렌더러, HDF5 I/O를 통합합니다.

---

## 주요 기능

| 기능 | 파일 |
|---|---|
| N체 시뮬레이터 (반사/주기 경계, 그리드 이웃 탐색) | `atom3D.go` |
| P³M 장거리 중력 (FFT + Ewald 단거리 보정) | `p3m.go` |
| 3D 투시 렌더링 (PNG 출력) | `render.go` |
| HDF5 스냅샷 저장/읽기 | `hdf5tools.go` |
| 3D 벡터·텐서 수학 | `vectortools3D.go` |
| 우주론 N체 시뮬레이션 테스트 (z=49→0) | `p3m_test.go` |

---

## 의존성

```
gonum.org/v1/hdf5          - HDF5 I/O (CGO 필요)
gonum.org/v1/gonum/dsp/fourier - 1D FFT
github.com/fogleman/gg     - 2D 그래픽 렌더링
github.com/flopp/go-findfont - 시스템 폰트 탐색
```

> HDF5 라이브러리(libhdf5)가 시스템에 설치되어 있어야 합니다.

---

## 빠른 시작

```go
import "github.com/junobonnie/go-atom3D"

pos := []atom3D.Vector{{0, 0, 0}, {5, 0, 0}}
vel := []atom3D.Vector{{0.1, 0, 0}, {-0.1, 0, 0}}
sim := atom3D.NewSimulator(0.01, []int{0, 1}, pos, vel, atom3D.Vector{0, 0, 0})

for i := 0; i < 1000; i++ {
    sim.Step()
    sim.PeriodicBoundary(10.0)
}
sim.Save("snapshots")
```

### P³M 중력 사용

```go
p3m := atom3D.NewP3M(32, 100.0, G)

sim.RegionSize = 100.0
sim.GridSize   = p3m.RCut
sim.MakeGrid()

forces := p3m.ComputeForces(sim)
```

---

## 우주론 시뮬레이션 실행

```bash
go test -v -run TestP3M -timeout 0
```

`ic_za.h5` 초기 조건 파일이 프로젝트 루트에 필요합니다.  
시뮬레이션 결과는 `snapshots_p3m/` (HDF5)와 `images_p3m/` (PNG)에 저장됩니다.

---

## 문서

| 파일 | 설명 |
|---|---|
| [atom3D.md](docs/atom3D.md) | 핵심 `Simulator` 구조체 및 메서드 |
| [p3m.md](docs/p3m.md) | P³M 중력 솔버 이론 및 API |
| [p3m_test.md](docs/p3m_test.md) | 우주론 N체 테스트 (`SimulatorP3M_`) |
| [render.md](docs/render.md) | 3D 렌더러 API |
| [hdf5tools.md](docs/hdf5tools.md) | HDF5 I/O 헬퍼 함수 |
| [vectortools3D.md](docs/vectortools3D.md) | 3D 벡터·텐서 수학 |

---

## 프로젝트 구조

```
go-atom3D/
├── atom3D.go           # 핵심 Simulator 구조체
├── p3m.go              # P³M 중력 솔버
├── p3m_test.go         # 우주론 N체 시뮬레이션 테스트
├── render.go           # 3D 소프트웨어 렌더러
├── hdf5tools.go        # HDF5 I/O 유틸리티
├── vectortools3D.go    # Vector / Tensor 수학
├── ic_za.h5            # Zel'dovich 근사 초기 조건
├── docs/               # 문서
│   ├── atom3D.md
│   ├── p3m.md
│   ├── p3m_test.md
│   ├── render.md
│   ├── hdf5tools.md
│   └── vectortools3D.md
├── snapshots_p3m/      # HDF5 스냅샷 출력
└── images_p3m/         # PNG 렌더 이미지 출력
```

---

## 물리 모델 개요

```
F_total = F_PM (FFT 장거리) + F_PP (Ewald 단거리 보정)

운동방정식 (공변 좌표, 음해적 leapfrog):
  du/dt = F_PM/a² - H(a)·u
  dx/dt = u/a
  da/dt = H(a)·a

H(a)/H₀ = √(Ω_m/a³ + Ω_Λ)
```

> 참고: Hockney & Eastwood, *Computer Simulation Using Particles*, 1988.
