package atom3D

import (
	"fmt"
	"log"
	"math"
	"os"

	"gonum.org/v1/hdf5"
)

type Simulator struct {
	Dt         float64
	T          float64
	Count      int
	N          int
	Id         []int
	Pos        []Vector
	Vel        []Vector
	Gravity    Vector
	RegionSize float64
	GridSize   float64
	Grid       [][]int
}

func NewSimulator(Dt float64, Id []int, Pos, Vel []Vector, Gravity Vector) *Simulator {
	return &Simulator{
		Dt:         Dt,
		T:          0.0,
		Count:      0,
		N:          len(Pos),
		Id:         Id,
		Pos:        Pos,
		Vel:        Vel,
		Gravity:    Gravity,
		RegionSize: 0.0,
		GridSize:   0.0,
		Grid:       [][]int{},
	}
}

func (simulator *Simulator) Step() {
	simulator.Count++
	simulator.T = float64(simulator.Count) * simulator.Dt

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
}

func (simulator *Simulator) SolidBoundary(length float64) {
	half_length := length / 2
	for i := 0; i < simulator.N; i++ {
		for {
			is_collision := false
			if simulator.Pos[i].X < -half_length {
				collision_time := (simulator.Pos[i].X + half_length) / (simulator.Vel[i].X * simulator.Dt)
				if (0. < collision_time) && (collision_time <= 1.) {
					Y_collision := simulator.Pos[i].Y - simulator.Vel[i].Y*simulator.Dt*collision_time
					Z_collision := simulator.Pos[i].Z - simulator.Vel[i].Z*simulator.Dt*collision_time
					if math.Abs(Y_collision) < half_length && math.Abs(Z_collision) < half_length {
						simulator.Vel[i].X = -simulator.Vel[i].X
						simulator.Pos[i].X = -half_length + simulator.Vel[i].X*simulator.Dt*collision_time
						is_collision = true
					}
				}
			}
			if half_length < simulator.Pos[i].X {
				collision_time := (simulator.Pos[i].X - half_length) / (simulator.Vel[i].X * simulator.Dt)
				if (0. < collision_time) && (collision_time <= 1.) {
					Y_collision := simulator.Pos[i].Y - simulator.Vel[i].Y*simulator.Dt*collision_time
					Z_collision := simulator.Pos[i].Z - simulator.Vel[i].Z*simulator.Dt*collision_time
					if math.Abs(Y_collision) < half_length && math.Abs(Z_collision) < half_length {
						simulator.Vel[i].X = -simulator.Vel[i].X
						simulator.Pos[i].X = half_length + simulator.Vel[i].X*simulator.Dt*collision_time
						is_collision = true
					}
				}
			}
			if simulator.Pos[i].Y < -half_length {
				collision_time := (simulator.Pos[i].Y + half_length) / (simulator.Vel[i].Y * simulator.Dt)
				if (0. < collision_time) && (collision_time <= 1.) {
					X_collision := simulator.Pos[i].X - simulator.Vel[i].X*simulator.Dt*collision_time
					Z_collision := simulator.Pos[i].Z - simulator.Vel[i].Z*simulator.Dt*collision_time
					if math.Abs(X_collision) < half_length && math.Abs(Z_collision) < half_length {
						simulator.Vel[i].Y = -simulator.Vel[i].Y
						simulator.Pos[i].Y = -half_length + simulator.Vel[i].Y*simulator.Dt*collision_time
						is_collision = true
					}
				}
			}
			if half_length < simulator.Pos[i].Y {
				collision_time := (simulator.Pos[i].Y - half_length) / (simulator.Vel[i].Y * simulator.Dt)
				if (0. < collision_time) && (collision_time <= 1.) {
					X_collision := simulator.Pos[i].X - simulator.Vel[i].X*simulator.Dt*collision_time
					Z_collision := simulator.Pos[i].Z - simulator.Vel[i].Z*simulator.Dt*collision_time
					if math.Abs(X_collision) < half_length && math.Abs(Z_collision) < half_length {
						simulator.Vel[i].Y = -simulator.Vel[i].Y
						simulator.Pos[i].Y = half_length + simulator.Vel[i].Y*simulator.Dt*collision_time
						is_collision = true
					}
				}
			}
			if simulator.Pos[i].Z < -half_length {
				collision_time := (simulator.Pos[i].Z + half_length) / (simulator.Vel[i].Z * simulator.Dt)
				if (0. < collision_time) && (collision_time <= 1.) {
					X_collision := simulator.Pos[i].X - simulator.Vel[i].X*simulator.Dt*collision_time
					Y_collision := simulator.Pos[i].Y - simulator.Vel[i].Y*simulator.Dt*collision_time
					if math.Abs(X_collision) < half_length && math.Abs(Y_collision) < half_length {
						simulator.Vel[i].Z = -simulator.Vel[i].Z
						simulator.Pos[i].Z = -half_length + simulator.Vel[i].Z*simulator.Dt*collision_time
						is_collision = true
					}
				}
			}
			if half_length < simulator.Pos[i].Z {
				collision_time := (simulator.Pos[i].Z - half_length) / (simulator.Vel[i].Z * simulator.Dt)
				if (0. < collision_time) && (collision_time <= 1.) {
					X_collision := simulator.Pos[i].X - simulator.Vel[i].X*simulator.Dt*collision_time
					Y_collision := simulator.Pos[i].Y - simulator.Vel[i].Y*simulator.Dt*collision_time
					if math.Abs(X_collision) < half_length && math.Abs(Y_collision) < half_length {
						simulator.Vel[i].Z = -simulator.Vel[i].Z
						simulator.Pos[i].Z = half_length + simulator.Vel[i].Z*simulator.Dt*collision_time
						is_collision = true
					}
				}
			}
			if is_collision == false {
				break
			}
		}
	}
}

