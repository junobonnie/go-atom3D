package atom3D

import (
	"math"
	"sync"

	"gonum.org/v1/gonum/dsp/fourier"
)

// P3M implements Particle-Particle Particle-Mesh (P³M) gravity.
//
// 주기 경계 조건 하에서 무한 범위 중력(1/r²)을 Ewald 분리로 계산합니다:
//
//	F_total = F_PM (장거리, FFT k-공간) + F_PP (단거리 보정, 실공간 직접합)
//
// PM: Ewald Green 함수  Φ̃(k) = -4πG · exp(-k²/4α²) / k²
// PP: 보정 커널        F = G · [erf(αr) - (2αr/√π)·e^{-α²r²}] / r³ · d
//
// 참고: Hockney & Eastwood, "Computer Simulation Using Particles", 1988.
type P3M struct {
	Ng    int     // PM 격자 크기 (차원당, 2의 거듭제곱 권장)
	L     float64 // 주기 박스 크기
	G     float64 // 중력 상수
	Alpha float64 // Ewald 분리 파라미터 (단위: 1/length)
	RCut  float64 // PP 컷오프 반경 (격자 간격의 약 2.5배)
}

// NewP3M은 자동 파라미터 선택으로 P3M 솔버를 생성합니다.
//
//	ng : PM 격자 해상도 (차원당 셀 수, 2의 거듭제공 권장)
//	L  : 주기 박스 크기
//	G  : 중력 상수 (단위계에 맞게 설정)
//
// 주의: PP 탐색에 sim.GetNearAtoms를 이용하므로
//
//	sim.GridSize <= P3M.RCut 를 만족해야 이웃이 누락되지 않습니다.
func NewP3M(ng int, L, G float64) *P3M {
	dx := L / float64(ng)
	rCut := 2.5 * dx    // PP 컷오프 ≈ 격자 간격 × 2.5
	alpha := 3.0 / rCut // erfc(alpha·rCut) ≈ 0.001
	return &P3M{
		Ng:    ng,
		L:     L,
		G:     G,
		Alpha: alpha,
		RCut:  rCut,
	}
}

// ── 내부 헬퍼 ────────────────────────────────────────────────────────────────

// wrap3D는 주기 경계를 포함한 3D 인덱스를 1D 인덱스로 변환합니다.
func (p *P3M) wrap3D(ix, iy, iz int) int {
	ng := p.Ng
	ix = ((ix % ng) + ng) % ng
	iy = ((iy % ng) + ng) % ng
	iz = ((iz % ng) + ng) % ng
	return ix + iy*ng + iz*ng*ng
}

// cicW는 CIC 보간 가중치를 반환합니다. (d=0: 현재 셀, d=1: 다음 셀)
func cicW(t float64, d int) float64 {
	if d == 0 {
		return 1.0 - t
	}
	return t
}

// psinc는 CIC 창함수 sinc(x) = sin(x)/x (x=0이면 1.0)를 반환합니다.
func psinc(x float64) float64 {
	if math.Abs(x) < 1e-10 {
		return 1.0
	}
	return math.Sin(x) / x
}

// ── 3D FFT ───────────────────────────────────────────────────────────────────

// fft3D는 ng×ng×ng 복소 배열에 대해 3D FFT를 수행합니다 (in-place).
// 인덱스 규칙: data[ix + iy*ng + iz*ng²]
// 역변환(inverse=true) 후에는 ng³으로 나눠야 정규화됩니다.
func fft3D(data []complex128, ng int, inverse bool) {
	f := fourier.NewCmplxFFT(ng)
	row := make([]complex128, ng)

	doFFT := func(buf []complex128) {
		if inverse {
			f.Sequence(buf, buf) // 역FFT (unnormalized, 별도 1/N 정규화 필요)
		} else {
			f.Coefficients(buf, buf) // 순방향 FFT
		}
	}

	// x 방향 (stride = 1)
	for iz := 0; iz < ng; iz++ {
		for iy := 0; iy < ng; iy++ {
			base := iy*ng + iz*ng*ng
			copy(row, data[base:base+ng])
			doFFT(row)
			copy(data[base:base+ng], row)
		}
	}

	// y 방향 (stride = ng)
	for iz := 0; iz < ng; iz++ {
		for ix := 0; ix < ng; ix++ {
			for iy := 0; iy < ng; iy++ {
				row[iy] = data[ix+iy*ng+iz*ng*ng]
			}
			doFFT(row)
			for iy := 0; iy < ng; iy++ {
				data[ix+iy*ng+iz*ng*ng] = row[iy]
			}
		}
	}

	// z 방향 (stride = ng²)
	for iy := 0; iy < ng; iy++ {
		for ix := 0; ix < ng; ix++ {
			for iz := 0; iz < ng; iz++ {
				row[iz] = data[ix+iy*ng+iz*ng*ng]
			}
			doFFT(row)
			for iz := 0; iz < ng; iz++ {
				data[ix+iy*ng+iz*ng*ng] = row[iz]
			}
		}
	}
}

