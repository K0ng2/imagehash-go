package imagehashgo

import (
	"math"
	"sync"
)

var slicePool = sync.Pool{
	New: func() any {
		return make([]float64, 0, 64) // Default size for 8x8
	},
}

func getSlice(size int) []float64 {
	s := slicePool.Get().([]float64)
	if cap(s) < size {
		return make([]float64, size)
	}
	return s[:size]
}

func putSlice(s []float64) {
	slicePool.Put(s)
}

// DCT2D computes the 2D Discrete Cosine Transform (DCT-II) of a matrix
func DCT2D(input [][]float64) [][]float64 {
	rows := len(input)
	if rows == 0 {
		return nil
	}
	cols := len(input[0])

	// DCT rows
	rowDCT := make([][]float64, rows)
	var wg sync.WaitGroup
	wg.Add(rows)
	for i := range rows {
		go func(rowIdx int) {
			defer wg.Done()
			rowDCT[rowIdx] = DCT1D(input[rowIdx])
		}(i)
	}
	wg.Wait()

	// DCT columns
	result := make([][]float64, rows)
	for i := range rows {
		result[i] = make([]float64, cols)
	}

	wg.Add(cols)
	for j := range cols {
		go func(colIdx int) {
			defer wg.Done()
			col := getSlice(rows)
			for i := range rows {
				col[i] = rowDCT[i][colIdx]
			}
			colDCT := DCT1D(col)
			for i := range rows {
				result[i][colIdx] = colDCT[i]
			}
			putSlice(col)
		}(j)
	}
	wg.Wait()

	return result
}

// DCT1D computes the 1D Discrete Cosine Transform (DCT-II) of a vector
func DCT1D(input []float64) []float64 {
	n := len(input)
	output := make([]float64, n)
	factor := math.Pi / float64(n)

	for k := range n {
		var sum float64
		for i := range n {
			sum += input[i] * math.Cos(factor*(float64(i)+0.5)*float64(k))
		}
		output[k] = sum
	}
	return output
}
