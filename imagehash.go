package imagehashgo

import (
	"fmt"
	"image"
	"math"

	"github.com/disintegration/imaging"
)

// ImageHash represents an image hash
type ImageHash struct {
	hash []bool
	rows int
	cols int
}

// NewImageHash creates a new ImageHash
func NewImageHash(hash []bool, rows, cols int) *ImageHash {
	return &ImageHash{
		hash: hash,
		rows: rows,
		cols: cols,
	}
}

// Distance returns the Hamming distance between this hash and another
func (h *ImageHash) Distance(other *ImageHash) (int, error) {
	if h.rows != other.rows || h.cols != other.cols {
		return 0, fmt.Errorf("ImageHashes must be of the same shape: (%d, %d) vs (%d, %d)", h.rows, h.cols, other.rows, other.cols)
	}

	dist := 0
	for i := range h.hash {
		if h.hash[i] != other.hash[i] {
			dist++
		}
	}
	return dist, nil
}

// ToString returns the hex string representation of the hash
func (h *ImageHash) ToString() string {
	if len(h.hash) == 0 {
		return ""
	}

	// Convert bit array to big integer bits
	// We want to match Python's approach: int(bit_string, 2)
	// where the first bit is the most significant.
	// But actually, in Python's _binary_array_to_hex, it's:
	// bit_string = ''.join(str(b) for b in 1 * arr.flatten())
	// int(bit_string, 2)
	// This means the last bit of the array is the least significant bit of the integer.

	hexLen := (len(h.hash) + 3) / 4
	result := make([]byte, hexLen)

	for i := range hexLen {
		var val uint8
		for j := range 4 {
			bitIdx := i*4 + j
			if bitIdx < len(h.hash) && h.hash[bitIdx] {
				val |= 1 << (3 - uint(j))
			}
		}
		if val < 10 {
			result[i] = '0' + val
		} else {
			result[i] = 'a' + (val - 10)
		}
	}

	return string(result)
}

// HexToHash converts a hex string back to an ImageHash
func HexToHash(hexStr string) (*ImageHash, error) {
	bitsPerHex := 4
	totalBits := len(hexStr) * bitsPerHex
	hashSize := int(math.Sqrt(float64(totalBits)))
	// Check if it's a square
	if hashSize*hashSize != totalBits {
		// Not a square, maybe it's just a flat hash
		// For now, assume square as most imagehashes are
	}

	hash := make([]bool, totalBits)
	for i, r := range hexStr {
		var val uint8
		if r >= '0' && r <= '9' {
			val = uint8(r - '0')
		} else if r >= 'a' && r <= 'f' {
			val = uint8(r - 'a' + 10)
		} else if r >= 'A' && r <= 'F' {
			val = uint8(r - 'A' + 10)
		} else {
			return nil, fmt.Errorf("invalid hex character: %c", r)
		}

		for j := range 4 {
			if (val & (1 << (3 - uint(j)))) != 0 {
				hash[i*4+j] = true
			}
		}
	}

	return &ImageHash{
		hash: hash,
		rows: hashSize,
		cols: hashSize,
	}, nil
}

// AverageHash computes the Average Hash of an image
func AverageHash(img image.Image, hashSize int) *ImageHash {
	if hashSize < 2 {
		hashSize = 8
	}

	// 1. Convert to grayscale
	gray := ToGrayscale(img)

	// 2. Resize to hashSize x hashSize
	resized := imaging.Resize(gray, hashSize, hashSize, imaging.Lanczos)
	// imaging.Resize returns *image.NRGBA, convert to grayscale pixels
	grayResized := ToGrayscale(resized)

	// 3. Compute average pixel value
	var sum uint64
	for y := range hashSize {
		for x := range hashSize {
			sum += uint64(grayResized.Pix[y*grayResized.Stride+x])
		}
	}
	avg := float64(sum) / float64(hashSize*hashSize)

	// 4. Create hash
	hash := make([]bool, hashSize*hashSize)
	for y := range hashSize {
		for x := range hashSize {
			hash[y*hashSize+x] = float64(grayResized.Pix[y*grayResized.Stride+x]) > avg
		}
	}

	return &ImageHash{
		hash: hash,
		rows: hashSize,
		cols: hashSize,
	}
}

