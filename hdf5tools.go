package main

import (
	"fmt"
	"gonum.org/v1/hdf5"
)


func CreateAttributeFloat(group *hdf5.Group, name string, value float64) {
	space, _ := hdf5.CreateSimpleDataspace([]uint{1}, nil)
	attr, _ := group.CreateAttribute(name, hdf5.T_NATIVE_DOUBLE, space)
	attr.Write(&value, hdf5.T_NATIVE_DOUBLE)
	attr.Close()
}

func CreateAttributeInt(group *hdf5.Group, name string, value int) {
	space, _ := hdf5.CreateSimpleDataspace([]uint{1}, nil)
	attr, _ := group.CreateAttribute(name, hdf5.T_NATIVE_INT, space)
	attr.Write(&value, hdf5.T_NATIVE_INT)
	attr.Close()
}

func CreateAttributeVector(group *hdf5.Group, name string, value Vector) {
	space, _ := hdf5.CreateSimpleDataspace([]uint{3}, nil)
	attr, _ := group.CreateAttribute(name, hdf5.T_NATIVE_DOUBLE, space)
	attr.Write(&value, hdf5.T_NATIVE_DOUBLE)
	attr.Close()
}

func CreateDatasetFloat(group *hdf5.Group, name string, data []float64, dims []uint) {
	space, _ := hdf5.CreateSimpleDataspace(dims, nil)
	dataset, _ := group.CreateDataset(name, hdf5.T_NATIVE_DOUBLE, space)
	dataset.Write(&data)
	dataset.Close()
}

func CreateDatasetInt(group *hdf5.Group, name string, data []int, dims []uint) {
	space, _ := hdf5.CreateSimpleDataspace(dims, nil)
	dataset, _ := group.CreateDataset(name, hdf5.T_NATIVE_INT, space)
	dataset.Write(&data)
	dataset.Close()
}

func ReadAttributeFloat(group *hdf5.Group, name string) float64 {
	attr, _ := group.OpenAttribute(name)
	var value float64
	attr.Read(&value, hdf5.T_NATIVE_DOUBLE)
	attr.Close()
	return value
}

func ReadAttributeInt(group *hdf5.Group, name string) int {
	attr, _ := group.OpenAttribute(name)
	var value int
	attr.Read(&value, hdf5.T_NATIVE_INT)
	attr.Close()
	return value
}

func ReadAttributeVector(group *hdf5.Group, name string) Vector {
	attr, _ := group.OpenAttribute(name)
	var value Vector
	attr.Read(&value, hdf5.T_NATIVE_DOUBLE)
	attr.Close()
	return value
}

func ReadDatasetVector(group *hdf5.Group, name string) []Vector {
	dataset, _ := group.OpenDataset(name)
	defer dataset.Close()
	dataspace := dataset.Space()
	dims, _, _ := dataspace.SimpleExtentDims()
	fmt.Println(dims)
	data := make([]Vector, dims[0])
	dataset.Read(&data)
	return data
}