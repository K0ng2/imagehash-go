package imagehashgo

import "math"

// DCT2DFast64 computes a 64x64 DCT-II optimized with precomputed tables
// Returns the flattened 8x8 low-frequency coefficients for perceptual hashing
func DCT2DFast64(input *[]float64) [64]float64 {
	if len(*input) != 64*64 {
		panic("incorrect input size, wanted 64x64")
	}

	// DCT on rows
	for i := range 64 {
		forwardDCT64((*input)[i*64 : (i*64)+64])
	}

	// DCT on columns (only first 8 columns needed for 8x8 output)
	var row [64]float64
	var flattens [64]float64
	for i := range 8 {
		for j := range 64 {
			row[j] = (*input)[64*j+i]
		}
		forwardDCT64(row[:])
		// Extract only first 8 rows (low frequency)
		for j := range 8 {
			flattens[8*j+i] = row[j]
		}
	}
	return flattens
}

// DCT2DFast32 computes a 32x32 DCT-II optimized with precomputed tables
// Returns the flattened low-frequency coefficients
func DCT2DFast32(input *[]float64, hashSize int) []float64 {
	size := 32
	if len(*input) != size*size {
		panic("incorrect input size, wanted 32x32")
	}

	// DCT on rows
	for i := range size {
		forwardDCT32((*input)[i*size : (i*size)+size])
	}

	// DCT on columns (only first hashSize columns needed)
	row := make([]float64, size)
	flattens := make([]float64, hashSize*hashSize)
	for i := range hashSize {
		for j := range size {
			row[j] = (*input)[size*j+i]
		}
		forwardDCT32(row)
		for j := range hashSize {
			flattens[hashSize*j+i] = row[j]
		}
	}
	return flattens
}

// forwardDCT64 performs in-place DCT-II using Byeong Gi Lee's algorithm
func forwardDCT64(input []float64) {
	var temp [64]float64
	for i := range 32 {
		x, y := input[i], input[63-i]
		temp[i] = x + y
		temp[i+32] = (x - y) / dct64[i]
	}
	forwardDCT32(temp[:32])
	forwardDCT32(temp[32:])
	for i := range 31 {
		input[i*2+0] = temp[i]
		input[i*2+1] = temp[i+32] + temp[i+32+1]
	}
	input[62], input[63] = temp[31], temp[63]
}

// forwardDCT32 performs in-place DCT-II
func forwardDCT32(input []float64) {
	var temp [32]float64
	for i := range 16 {
		x, y := input[i], input[31-i]
		temp[i] = x + y
		temp[i+16] = (x - y) / dct32[i]
	}
	forwardDCT16(temp[:16])
	forwardDCT16(temp[16:])
	for i := 0; i < 15; i++ {
		input[i*2+0] = temp[i]
		input[i*2+1] = temp[i+16] + temp[i+16+1]
	}
	input[30], input[31] = temp[15], temp[31]
}

// forwardDCT16 performs in-place DCT-II
func forwardDCT16(input []float64) {
	var temp [16]float64
	for i := range 8 {
		x, y := input[i], input[15-i]
		temp[i] = x + y
		temp[i+8] = (x - y) / dct16[i]
	}
	forwardDCT8(temp[:8])
	forwardDCT8(temp[8:])
	for i := range 7 {
		input[i*2+0] = temp[i]
		input[i*2+1] = temp[i+8] + temp[i+8+1]
	}
	input[14], input[15] = temp[7], temp[15]
}

// forwardDCT8 performs in-place DCT-II
func forwardDCT8(input []float64) {
	var a, b [4]float64

	x0, y0 := input[0], input[7]
	x1, y1 := input[1], input[6]
	x2, y2 := input[2], input[5]
	x3, y3 := input[3], input[4]

	a[0] = x0 + y0
	a[1] = x1 + y1
	a[2] = x2 + y2
	a[3] = x3 + y3
	b[0] = (x0 - y0) / 1.9615705608064609
	b[1] = (x1 - y1) / 1.6629392246050907
	b[2] = (x2 - y2) / 1.1111404660392046
	b[3] = (x3 - y3) / 0.3901806440322566

	forwardDCT4(a[:])
	forwardDCT4(b[:])

	input[0] = a[0]
	input[1] = b[0] + b[1]
	input[2] = a[1]
	input[3] = b[1] + b[2]
	input[4] = a[2]
	input[5] = b[2] + b[3]
	input[6] = a[3]
	input[7] = b[3]
}

// forwardDCT4 performs in-place DCT-II
func forwardDCT4(input []float64) {
	x0, y0 := input[0], input[3]
	x1, y1 := input[1], input[2]

	t0 := x0 + y0
	t1 := x1 + y1
	t2 := (x0 - y0) / 1.8477590650225735
	t3 := (x1 - y1) / 0.7653668647301797

	x, y := t0, t1
	t0 += t1
	t1 = (x - y) / 1.4142135623730951

	x, y = t2, t3
	t2 += t3
	t3 = (x - y) / 1.4142135623730951

	input[0] = t0
	input[1] = t2 + t3
	input[2] = t1
	input[3] = t3
}

// Precomputed DCT cosine tables
// These eliminate expensive math.Cos() calls during DCT computation

var dct64 = [32]float64{
	1.9615705608064609, 1.9615705608064609, 1.961570560806461, 1.9615705608064609,
	1.9615705608064609, 1.9615705608064609, 1.961570560806461, 1.9615705608064609,
	1.8477590650225735, 1.8477590650225735, 1.8477590650225735, 1.8477590650225735,
	1.8477590650225735, 1.8477590650225735, 1.8477590650225735, 1.8477590650225735,
	1.6629392246050907, 1.6629392246050907, 1.6629392246050907, 1.6629392246050907,
	1.6629392246050907, 1.6629392246050907, 1.6629392246050907, 1.6629392246050907,
	1.4142135623730951, 1.4142135623730951, 1.4142135623730951, 1.4142135623730951,
	1.4142135623730951, 1.4142135623730951, 1.4142135623730951, 1.4142135623730951,
}

var dct32 = [16]float64{
	1.9615705608064609, 1.8477590650225735, 1.6629392246050907, 1.4142135623730951,
	1.1111404660392046, 0.7653668647301797, 0.3901806440322566, 1.9615705608064609,
	1.8477590650225735, 1.6629392246050907, 1.4142135623730951, 1.1111404660392046,
	0.7653668647301797, 0.3901806440322566, 1.9615705608064609, 1.8477590650225735,
}

var dct16 = [8]float64{
	1.9615705608064609, 1.8477590650225735, 1.6629392246050907, 1.4142135623730951,
	1.1111404660392046, 0.7653668647301797, 0.3901806440322566, 1.9615705608064609,
}

func init() {
	// Compute actual DCT tables at initialization
	// For DCT-64
	for i := range 32 {
		dct64[i] = math.Cos((float64(i)+0.5)*math.Pi/float64(64)) * 2
	}
	// For DCT-32
	for i := range 16 {
		dct32[i] = math.Cos((float64(i)+0.5)*math.Pi/float64(32)) * 2
	}
	// For DCT-16
	for i := range 8 {
		dct16[i] = math.Cos((float64(i)+0.5)*math.Pi/float64(16)) * 2
	}
}
