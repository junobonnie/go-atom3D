package atom3D

import (
	"fmt"
	"math/rand"
	"testing"
	"net/http"
	_ "net/http/pprof"
)

type Simulator_ struct {
	*Simulator
	Density []float64
	Velocity [][]float64
	Rho0 float64
	FlipRatio float64
	O float64
}

func (simulator Simulator_) collision(i, j int) {
	d := simulator.Pos[i].Sub(simulator.Pos[j])
	v1 := simulator.Vel[i]
	v2 := simulator.Vel[j]

	v1_ := d.Mul(v1.Dot(d)).Div(d.Dot(d))
	v2_ := d.Mul(v2.Dot(d)).Div(d.Dot(d))

	if d.Dot(d) < 25. && d.Dot(v1_.Sub(v2_)) < 0. {
		//simulator.Pos[i] = simulator.Pos[i].Sub(v1.Mul(simulator.Dt))
		//simulator.Pos[j] = simulator.Pos[j].Sub(v2.Mul(simulator.Dt))
		simulator.Vel[i] = v1.Sub(v1_).Add(v2_)
		simulator.Vel[j] = v2.Sub(v2_).Add(v1_)
	}
}

func (simulator Simulator_) SetDensity() {
	n := int(simulator.RegionSize/simulator.GridSize+1)
	simulator.Density = make([]float64, n*n*n)
	for i := 0; i < n*n*n; i++ {
		simulator.Density[i] = float64(len(simulator.Grid[i]))
	}
}

func (simulator Simulator_) SetVelocity() {
	n := int(simulator.RegionSize/simulator.GridSize+1)
	simulator.Velocity = make([][]float64, 3)
	for i := 0; i < 3; i++ {
		simulator.Velocity[i] = make([]float64, (n+1)*n*n)
	}
	for i := 0; i < simulator.N; i++ {
		particle_vel := []float64{simulator.Vel[i].X, simulator.Vel[i].Y, simulator.Vel[i].Z}
		for j, factor := range [][]float64{{1, 0., 0.}, {0., 1, 0.}, {0., 0., 1}} {
			X := simulator.Pos[i].X + simulator.RegionSize/2 - factor[0]*simulator.GridSize/2
			Y := simulator.Pos[i].Y + simulator.RegionSize/2 - factor[1]*simulator.GridSize/2
			Z := simulator.Pos[i].Z + simulator.RegionSize/2 - factor[2]*simulator.GridSize/2
			x_cell := int(X / simulator.GridSize)
			y_cell := int(Y / simulator.GridSize)
			z_cell := int(Z / simulator.GridSize)
			dx := X/simulator.GridSize - float64(x_cell)
			dy := Y/simulator.GridSize - float64(y_cell)
			dz := Z/simulator.GridSize - float64(z_cell)
			for i_ := 0; i_ < 2; i_++ {
				for j_ := 0; j_ < 2; j_++ {
					for k_ := 0; k_ < 2; k_++ {
						x_ := x_cell + i_
						y_ := y_cell + j_
						z_ := z_cell + k_

						simulator.Velocity[j][x_+y_*n+z_*n*n] += particle_vel[j]*(dx+float64(i_)*(1-2.*dx))*(dy+float64(j_)*(1-2.*dy))*(dz+float64(k_)*(1-2.*dz))
					}
				}
			}
		}
	}
}

func (simulator Simulator_) AddIncompressForce() {
	simulator.SetDensity()
	n := int(simulator.RegionSize/simulator.GridSize+1)
	for j := 0; j < 20; j++ {
	divergence := make([]float64, n*n*n)
		for i := 0; i < n*n*n; i++ {
			if simulator.Density[i] != 0. {
			divergence[i] = simulator.O*(simulator.Velocity[0][i+1] - simulator.Velocity[0][i] + simulator.Velocity[1][i+n] - simulator.Velocity[1][i] + simulator.Velocity[2][i+n*n] - simulator.Velocity[2][i]) + (simulator.Density[i] - simulator.Rho0)
			}
		}
		for i := 0; i < n*n*n; i++ {
			force := divergence[i]/6
			simulator.Velocity[0][i] += force
			simulator.Velocity[0][i+1] -= force
			simulator.Velocity[1][i] += force
			simulator.Velocity[1][i+n] -= force
			simulator.Velocity[2][i] += force
			simulator.Velocity[2][i+n*n] -= force
		}
	}
}