func (simulator *Simulator) PeriodicBoundary(length float64) {
	half_length := length / 2
	for i := 0; i < simulator.N; i++ {
		if simulator.Pos[i].X > half_length {
			simulator.Pos[i].X = simulator.Pos[i].X - length
		}
		if simulator.Pos[i].X < -half_length {
			simulator.Pos[i].X = simulator.Pos[i].X + length
		}
		if simulator.Pos[i].Y > half_length {
			simulator.Pos[i].Y = simulator.Pos[i].Y - length
		}
		if simulator.Pos[i].Y < -half_length {
			simulator.Pos[i].Y = simulator.Pos[i].Y + length
		}
		if simulator.Pos[i].Z > half_length {
			simulator.Pos[i].Z = simulator.Pos[i].Z - length
		}
		if simulator.Pos[i].Z < -half_length {
			simulator.Pos[i].Z = simulator.Pos[i].Z + length
		}
	}
}

func (simulator *Simulator) MakeGrid() {
	n := int(simulator.RegionSize/simulator.GridSize + 1)
	grid := make([][]int, n*n*n)
	for i := 0; i < simulator.N; i++ {
		x := int((simulator.Pos[i].X + simulator.RegionSize/2) / simulator.GridSize)
		y := int((simulator.Pos[i].Y + simulator.RegionSize/2) / simulator.GridSize)
		z := int((simulator.Pos[i].Z + simulator.RegionSize/2) / simulator.GridSize)
		grid[x+y*n+z*n*n] = append(grid[x+y*n+z*n*n], i)
	}
	simulator.Grid = grid
}

func Mod(a, b int) int {
	result := a % b
	if result < 0 {
		result += b
	}
	return result
}

