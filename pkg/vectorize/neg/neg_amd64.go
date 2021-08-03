// +build amd64

package neg

import (
	"golang.org/x/sys/cpu"
)

func int8NegAvx2Asm([]int8, []int8)
func int8NegAvx512Asm([]int8, []int8)
func int16NegAvx2Asm([]int16, []int16)
func int16NegAvx512Asm([]int16, []int16)
func int32NegAvx2Asm([]int32, []int32)
func int32NegAvx512Asm([]int32, []int32)
func int64NegAvx2Asm([]int64, []int64)
func int64NegAvx512Asm([]int64, []int64)
func float32NegAvx2Asm([]float32, []float32)
func float32NegAvx512Asm([]float32, []float32)
func float64NegAvx2Asm([]float64, []float64)
func float64NegAvx512Asm([]float64, []float64)

func init() {
	if cpu.X86.HasAVX512 {
		Int8Neg = int8NegAvx512
		Int16Neg = int16NegAvx512
		Int32Neg = int32NegAvx512
		Int64Neg = int64NegAvx512
		Float32Neg = float32NegAvx512
		Float64Neg = float64NegAvx512
	} else if cpu.X86.HasAVX2 {
		Int8Neg = int8NegAvx2
		Int16Neg = int16NegAvx2
		Int32Neg = int32NegAvx2
		Int64Neg = int64NegAvx2
		Float32Neg = float32NegAvx2
		Float64Neg = float64NegAvx2
	} else {
		Int8Neg = int8Neg
		Int16Neg = int16Neg
		Int32Neg = int32Neg
		Int64Neg = int64Neg
		Float32Neg = float32Neg
		Float64Neg = float64Neg
	}
}

func int8NegAvx2(xs, rs []int8) []int8 {
	const regItems int = 32 / 1
	n := len(xs) / regItems
	int8NegAvx2Asm(xs[:n*regItems], rs[:n*regItems])
	for i, j := n*regItems, len(xs); i < j; i++ {
		rs[i] = -xs[i]
	}
	return rs
}

func int8NegAvx512(xs, rs []int8) []int8 {
	const regItems int = 64 / 1
	n := len(xs) / regItems
	int8NegAvx512Asm(xs[:n*regItems], rs[:n*regItems])
	for i, j := n*regItems, len(xs); i < j; i++ {
		rs[i] = -xs[i]
	}
	return rs
}

func int16NegAvx2(xs, rs []int16) []int16 {
	const regItems int = 32 / 2
	n := len(xs) / regItems
	int16NegAvx2Asm(xs[:n*regItems], rs[:n*regItems])
	for i, j := n*regItems, len(xs); i < j; i++ {
		rs[i] = -xs[i]
	}
	return rs
}

func int16NegAvx512(xs, rs []int16) []int16 {
	const regItems int = 64 / 2
	n := len(xs) / regItems
	int16NegAvx512Asm(xs[:n*regItems], rs[:n*regItems])
	for i, j := n*regItems, len(xs); i < j; i++ {
		rs[i] = -xs[i]
	}
	return rs
}

func int32NegAvx2(xs, rs []int32) []int32 {
	const regItems int = 32 / 4
	n := len(xs) / regItems
	int32NegAvx2Asm(xs[:n*regItems], rs[:n*regItems])
	for i, j := n*regItems, len(xs); i < j; i++ {
		rs[i] = -xs[i]
	}
	return rs
}

func int32NegAvx512(xs, rs []int32) []int32 {
	const regItems int = 64 / 4
	n := len(xs) / regItems
	int32NegAvx512Asm(xs[:n*regItems], rs[:n*regItems])
	for i, j := n*regItems, len(xs); i < j; i++ {
		rs[i] = -xs[i]
	}
	return rs
}

func int64NegAvx2(xs, rs []int64) []int64 {
	const regItems int = 32 / 8
	n := len(xs) / regItems
	int64NegAvx2Asm(xs[:n*regItems], rs[:n*regItems])
	for i, j := n*regItems, len(xs); i < j; i++ {
		rs[i] = -xs[i]
	}
	return rs
}

func int64NegAvx512(xs, rs []int64) []int64 {
	const regItems int = 64 / 8
	n := len(xs) / regItems
	int64NegAvx512Asm(xs[:n*regItems], rs[:n*regItems])
	for i, j := n*regItems, len(xs); i < j; i++ {
		rs[i] = -xs[i]
	}
	return rs
}

func float32NegAvx2(xs, rs []float32) []float32 {
	const regItems int = 32 / 4
	n := len(xs) / regItems
	float32NegAvx2Asm(xs[:n*regItems], rs[:n*regItems])
	for i, j := n*regItems, len(xs); i < j; i++ {
		rs[i] = -xs[i]
	}
	return rs
}

func float32NegAvx512(xs, rs []float32) []float32 {
	const regItems int = 64 / 4
	n := len(xs) / regItems
	float32NegAvx512Asm(xs[:n*regItems], rs[:n*regItems])
	for i, j := n*regItems, len(xs); i < j; i++ {
		rs[i] = -xs[i]
	}
	return rs
}

func float64NegAvx2(xs, rs []float64) []float64 {
	const regItems int = 32 / 8
	n := len(xs) / regItems
	float64NegAvx2Asm(xs[:n*regItems], rs[:n*regItems])
	for i, j := n*regItems, len(xs); i < j; i++ {
		rs[i] = -xs[i]
	}
	return rs
}

func float64NegAvx512(xs, rs []float64) []float64 {
	const regItems int = 64 / 8
	n := len(xs) / regItems
	float64NegAvx512Asm(xs[:n*regItems], rs[:n*regItems])
	for i, j := n*regItems, len(xs); i < j; i++ {
		rs[i] = -xs[i]
	}
	return rs
}