func (simulator Simulator_) PIC_FLIP() {
	n := int(simulator.RegionSize/simulator.GridSize+1)
	OriginalVelocity := make([][]float64, 3)
	for i := 0; i < 3; i++ {
		OriginalVelocity[i] = make([]float64, (n+1)*n*n)
	}
	copy(OriginalVelocity, simulator.Velocity)

	simulator.AddIncompressForce()

	for i := 0; i < simulator.N; i++ {
		PIC_vel := Vector{0., 0., 0.}
		FLIP_vel := Vector{0., 0., 0.}
		for j, factor := range [][]float64{{1, 0., 0.}, {0., 1, 0.}, {0., 0., 1}} {
			X := simulator.Pos[i].X + simulator.RegionSize/2 - factor[0]*simulator.GridSize/2
			Y := simulator.Pos[i].Y + simulator.RegionSize/2 - factor[1]*simulator.GridSize/2
			Z := simulator.Pos[i].Z + simulator.RegionSize/2 - factor[2]*simulator.GridSize/2
			x_cell := int(X / simulator.GridSize)
			y_cell := int(Y / simulator.GridSize)
			z_cell := int(Z / simulator.GridSize)
			dx := X/simulator.GridSize - float64(x_cell)
			dy := Y/simulator.GridSize - float64(y_cell)
			dz := Z/simulator.GridSize - float64(z_cell)
			for i_ := 0; i_ < 2; i_++ {
				for j_ := 0; j_ < 2; j_++ {
					for k_ := 0; k_ < 2; k_++ {
						x_ := x_cell + i_
						y_ := y_cell + j_
						z_ := z_cell + k_
						if j == 0 {
							PIC_vel.X += simulator.Velocity[0][x_+y_*n+z_*n*n]*(dx+float64(i_)*(1-2.*dx))*(dy+float64(j_)*(1-2.*dy))*(dz+float64(k_)*(1-2.*dz))
							FLIP_vel.X += (simulator.Velocity[0][x_+y_*n+z_*n*n] - OriginalVelocity[0][x_+y_*n+z_*n*n])*(dx+float64(i_)*(1-2.*dx))*(dy+float64(j_)*(1-2.*dy))*(dz+float64(k_)*(1-2.*dz))
						}
						if j == 1 {
							PIC_vel.Y += simulator.Velocity[1][x_+y_*n+z_*n*n]*(dx+float64(i_)*(1-2.*dx))*(dy+float64(j_)*(1-2.*dy))*(dz+float64(k_)*(1-2.*dz))
							FLIP_vel.Y += (simulator.Velocity[1][x_+y_*n+z_*n*n] - OriginalVelocity[1][x_+y_*n+z_*n*n])*(dx+float64(i_)*(1-2.*dx))*(dy+float64(j_)*(1-2.*dy))*(dz+float64(k_)*(1-2.*dz))
						}
						if j == 2 {
							PIC_vel.Z += simulator.Velocity[2][x_+y_*n+z_*n*n]*(dx+float64(i_)*(1-2.*dx))*(dy+float64(j_)*(1-2.*dy))*(dz+float64(k_)*(1-2.*dz))
							FLIP_vel.Z += (simulator.Velocity[2][x_+y_*n+z_*n*n] - OriginalVelocity[2][x_+y_*n+z_*n*n])*(dx+float64(i_)*(1-2.*dx))*(dy+float64(j_)*(1-2.*dy))*(dz+float64(k_)*(1-2.*dz))
						}
					}
				}
			}
		}
		simulator.Vel[i] = PIC_vel.Mul(1.-simulator.FlipRatio).Add(simulator.Vel[i].Add(FLIP_vel).Mul(simulator.FlipRatio))
	}
}

func (simulator Simulator_) Step() {
	simulator.Count++
	simulator.T = float64(simulator.Count) * simulator.Dt

	simulator.MakeGrid()
	simulator.SetVelocity()
	simulator.PIC_FLIP()

	for i := 0; i < simulator.N; i++ {
		for _, j := range simulator.GetNearAtoms(i) {
			if i != j {
				simulator.collision(i, j)
			}
		}
	}

	x_ := make([]Vector, simulator.N)
	v_ := make([]Vector, simulator.N)
	for i := 0; i < simulator.N; i++ {
		new_vel := simulator.Vel[i].Add(simulator.Gravity.Mul(simulator.Dt))
		v_[i] = new_vel
		x_[i] = simulator.Pos[i].Add(new_vel.Mul(simulator.Dt))
	}

	for i := 0; i < simulator.N; i++ {
		simulator.Pos[i] = x_[i]
		simulator.Vel[i] = v_[i]
	}
	
	simulator.SolidBoundary(100.)
}