// ── CIC 밀도 할당 ────────────────────────────────────────────────────────────

// AssignDensity는 파티클 위치를 CIC(Cloud-In-Cell) 방식으로 격자 밀도장에 사상합니다.
// 반환값: ρ(ix,iy,iz) [파티클/셀]
func (p *P3M) AssignDensity(pos []Vector) []float64 {
	ng := p.Ng
	rho := make([]float64, ng*ng*ng)

	for _, r := range pos {
		// 격자 좌표 [0, ng): 셀 중심이 정수에 오도록 -0.5 이동
		gx := (r.X/p.L+0.5)*float64(ng) - 0.5
		gy := (r.Y/p.L+0.5)*float64(ng) - 0.5
		gz := (r.Z/p.L+0.5)*float64(ng) - 0.5

		ix0 := int(math.Floor(gx))
		iy0 := int(math.Floor(gy))
		iz0 := int(math.Floor(gz))

		tx := gx - float64(ix0) // 셀 내 소수 위치 [0, 1)
		ty := gy - float64(iy0)
		tz := gz - float64(iz0)

		for di := 0; di <= 1; di++ {
			for dj := 0; dj <= 1; dj++ {
				for dk := 0; dk <= 1; dk++ {
					w := cicW(tx, di) * cicW(ty, dj) * cicW(tz, dk)
					rho[p.wrap3D(ix0+di, iy0+dj, iz0+dk)] += w
				}
			}
		}
	}
	return rho
}

// ── 포아송 방정식 풀기 ───────────────────────────────────────────────────────

// SolvePotential은 밀도장 ρ로부터 중력 포텐셜 Φ를 계산합니다.
// k-공간에서 Ewald Green 함수를 곱한 뒤 역FFT합니다.
// CIC 창함수 디콘볼루션(역보간 포함 2회)을 적용합니다.
func (p *P3M) SolvePotential(rho []float64) []float64 {
	ng := p.Ng
	size := ng * ng * ng
	dk := 2 * math.Pi / p.L // k-공간 격자 간격
	dx := p.L / float64(ng) // PM 셀 크기

	// ρ → 복소수 배열
	data := make([]complex128, size)
	for i, v := range rho {
		data[i] = complex(v, 0)
	}

	// 순방향 FFT
	fft3D(data, ng, false)

	// k-공간에서 Ewald Green 함수 곱셈
	for iz := 0; iz < ng; iz++ {
		nz := iz
		if nz > ng/2 {
			nz -= ng
		}
		kz := float64(nz) * dk

		for iy := 0; iy < ng; iy++ {
			ny := iy
			if ny > ng/2 {
				ny -= ng
			}
			ky := float64(ny) * dk

			for ix := 0; ix < ng; ix++ {
				nx := ix
				if nx > ng/2 {
					nx -= ng
				}
				kx := float64(nx) * dk

				k2 := kx*kx + ky*ky + kz*kz
				idx := ix + iy*ng + iz*ng*ng

				if k2 == 0 {
					data[idx] = 0 // k=0: 평균 포텐셜 = 0 (주기 박스 조건)
					continue
				}

				// Ewald Green 함수: -4πG·(Ng/L)³·exp(-k²/4α²)/k²
				// AssignDensity는 particles/cell 단위 → 물리 질량밀도 변환: ×Ng³/L³
				volumeFactor := float64(ng*ng*ng) / (p.L * p.L * p.L)
				green := -4 * math.Pi * p.G * volumeFactor * math.Exp(-k2/(4*p.Alpha*p.Alpha)) / k2

				// CIC 창함수 디콘볼루션 (할당 + 읽기 2회 보정)
				wx := psinc(kx * dx / 2.0)
				wy := psinc(ky * dx / 2.0)
				wz := psinc(kz * dx / 2.0)
				w2 := (wx * wy * wz) * (wx * wy * wz)
				if w2 < 1e-10 {
					w2 = 1e-10
				}

				data[idx] = complex(green/w2, 0) * data[idx]
			}
		}
	}

	// 역FFT
	fft3D(data, ng, true)

	// 정규화 (gonum IFFT는 1/N 정규화를 하지 않음)
	scale := 1.0 / float64(size)
	phi := make([]float64, size)
	for i, v := range data {
		phi[i] = real(v) * scale
	}
	return phi
}

// ── PM 힘 계산 ───────────────────────────────────────────────────────────────

