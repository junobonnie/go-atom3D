# hdf5tools.go — HDF5 I/O 헬퍼

`gonum.org/v1/hdf5` 를 래핑한 저수준 I/O 유틸리티 함수 모음.
`atom3D.Simulator.Save()` / `Read()` 에 내부적으로 사용됩니다.

---

## Attribute 쓰기

| 함수 | 설명 |
|---|---|
| `CreateAttributeFloat(group, name, value)` | float64 스칼라 속성 생성 |
| `CreateAttributeInt(group, name, value)` | int 스칼라 속성 생성 |
| `CreateAttributeVector(group, name, value)` | `Vector` (3개 float64) 속성 생성 |

## Dataset 쓰기

| 함수 | 설명 |
|---|---|
| `CreateDatasetFloat(group, name, data, dims)` | float64 배열 데이터셋 생성 |
| `CreateDatasetInt(group, name, data, dims)` | int 배열 데이터셋 생성 |

## Attribute 읽기

| 함수 | 반환값 | 설명 |
|---|---|---|
| `ReadAttributeFloat(group, name)` | `float64` | float64 스칼라 속성 읽기 |
| `ReadAttributeInt(group, name)` | `int` | int 스칼라 속성 읽기 |
| `ReadAttributeVector(group, name)` | `Vector` | 3D 벡터 속성 읽기 |

## Dataset 읽기

| 함수 | 반환값 | 설명 |
|---|---|---|
| `ReadDatasetFloat(group, name)` | `[]float64` | float64 1D 배열 읽기 |
| `ReadDatasetInt(group, name)` | `[]int` | int 1D 배열 읽기 |
| `ReadDatasetVector(group, name)` | `[]Vector` | (N,3) float64 배열 → `[]Vector` 변환 읽기 |

---

## HDF5 스냅샷 파일 구조

`Simulator.Save()` 가 생성하는 HDF5 파일의 스키마:

```
/
├── Attributes
│   ├── Dt      (float64)
│   ├── T       (float64)
│   ├── Count   (int)
│   ├── N       (int)
│   └── Gravity (float64[3])
└── Datasets
    ├── Id      (int[N])
    ├── Pos     (float64[N][3])
    └── Vel     (float64[N][3])
```

> **주의**: `ReadDatasetVector` 는 (N,3) 형태의 평탄 배열을 `[]Vector` 로 재구성합니다.
> 1D 배열 형식으로 저장된 경우 `dims[0]` 을 벡터 수로 사용합니다.