func InvRenderSO3(angle Vector) Tensor {
	var angle_rotaion Tensor
	if angle.X == 0 && angle.Y == 0 && angle.Z == 0 {
		angle_rotaion = Tensor{
			1, 0, 0,
			0, 1, 0,
			0, 0, 1}
	} else {
		angle_rotaion = SO3_z(-angle.Z).DotT(
			SO3_y(-angle.Y)).DotT(
				SO3_x(-angle.X))
	}
	return angle_rotaion
}

func TestMain(t *testing.T) {
	go func() {
		http.ListenAndServe("localhost:6060", nil)
	}()
	N := 16000
	id := make([]int, N)
	pos := make([]Vector, N)
	vel := make([]Vector, N)
	for i := 0; i < N; i++ {
		x_ := 2.5*float64(i%20) - 48.75 + 0.1*rand.Float64()
		y_ := 2.5*float64((i%800)/20) - 48.75 + 0.1*rand.Float64()
		z_ := 2.5*float64(i/800) + 1.25 + 0.1*rand.Float64()
		pos[i] = Vector{x_, y_, z_}
		vel[i] = Vector{0.0, 0.0, 0.0}
	}
	gravity := Vector{0.0, 0.0, -100.0}

	// Create a new simulator
	n := 23
	Vel := make([][]float64, 3)
	for i := 0; i < 3; i++ {
		Vel[i] = make([]float64, (n+1)*n*n)
	}
	simulator := Simulator_{NewSimulator(0.01, id, pos, vel, gravity), make([]float64, n*n*n), Vel, 8.0, 0.9, 1.9}
	simulator.RegionSize = 110.
	simulator.GridSize = 5.
	render := Render{Width: 150., Height: 150., Depth: 500., Angle: Vector{0.0, 0.0, 0.0}, FocusFactor: 1.}
	simulator.Load("snapshots/snapshot_0000009980.hdf5")
	time := 500
	simulator.MakeGrid()
	for i := 0; i < 10*time; i++ {
		render.Angle = Vector{0.5, 0.2 * simulator.T, 0.5}
		if i%10 == 0 {
			simulator.Save("snapshots")
			fig := render.Figure()
			render.Background(fig, []float64{1, 1, 1, 1})
			indices := render.GetSortedIndices(simulator.Pos)
			for _, j := range indices {
				if j < 8000 {
					render.DrawAtom(fig, simulator.Pos[j], 2.5, []float64{0., 0., 1., 0.2})
				} else {
					render.DrawAtom(fig, simulator.Pos[j], 2.5, []float64{0., 1., 0., 0.2})
				}
			}
			render.DrawText(fig, simulator.Pos[0], fmt.Sprintf("<-(%0.2f,%0.2f, %0.2f)", simulator.Pos[0].X, simulator.Pos[0].Y, simulator.Pos[0].Z), 2.5, "D2CodingNerd.ttf", []float64{1, 0, 0, 1})
			x := int((simulator.Pos[0].X + simulator.RegionSize/2) / simulator.GridSize)
			y := int((simulator.Pos[0].Y + simulator.RegionSize/2) / simulator.GridSize)
			z := int((simulator.Pos[0].Z + simulator.RegionSize/2) / simulator.GridSize)
			x_ := float64(x)*simulator.GridSize - simulator.RegionSize/2 + simulator.GridSize/2
			y_ := float64(y)*simulator.GridSize - simulator.RegionSize/2 + simulator.GridSize/2
			z_ := float64(z)*simulator.GridSize - simulator.RegionSize/2 + simulator.GridSize/2
			render.DrawCube(fig, Vector{x_, y_, z_}, 3.*simulator.GridSize, 1., []float64{0, 1, 0, 1})
			for _, j := range simulator.GetNearAtoms(0) {
				render.DrawCube(fig, simulator.Pos[j], 5., 1., []float64{1, 0, 0, 1})
			}
			render.DrawPlaneText(fig, -50., -50., fmt.Sprintf("Time: %.2f", simulator.T), 5., "D2CodingNerd.ttf", []float64{1, 0, 0, 1})
			render.DrawCube(fig, Vector{0., 0., 0.}, 102., 1., []float64{1, 0, 0, 1})
			render.DrawAxis(fig, 5., 5., 2., "D2CodingNerd.ttf")
			render.DrawLine(fig, Vector{0., 0., 0.}, simulator.Gravity, 1., []float64{1, 0, 0, 1})
			render.Save(fig, "images", simulator.Count)
		}
		simulator.Gravity = InvRenderSO3(render.Angle).DotV(Vector{0.0, 0.0, -100.0})
		simulator.Step()

		// simulator.PeriodicBoundary(102.)
		fmt.Println(simulator.T)
	}
}