// PMForces는 PM(장거리) 중력 가속도를 각 파티클에 대해 계산합니다.
// CIC 밀도 할당 → 포아송 방정식(FFT) → 기울기(유한차분) → CIC 역보간 순서로 진행합니다.
func (p *P3M) PMForces(pos []Vector) []Vector {
	ng := p.Ng
	dx := p.L / float64(ng)
	N := len(pos)

	// 1. 밀도 할당
	rho := p.AssignDensity(pos)

	// 2. 포텐셜 계산
	phi := p.SolvePotential(rho)

	// 3. 격자에서 F = -∇Φ (중앙 유한차분)
	fxG := make([]float64, ng*ng*ng)
	fyG := make([]float64, ng*ng*ng)
	fzG := make([]float64, ng*ng*ng)

	for iz := 0; iz < ng; iz++ {
		for iy := 0; iy < ng; iy++ {
			for ix := 0; ix < ng; ix++ {
				i := ix + iy*ng + iz*ng*ng
				fxG[i] = -(phi[p.wrap3D(ix+1, iy, iz)] - phi[p.wrap3D(ix-1, iy, iz)]) / (2 * dx)
				fyG[i] = -(phi[p.wrap3D(ix, iy+1, iz)] - phi[p.wrap3D(ix, iy-1, iz)]) / (2 * dx)
				fzG[i] = -(phi[p.wrap3D(ix, iy, iz+1)] - phi[p.wrap3D(ix, iy, iz-1)]) / (2 * dx)
			}
		}
	}

	// 4. 격자 힘 → 파티클으로 CIC 역보간
	forces := make([]Vector, N)
	for pi, r := range pos {
		gx := (r.X/p.L+0.5)*float64(ng) - 0.5
		gy := (r.Y/p.L+0.5)*float64(ng) - 0.5
		gz := (r.Z/p.L+0.5)*float64(ng) - 0.5

		ix0 := int(math.Floor(gx))
		iy0 := int(math.Floor(gy))
		iz0 := int(math.Floor(gz))

		tx := gx - float64(ix0)
		ty := gy - float64(iy0)
		tz := gz - float64(iz0)

		var f Vector
		for di := 0; di <= 1; di++ {
			for dj := 0; dj <= 1; dj++ {
				for dk := 0; dk <= 1; dk++ {
					w := cicW(tx, di) * cicW(ty, dj) * cicW(tz, dk)
					i := p.wrap3D(ix0+di, iy0+dj, iz0+dk)
					f.X += w * fxG[i]
					f.Y += w * fyG[i]
					f.Z += w * fzG[i]
				}
			}
		}
		forces[pi] = f
	}
	return forces
}

// ── PP 단거리 보정 ───────────────────────────────────────────────────────────

// ppForce는 파티클 j가 i에 미치는 단거리 Ewald 보정 힘을 반환합니다.
//
//	d = r_j - r_i,  r = |d|
//	F = G · [erf(αr) - (2αr/√π)·exp(-α²r²)] / r³ · d
//
// r→0 극한에서 분자가 0에 수렴하므로 특이점이 없습니다.
func (p *P3M) ppForce(d Vector, r float64) Vector {
	ar := p.Alpha * r
	bracket := math.Erf(ar) - 2*ar/math.Sqrt(math.Pi)*math.Exp(-ar*ar)
	factor := p.G * bracket / (r * r * r)
	return d.Mul(factor)
}

// PPCorrections는 각 파티클의 단거리 PP 보정 가속도를 병렬로 계산합니다.
//
// 사전 조건: sim.MakeGrid()가 호출된 상태여야 합니다.
// 사전 조건: sim.GridSize <= p.RCut (이웃 누락 방지)
func (p *P3M) PPCorrections(sim *Simulator) []Vector {
	N := sim.N
	corrections := make([]Vector, N)

	numWorkers := 4
	workChan := make(chan int, N)
	var wg sync.WaitGroup

	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range workChan {
				var corr Vector
				for _, j := range sim.GetNearAtoms(i, true) {
					if j == i {
						continue
					}
					d := sim.PeriodicDisplacement(i, j)
					r := d.Abs()
					if r > 0 && r < p.RCut {
						corr = corr.Add(p.ppForce(d, r))
					}
				}
				corrections[i] = corr
			}
		}()
	}

	for i := 0; i < N; i++ {
		workChan <- i
	}
	close(workChan)
	wg.Wait()

	return corrections
}

// ── 메인 인터페이스 ──────────────────────────────────────────────────────────

// ComputeForces는 각 파티클에 작용하는 P³M 중력 가속도를 반환합니다.
//
// 사용 예시:
//
//	p3m := NewP3M(32, 100., 1.0)
//	sim.MakeGrid()
//	forces := p3m.ComputeForces(sim)
//
// 주의: 호출 전에 반드시 sim.MakeGrid()를 실행하세요.
func (p *P3M) ComputeForces(sim *Simulator) []Vector {
	pmF := p.PMForces(sim.Pos)
	ppF := p.PPCorrections(sim)

	total := make([]Vector, sim.N)
	for i := 0; i < sim.N; i++ {
		total[i] = pmF[i].Add(ppF[i])
	}
	return total
}
