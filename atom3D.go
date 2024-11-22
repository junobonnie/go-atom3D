package atom3D

import (
	"fmt"
	"log"
	"os"

	"gonum.org/v1/hdf5"
)

type Simulator struct {
	Dt      float64
	T       float64
	Count   int
	N       int
	Id      []int
	Pos     []Vector
	Vel     []Vector
	Gravity Vector
}

func NewSimulator(Dt float64, Id []int, Pos, Vel []Vector, Gravity Vector) *Simulator {
	return &Simulator{
		Dt:      Dt,
		T:       0.0,
		Count:   0,
		N:       len(Pos),
		Id:      Id,
		Pos:     Pos,
		Vel:     Vel,
		Gravity: Gravity,
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

func Read(filename string) (float64, float64, int, int, Vector, []Vector, []Vector) {
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

func main() {
	// Create a new simulator
	id := []int{0, 1}
	pos := []Vector{Vector{0.0, 0.0, 0.0}, Vector{0.0, 0.0, 1.0}}
	vel := []Vector{Vector{0.0, 0.0, 0.0}, Vector{0.0, 0.0, 0.0}}
	simulator := NewSimulator(0.01, id, pos, vel, Vector{0.0, 0.0, -9.8})
	//simulator.Load("output/snapshot_00000991.hdf5")

	for i := 0; i < 1000; i++ {
		if i%10 == 0 {
			simulator.Save("output")
		}
		simulator.Step()
	}
}
