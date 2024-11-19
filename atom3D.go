package main

import (
	"fmt"
	"log"
	"os"

	"gonum.org/v1/hdf5"
)

type Simulator struct {
	dt      float64
	t       float64
	count   int
	N       int
	pos     []Vector
	vel     []Vector
	gravity Vector
}

func NewSimulator(dt float64, pos, vel []Vector, gravity Vector) *Simulator {
	return &Simulator{
		dt:      dt,
		t:       0.0,
		count:   0,
		N:       len(pos),
		pos:     pos,
		vel:     vel,
		gravity: gravity,
	}
}

func (simulator *Simulator) Step() {
	simulator.count++
	simulator.t = float64(simulator.count) * simulator.dt

	x_ := make([]Vector, simulator.N)
	v_ := make([]Vector, simulator.N)
	for i := 0; i < simulator.N; i++ {
		new_vel := simulator.vel[i].Add(simulator.gravity.Mul(simulator.dt))
		v_[i] = new_vel
		x_[i] = simulator.pos[i].Add(new_vel.Mul(simulator.dt))
	}

	for i := 0; i < simulator.N; i++ {
		simulator.pos[i] = x_[i]
		simulator.vel[i] = v_[i]
	}
}

func (simulator *Simulator) Save(directory string) {

	if _, err := os.Stat(directory); os.IsNotExist(err) {
		os.Mkdir(directory, os.ModeDir|0755)
	}
	filename := directory + fmt.Sprintf("/snapshot_%08d.hdf5", simulator.count)
	// HDF5 파일 생성
	f, err := hdf5.CreateFile(filename, hdf5.F_ACC_TRUNC)
	if err != nil {
		log.Fatalf("Error creating file: %s", err)
	}
	defer f.Close()

	rootGroup, _ := f.OpenGroup("/")
	defer rootGroup.Close()

	// Attribute 생성
	CreateAttributeVector(rootGroup, "gravity", simulator.gravity)
	CreateAttributeInt(rootGroup, "N", simulator.N)
	CreateAttributeInt(rootGroup, "count", simulator.count)
	CreateAttributeFloat(rootGroup, "t", simulator.t)
	CreateAttributeFloat(rootGroup, "dt", simulator.dt)

	pos := make([]float64, 3*simulator.N)
	vel := make([]float64, 3*simulator.N)

	for i := 0; i < simulator.N; i++ {
		pos[3*i] = simulator.pos[i].x
		pos[3*i+1] = simulator.pos[i].y
		pos[3*i+2] = simulator.pos[i].z

		vel[3*i] = simulator.vel[i].x
		vel[3*i+1] = simulator.vel[i].y
		vel[3*i+2] = simulator.vel[i].z
	}
	CreateDatasetFloat(rootGroup, "pos", pos, []uint{uint(simulator.N), 3})
	CreateDatasetFloat(rootGroup, "vel", vel, []uint{uint(simulator.N), 3})
}

func (simulator *Simulator) Load(filename string) {
	// HDF5 파일 열기
	file, err := hdf5.OpenFile(filename, hdf5.F_ACC_RDONLY)
	if err != nil {
		log.Fatalf("파일을 열 수 없습니다: %v", err)
	}
	defer file.Close()

	rootGroup, _ := file.OpenGroup("/")
	defer rootGroup.Close()

	dt := ReadAttributeFloat(rootGroup, "dt")
	t := ReadAttributeFloat(rootGroup, "t")
	count := ReadAttributeInt(rootGroup, "count")
	N := ReadAttributeInt(rootGroup, "N")
	gravity := ReadAttributeVector(rootGroup, "gravity")

	pos := ReadDatasetVector(rootGroup, "pos")
	vel := ReadDatasetVector(rootGroup, "vel")

	simulator.dt = dt
	simulator.t = t
	simulator.count = count
	simulator.N = N
	simulator.gravity = gravity

	simulator.pos = pos
	simulator.vel = vel
}

func main() {
	pos := []Vector{Vector{0.0, 0.0, 0.0}, Vector{0.0, 0.0, 1.0}}
	vel := []Vector{Vector{0.0, 0.0, 0.0}, Vector{0.0, 0.0, 0.0}}
	simulator := NewSimulator(0.01, pos, vel, Vector{0.0, 0.0, -9.8})
	//simulator.Load("output/snapshot_00000991.hdf5")

	for i := 0; i < 1000; i++ {
		simulator.Step()
		if i%10 == 0 {
			simulator.Save("output")
		}
	}
}
