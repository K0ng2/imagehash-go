package imagehashgo

import (
	"fmt"
	"image"
	"math"
	"sort"
	"sync"

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

	// 1. Convert to grayscale using fast path
	gray := ToGrayscaleFast(img)

	// 2. Resize to hashSize x hashSize
	resized := imaging.Resize(gray, hashSize, hashSize, imaging.Lanczos)
	// imaging.Resize returns *image.NRGBA, convert to grayscale pixels
	grayResized := ToGrayscaleFast(resized)

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

	// 1. Convert to grayscale using fast path
	gray := ToGrayscaleFast(img)

	// 2. Resize to (hashSize + 1) x hashSize
	resized := imaging.Resize(gray, hashSize+1, hashSize, imaging.Lanczos)
	grayResized := ToGrayscaleFast(resized)

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

	// 1. Convert to grayscale using fast path
	gray := ToGrayscaleFast(img)

	// 2. Resize to hashSize x (hashSize + 1)
	resized := imaging.Resize(gray, hashSize, hashSize+1, imaging.Lanczos)
	grayResized := ToGrayscaleFast(resized)

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

// Memory pools for pixel buffers
var (
	pixelPool32 = sync.Pool{
		New: func() any {
			p := make([]float64, 32*32)
			return &p
		},
	}
	pixelPool64 = sync.Pool{
		New: func() any {
			p := make([]float64, 64*64)
			return &p
		},
	}
)

// PerceptualHash computes the Perceptual Hash of an image
func PerceptualHash(img image.Image, hashSize int, highfreqFactor int) *ImageHash {
	if hashSize < 2 {
		hashSize = 8
	}
	if highfreqFactor < 1 {
		highfreqFactor = 4
	}

	imgSize := hashSize * highfreqFactor

	// Use optimized fast DCT for common sizes
	if imgSize == 32 && hashSize == 8 {
		return perceptualHashFast32(img)
	} else if imgSize == 64 && hashSize == 8 {
		return perceptualHashFast64(img)
	}

	// Fallback to general implementation for other sizes
	// 1. Convert to grayscale using fast path
	gray := ToGrayscaleFast(img)

	// 2. Resize to imgSize x imgSize
	resized := imaging.Resize(gray, imgSize, imgSize, imaging.Lanczos)
	grayResized := ToGrayscaleFast(resized)

	// 3. Compute 2D DCT
	pixels := grayResized.Pix
	matrix := make([][]float64, imgSize)
	for y := range imgSize {
		matrix[y] = make([]float64, imgSize)
		rowStride := y * grayResized.Stride
		for x := range imgSize {
			matrix[y][x] = float64(pixels[rowStride+x])
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

// perceptualHashFast64 uses optimized DCT for 64x64 -> 8x8 hash (default params)
func perceptualHashFast64(img image.Image) *ImageHash {
	// 1. Convert to grayscale using fast path
	gray := ToGrayscaleFast(img)

	// 2. Resize to 64x64
	resized := imaging.Resize(gray, 64, 64, imaging.Lanczos)
	grayResized := ToGrayscaleFast(resized)

	// 3. Get pixel buffer from pool
	pixelsPtr := pixelPool64.Get().(*[]float64)
	defer pixelPool64.Put(pixelsPtr)
	pixels := *pixelsPtr

	// 4. Copy image data to buffer
	pix := grayResized.Pix
	for i := range 64 {
		rowStride := i * grayResized.Stride
		for j := range 64 {
			pixels[i*64+j] = float64(pix[rowStride+j])
		}
	}

	// 5. Compute fast DCT (returns 8x8 low freq coefficients)
	dctLowFreq := DCT2DFast64(pixelsPtr)

	// 6. Compute median
	med := medianFast64(dctLowFreq[:])

	// 7. Create hash
	hash := make([]bool, 64)
	for i, val := range dctLowFreq {
		hash[i] = val > med
	}

	return &ImageHash{
		hash: hash,
		rows: 8,
		cols: 8,
	}
}

// perceptualHashFast32 uses optimized DCT for 32x32 -> 8x8 hash
func perceptualHashFast32(img image.Image) *ImageHash {
	// 1. Convert to grayscale using fast path
	gray := ToGrayscaleFast(img)

	// 2. Resize to 32x32
	resized := imaging.Resize(gray, 32, 32, imaging.Lanczos)
	grayResized := ToGrayscaleFast(resized)

	// 3. Get pixel buffer from pool
	pixelsPtr := pixelPool32.Get().(*[]float64)
	defer pixelPool32.Put(pixelsPtr)
	pixels := *pixelsPtr

	// 4. Copy image data to buffer
	pix := grayResized.Pix
	for i := range 32 {
		rowStride := i * grayResized.Stride
		for j := range 32 {
			pixels[i*32+j] = float64(pix[rowStride+j])
		}
	}

	// 5. Compute fast DCT (returns 8x8 low freq coefficients)
	dctLowFreq := DCT2DFast32(pixelsPtr, 8)

	// 6. Compute median
	med := median(dctLowFreq)

	// 7. Create hash
	hash := make([]bool, 64)
	for i, val := range dctLowFreq {
		hash[i] = val > med
	}

	return &ImageHash{
		hash: hash,
		rows: 8,
		cols: 8,
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

	sort.Float64s(sorted)

	if length%2 == 0 {
		return (sorted[length/2-1] + sorted[length/2]) / 2
	}
	return sorted[length/2]
}

// medianFast64 is optimized for fixed-size 64-element array
func medianFast64(data []float64) float64 {
	if len(data) != 64 {
		return median(data)
	}

	// Make a copy to avoid modifying original data
	var sorted [64]float64
	copy(sorted[:], data)

	sort.Float64s(sorted[:])

	// For even length (64), return average of middle two elements
	return (sorted[31] + sorted[32]) / 2
}
