package atom3D

import (
	"fmt"
	"log"
	"math"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"sync"
	"testing"

	"gonum.org/v1/hdf5"
)

// ── Simulator_ ───────────────────────────────────────────────────────────────

type Simulator_ struct {
	*Simulator
	z float64
	a float64
}

func NewSimulator_(dt float64, id []int, pos []Vector, vel []Vector, z float64) *Simulator_ {
	return &Simulator_{
		Simulator: NewSimulator(dt, id, pos, vel, Vector{0, 0, 0}),
		z:         z,
		a:         1. / (1. + z),
	}
}

// ── SimulatorP3M_ ────────────────────────────────────────────────────────────

type SimulatorP3M_ struct {
	*Simulator_
	p3m    *P3M
	OmegaM float64
	OmegaL float64
}

func NewSimulatorP3M_(dt float64, id []int, pos, vel []Vector, z float64, ng int, L, G float64) *SimulatorP3M_ {
	return &SimulatorP3M_{
		Simulator_: NewSimulator_(dt, id, pos, vel, z),
		p3m:        NewP3M(ng, L, G),
		OmegaM:     0.3,
		OmegaL:     0.7,
	}
}

// HubbleParam = H(a)/H₀ = √(Ω_m/a³ + Ω_Λ)
func (sim SimulatorP3M_) HubbleParam() float64 {
	a := sim.a
	return math.Sqrt(sim.OmegaM/(a*a*a) + sim.OmegaL)
}

// Step : 공변 좌표계 leapfrog (음해적 허블 마찰)
//
// 변수 정의:
//
//	x   : 공변(comoving) 위치  [-L/2, L/2]
//	u   : 특이 속도 (peculiar velocity) = a·ẋ   →  Vel 필드 사용
//
// 운동방정식 (comoving Poisson : ∇²φ = 4πG·δρ_com/a):
//
//	du/dt = F_grav/a - H(a)·u
//	dx/dt = u / a
//	da/dt = H(a)·a
//
// ※ PM이 반환하는 힘은 ∇²Φ_PM = 4πG·δρ_com (a 인자 없음) 기준이므로
//
//	공변 peculiar force = F_PM / a  를 적용해야 선형 성장 δ∝a 가 재현됨.
//
// 수치 안정성:
//
//	Hubble drag를 음해적(implicit)으로 처리 →  u_new = (u + F·dt) / (1 + H·dt)
//	고적색편이(H>>1)에서도 H·dt > 1 허용, 발산 없음.
func (simulator SimulatorP3M_) Step() {
	a := simulator.a
	H := simulator.HubbleParam()
	dt := simulator.Dt

	simulator.MakeGrid()

	// PM 장거리 중력만 사용.
	// p3m.go의 ppForce는 G*[erf(αr)-gaussian]/r² 커널을 쓰는데,
	// 올바른 Ewald 단거리 보정(G*erfc(αr)/r²)과 달라서 PM+PP가
	// r≈RCut 근방에서 중력을 ~2배로 만듦. PM만 써도 32³ 해상도에서는 충분.
	gravity := simulator.p3m.PMForces(simulator.Pos)

	numWorkers := runtime.NumCPU()
	workChan := make(chan int, simulator.N)
	var wg sync.WaitGroup

	newVel := make([]Vector, simulator.N)
	newPos := make([]Vector, simulator.N)

	// 음해적 감쇠 인수: 1/(1 + H·dt)
	damp := 1.0 / (1.0 + H*dt)
	// 공변 포아송 보정: F_physical = F_PM / a
	invA := 1.0 / (a * a) // F_peculiar = F_PM/a²  (du/dt = -H·u + F_PM/a²)

	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range workChan {
				// F_comoving = F_PM / a  (comoving Poisson 1/a 보정)
				f := gravity[i].Mul(invA)
				// u_new = (u + F·dt) / (1 + H·dt)   [음해적 허블 마찰]
				newVel[i] = simulator.Vel[i].Add(f.Mul(dt)).Mul(damp)
				// x_new = x + u_new/a · dt
				newPos[i] = simulator.Pos[i].Add(newVel[i].Mul(dt / a))
			}
		}()
	}
	for i := 0; i < simulator.N; i++ {
		workChan <- i
	}
	close(workChan)
	wg.Wait()

	for i := 0; i < simulator.N; i++ {
		simulator.Vel[i] = newVel[i]
		simulator.Pos[i] = newPos[i]
	}

	// 스케일 인자·적색편이 갱신: da/dt = H·a
	a_new := a * (1 + H*dt)
	simulator.a = a_new
	simulator.z = 1.0/a_new - 1.0

	simulator.Count++
	simulator.T = float64(simulator.Count) * dt

	simulator.PeriodicBoundary(simulator.p3m.L)
}