func removeValue(slice []int, value int) []int {
	for i, v := range slice {
		if v == value {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}

func (simulator *Simulator) GetNearAtoms(atom_index int, is_periodic ...bool) []int {

	if len(is_periodic) == 0 {
		is_periodic = []bool{false}
	}

	n := int(simulator.RegionSize/simulator.GridSize + 1)
	x := int((simulator.Pos[atom_index].X + simulator.RegionSize/2) / simulator.GridSize)
	y := int((simulator.Pos[atom_index].Y + simulator.RegionSize/2) / simulator.GridSize)
	z := int((simulator.Pos[atom_index].Z + simulator.RegionSize/2) / simulator.GridSize)
	indices := []int{}

	if is_periodic[0] == false {
		for i := x - 1; i <= x+1; i++ {
			for j := y - 1; j <= y+1; j++ {
				for k := z - 1; k <= z+1; k++ {
					if 0 <= i && i < n && 0 <= j && j < n && 0 <= k && k < n {
						indices = append(indices, simulator.Grid[i+j*n+k*n*n]...)
					}
				}
			}
		}
	} else {
		x_min := 1
		if x == 0 {
			x_min = 2
		}
		y_min := 1
		if y == 0 {
			y_min = 2
		}
		z_min := 1
		if z == 0 {
			z_min = 2
		}
		x_max := 1
		if x == n-2 {
			x_max = 2
		}
		y_max := 1
		if y == n-2 {
			y_max = 2
		}
		z_max := 1
		if z == n-2 {
			z_max = 2
		}
		for i := x - x_min; i <= x+x_max; i++ {
			for j := y - y_min; j <= y+y_max; j++ {
				for k := z - z_min; k <= z+z_max; k++ {
					i_ := Mod(i, n)
					j_ := Mod(j, n)
					k_ := Mod(k, n)
					// if atom_index == 0 {
					// 	fmt.Println(n)
					// 	fmt.Println("i, j, k:", i_, j_, k_)

					// }

					indices = append(indices, simulator.Grid[i_+j_*n+k_*n*n]...)
				}
			}
		}
	}

	return removeValue(indices, atom_index)
}

func (simulator *Simulator) PeriodicDisplacement(atom_index int, another_atom_index int) Vector {
	half_region := simulator.RegionSize / 2
	dx := simulator.Pos[another_atom_index].X - simulator.Pos[atom_index].X
	dy := simulator.Pos[another_atom_index].Y - simulator.Pos[atom_index].Y
	dz := simulator.Pos[another_atom_index].Z - simulator.Pos[atom_index].Z

	if dx > half_region {
		dx -= simulator.RegionSize
	} else if dx < -half_region {
		dx += simulator.RegionSize
	}
	if dy > half_region {
		dy -= simulator.RegionSize
	} else if dy < -half_region {
		dy += simulator.RegionSize
	}
	if dz > half_region {
		dz -= simulator.RegionSize
	} else if dz < -half_region {
		dz += simulator.RegionSize
	}

	return Vector{X: dx, Y: dy, Z: dz}
}

func (simulator *Simulator) Save(directory string) {
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		os.Mkdir(directory, os.ModeDir|0755)
	}
	filename := directory + fmt.Sprintf("/snapshot_%010d.hdf5", simulator.Count)
	// HDF5 파일 생성
	f, err := hdf5.CreateFile(filename, hdf5.F_ACC_TRUNC)
	if err != nil {
		log.Fatalf("Error creating file: %s", err)
	}
	defer f.Close()

	rootGroup, _ := f.OpenGroup("/")
	defer rootGroup.Close()

	// Attribute 생성
	CreateAttributeFloat(rootGroup, "Dt", simulator.Dt)
	CreateAttributeFloat(rootGroup, "T", simulator.T)
	CreateAttributeInt(rootGroup, "Count", simulator.Count)
	CreateAttributeInt(rootGroup, "N", simulator.N)
	CreateAttributeVector(rootGroup, "Gravity", simulator.Gravity)

	// Dataset 생성
	CreateDatasetInt(rootGroup, "Id", simulator.Id, []uint{uint(simulator.N)})

	pos := make([]float64, 3*simulator.N)
	vel := make([]float64, 3*simulator.N)

	for i := 0; i < simulator.N; i++ {
		pos[3*i] = simulator.Pos[i].X
		pos[3*i+1] = simulator.Pos[i].Y
		pos[3*i+2] = simulator.Pos[i].Z

		vel[3*i] = simulator.Vel[i].X
		vel[3*i+1] = simulator.Vel[i].Y
		vel[3*i+2] = simulator.Vel[i].Z
	}
	CreateDatasetFloat(rootGroup, "Pos", pos, []uint{uint(simulator.N), 3})
	CreateDatasetFloat(rootGroup, "Vel", vel, []uint{uint(simulator.N), 3})
}

func Read(filename string) (float64, float64, int, int, Vector, []int, []Vector, []Vector) {
	// HDF5 파일 열기
	file, err := hdf5.OpenFile(filename, hdf5.F_ACC_RDONLY)
	if err != nil {
		log.Fatalf("파일을 열 수 없습니다: %v", err)
	}
	defer file.Close()

	rootGroup, _ := file.OpenGroup("/")
	defer rootGroup.Close()

	dt := ReadAttributeFloat(rootGroup, "Dt")
	t := ReadAttributeFloat(rootGroup, "T")
	count := ReadAttributeInt(rootGroup, "Count")
	N := ReadAttributeInt(rootGroup, "N")
	gravity := ReadAttributeVector(rootGroup, "Gravity")

	id := ReadDatasetInt(rootGroup, "Id")

	pos := ReadDatasetVector(rootGroup, "Pos")
	vel := ReadDatasetVector(rootGroup, "Vel")

	return dt, t, count, N, gravity, id, pos, vel
}

func (simulator *Simulator) Load(filename string) {
	// HDF5 파일 읽기
	dt, t, count, N, gravity, id, pos, vel := Read(filename)

	simulator.Dt = dt
	simulator.T = t
	simulator.Count = count
	simulator.N = N
	simulator.Gravity = gravity

	simulator.Id = id
	simulator.Pos = pos
	simulator.Vel = vel
}