// DifferenceHash computes the Difference Hash of an image
func DifferenceHash(img image.Image, hashSize int) *ImageHash {
	if hashSize < 2 {
		hashSize = 8
	}

	// 1. Convert to grayscale
	gray := ToGrayscale(img)

	// 2. Resize to (hashSize + 1) x hashSize
	resized := imaging.Resize(gray, hashSize+1, hashSize, imaging.Lanczos)
	grayResized := ToGrayscale(resized)

	// 3. Compute differences between columns
	pixels := grayResized.Pix
	// grayResized has hashSize+1 columns and hashSize rows
	hash := make([]bool, hashSize*hashSize)
	for y := range hashSize {
		for x := range hashSize {
			// p[x, y] vs p[x+1, y]
			left := pixels[y*grayResized.Stride+x]
			right := pixels[y*grayResized.Stride+x+1]
			hash[y*hashSize+x] = right > left
		}
	}

	return &ImageHash{
		hash: hash,
		rows: hashSize,
		cols: hashSize,
	}
}

// DifferenceHashVertical computes the vertical Difference Hash of an image
func DifferenceHashVertical(img image.Image, hashSize int) *ImageHash {
	if hashSize < 2 {
		hashSize = 8
	}

	// 1. Convert to grayscale
	gray := ToGrayscale(img)

	// 2. Resize to hashSize x (hashSize + 1)
	resized := imaging.Resize(gray, hashSize, hashSize+1, imaging.Lanczos)
	grayResized := ToGrayscale(resized)

	// 3. Compute differences between rows
	pixels := grayResized.Pix
	hash := make([]bool, hashSize*hashSize)
	for y := range hashSize {
		for x := range hashSize {
			// p[x, y] vs p[x, y+1]
			top := pixels[y*grayResized.Stride+x]
			bottom := pixels[(y+1)*grayResized.Stride+x]
			hash[y*hashSize+x] = bottom > top
		}
	}

	return &ImageHash{
		hash: hash,
		rows: hashSize,
		cols: hashSize,
	}
}

// PerceptualHash computes the Perceptual Hash of an image
func PerceptualHash(img image.Image, hashSize int, highfreqFactor int) *ImageHash {
	if hashSize < 2 {
		hashSize = 8
	}
	if highfreqFactor < 1 {
		highfreqFactor = 4
	}

	imgSize := hashSize * highfreqFactor

	// 1. Convert to grayscale
	gray := ToGrayscale(img)

	// 2. Resize to imgSize x imgSize
	resized := imaging.Resize(gray, imgSize, imgSize, imaging.Lanczos)
	grayResized := ToGrayscale(resized)

	// 3. Compute 2D DCT
	pixels := grayResized.Pix
	matrix := make([][]float64, imgSize)
	for y := range imgSize {
		matrix[y] = make([]float64, imgSize)
		for x := range imgSize {
			matrix[y][x] = float64(pixels[y*grayResized.Stride+x])
		}
	}

	dct := DCT2D(matrix)

	// 4. Extract low frequency part (hashSize x hashSize)
	dctLowFreq := make([]float64, hashSize*hashSize)
	for y := range hashSize {
		for x := range hashSize {
			dctLowFreq[y*hashSize+x] = dct[y][x]
		}
	}

	// 5. Compute median
	med := median(dctLowFreq)

	// 6. Create hash
	hash := make([]bool, hashSize*hashSize)
	for i, val := range dctLowFreq {
		hash[i] = val > med
	}

	return &ImageHash{
		hash: hash,
		rows: hashSize,
		cols: hashSize,
	}
}

func median(data []float64) float64 {
	length := len(data)
	if length == 0 {
		return 0
	}

	// Make a copy to avoid modifying original data
	sorted := make([]float64, length)
	copy(sorted, data)

	// Simple sort (for small data sets like 8x8 it's okay)
	for i := range length {
		for j := i + 1; j < length; j++ {
			if sorted[i] > sorted[j] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	if length%2 == 0 {
		return (sorted[length/2-1] + sorted[length/2]) / 2
	}
	return sorted[length/2]
}