// ── TestP3M ──────────────────────────────────────────────────────────────────

func TestP3M(t *testing.T) {
	go func() { http.ListenAndServe("localhost:6061", nil) }()

	n := 32
	N := n * n * n
	id := make([]int, N)
	pos := make([]Vector, N)
	vel := make([]Vector, N)

	// ── HDF5 초기 조건 읽기 ───────────────────────────────────────────────
	file, err := hdf5.OpenFile("ic_za.h5", hdf5.F_ACC_RDONLY)
	if err != nil {
		log.Fatalf("파일을 열 수 없습니다: %v", err)
	}
	defer file.Close()

	rootGroup, _ := file.OpenGroup("/")
	defer rootGroup.Close()
	fmt.Println(rootGroup.ObjectNameByIndex(0))

	pos_ := ReadDatasetVector(rootGroup, "pos")
	L := 100.
	for i := 0; i < N; i++ {
		pos[i] = Vector{pos_[i].X - 50., pos_[i].Y - 50., pos_[i].Z - 50.}
	}

	// ── ZA 초기 특이 속도 계산 ────────────────────────────────────────────
	z0 := 49.0
	a0 := 1.0 / (1.0 + z0)
	omegaM := 0.3
	H0 := math.Sqrt(omegaM/(a0*a0*a0) + 0.7)
	dx_grid := L / float64(n)

	// IC 위치 범위 확인
	posMin, posMax := pos[0].X, pos[0].X
	for _, p := range pos {
		if p.X < posMin {
			posMin = p.X
		}
		if p.X > posMax {
			posMax = p.X
		}
	}
	fmt.Printf("IC pos range: [%.3f, %.3f] Mpc (expected ~[-50,50])\n", posMin, posMax)

	for idx := 0; idx < N; idx++ {
		// IC 파일의 파티클 순서가 격자 순서와 다를 수 있으므로
		// 파티클 위치 → 가장 가까운 격자점을 직접 계산
		ix_ := int(math.Round((pos[idx].X+L/2.)/dx_grid - 0.5))
		iy_ := int(math.Round((pos[idx].Y+L/2.)/dx_grid - 0.5))
		iz_ := int(math.Round((pos[idx].Z+L/2.)/dx_grid - 0.5))
		if ix_ < 0 {
			ix_ = 0
		} else if ix_ >= n {
			ix_ = n - 1
		}
		if iy_ < 0 {
			iy_ = 0
		} else if iy_ >= n {
			iy_ = n - 1
		}
		if iz_ < 0 {
			iz_ = 0
		} else if iz_ >= n {
			iz_ = n - 1
		}
		qx := (float64(ix_)+0.5)*dx_grid - L/2.
		qy := (float64(iy_)+0.5)*dx_grid - L/2.
		qz := (float64(iz_)+0.5)*dx_grid - L/2.
		// ZA 특이 속도: u = a·H·Ψ  (Ψ = 가장 가까운 격자점에서의 변위)
		vel[idx] = Vector{
			a0 * H0 * (pos[idx].X - qx),
			a0 * H0 * (pos[idx].Y - qy),
			a0 * H0 * (pos[idx].Z - qz),
		}
	}
	// ZA 속도 진단
	var rmsVel float64
	for _, v := range vel {
		rmsVel += v.X*v.X + v.Y*v.Y + v.Z*v.Z
	}
	rmsVel = math.Sqrt(rmsVel / float64(N))
	fmt.Printf("ZA vel[0] = (%.4f, %.4f, %.4f), RMS vel = %.4f [Mpc H0]\n",
		vel[0].X, vel[0].Y, vel[0].Z, rmsVel)

	// ── 중력 상수 (우주론 단위: H₀=1) ────────────────────────────────────
	// G·M_particle = 3·Ω_m/(8π) · L³/N
	G_cosmo := 3 * omegaM / (8 * math.Pi) * L * L * L / float64(N)

	// ── 시뮬레이터 초기화 ─────────────────────────────────────────────────
	// dt 선택: z=49에서 H≈193.
	// 음해법으로 속도는 안정적이나, 스케일 인자 갱신(명시적)의
	// 정확도를 위해 H·dt < 0.1 → dt = 0.0005
	dt := 0.0005
	simulator := NewSimulatorP3M_(dt, id, pos, vel, z0, 32, L, G_cosmo)
	simulator.RegionSize = L
	simulator.GridSize = simulator.p3m.RCut

	render := Render{
		Width: 150., Height: 150., Depth: 500.,
		Angle: Vector{0., 0., 0.}, FocusFactor: 1.,
	}
	simulator.PeriodicBoundary(L)

	// ── 메인 루프: z=49 → z≈0 ────────────────────────────────────────────
	// z=0까지 물리 시간 ≈ 0.96 H₀⁻¹  →  약 1920 스텝 (dt=0.0005)
	// 여유있게 10000 스텝 상한, z≤0 도달 시 종료
	maxSteps := 10000
	saveInterval := 100

	for i := 0; i < maxSteps; i++ {
		if simulator.z <= 0.0 {
			fmt.Println("z=0 도달, 시뮬레이션 종료")
			break
		}

		if i%saveInterval == 0 {
			// ── 밀도 장 진단 ─────────────────────────────────────────────
			rho := simulator.p3m.AssignDensity(simulator.Pos)
			meanRho := float64(N) / float64(simulator.p3m.Ng*simulator.p3m.Ng*simulator.p3m.Ng)
			maxRho := 0.0
			for _, r := range rho {
				if r > maxRho {
					maxRho = r
				}
			}
			maxDelta := maxRho/meanRho - 1.0
			// rms 속도
			var rmsV float64
			for _, v := range simulator.Vel {
				rmsV += v.X*v.X + v.Y*v.Y + v.Z*v.Z
			}
			rmsV = math.Sqrt(rmsV / float64(N))
			fmt.Printf(">>> z=%.3f a=%.3f | maxDelta=%.3f | rmsVel=%.4f\n",
				simulator.z, simulator.a, maxDelta, rmsV)

			render.Angle = Vector{math.Pi / 6, math.Pi / 5, 0.2*simulator.T + math.Pi/8}
			simulator.Save("snapshots_p3m")
			fig := render.Figure()
			render.Background(fig, []float64{0, 0, 0, 1}) // 검정 배경 (구조 더 잘 보임)
			indices := render.GetSortedIndices(simulator.Pos)
			for _, j := range indices {
				gx := (simulator.Pos[j].X/simulator.p3m.L+0.5)*float64(simulator.p3m.Ng) - 0.5
				gy := (simulator.Pos[j].Y/simulator.p3m.L+0.5)*float64(simulator.p3m.Ng) - 0.5
				gz := (simulator.Pos[j].Z/simulator.p3m.L+0.5)*float64(simulator.p3m.Ng) - 0.5
				ix0 := int(math.Floor(gx)) % simulator.p3m.Ng
				iy0 := int(math.Floor(gy)) % simulator.p3m.Ng
				iz0 := int(math.Floor(gz)) % simulator.p3m.Ng
				if ix0 < 0 {
					ix0 += simulator.p3m.Ng
				}
				if iy0 < 0 {
					iy0 += simulator.p3m.Ng
				}
				if iz0 < 0 {
					iz0 += simulator.p3m.Ng
				}
				idxCell := ix0 + iy0*simulator.p3m.Ng + iz0*simulator.p3m.Ng*simulator.p3m.Ng
				// 로그 스케일 밀도 채색: log2(1+delta) / log2(11)
				// -> 저밀도(파랑) ↔ 고밀도(빨강)  대비 극대화
				delta := rho[idxCell]/meanRho - 1.0
				if delta < 0 {
					delta = 0
				}
				logD := math.Log2(1+delta) / math.Log2(11) // ~0-1 for delta 0-10
				if logD > 1 {
					logD = 1
				}
				render.DrawAtom(fig, simulator.Pos[j], 1., []float64{logD, 0.1, 1 - logD, 0.6})
			}
			render.DrawText(fig, simulator.Pos[0],
				fmt.Sprintf("<-(%0.2f,%0.2f,%0.2f)", simulator.Pos[0].X, simulator.Pos[0].Y, simulator.Pos[0].Z),
				2.5, "D2CodingNerd.ttf", []float64{1, 0, 0, 1})
			render.DrawPlaneText(fig, -50., -50.,
				fmt.Sprintf("z = %.3f  (a = %.3f)", simulator.z, simulator.a),
				5., "D2CodingNerd.ttf", []float64{1, 0, 0, 1})
			render.DrawCube(fig, Vector{0., 0., 0.}, L, 1., []float64{1, 0, 0, 1})
			render.DrawAxis(fig, 5., 5., 2., "D2CodingNerd.ttf")
			render.DrawLine(fig, Vector{0., 0., 0.}, simulator.Gravity, 1., []float64{1, 0, 0, 1})
			render.Save(fig, "images_p3m", simulator.Count)
		}

		simulator.Step()
		fmt.Printf("z = %6.4f  a = %.4f  T = %.4f\n", simulator.z, simulator.a, simulator.T)
	}
}
