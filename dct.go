package imagehashgo

import (
	"math"
)

// DCT2D computes the 2D Discrete Cosine Transform (DCT-II) of a matrix
func DCT2D(input [][]float64) [][]float64 {
	rows := len(input)
	if rows == 0 {
		return nil
	}
	cols := len(input[0])

	// DCT rows
	rowDCT := make([][]float64, rows)
	for i := 0; i < rows; i++ {
		rowDCT[i] = DCT1D(input[i])
	}

	// DCT columns
	result := make([][]float64, rows)
	for i := 0; i < rows; i++ {
		result[i] = make([]float64, cols)
	}

	for j := 0; j < cols; j++ {
		col := make([]float64, rows)
		for i := 0; i < rows; i++ {
			col[i] = rowDCT[i][j]
		}
		colDCT := DCT1D(col)
		for i := 0; i < rows; i++ {
			result[i][j] = colDCT[i]
		}
	}

	return result
}

// DCT1D computes the 1D Discrete Cosine Transform (DCT-II) of a vector
func DCT1D(input []float64) []float64 {
	n := len(input)
	output := make([]float64, n)
	factor := math.Pi / float64(n)

	for k := 0; k < n; k++ {
		var sum float64
		for i := 0; i < n; i++ {
			sum += input[i] * math.Cos(factor*(float64(i)+0.5)*float64(k))
		}
		output[k] = sum
	}
	return output
}
