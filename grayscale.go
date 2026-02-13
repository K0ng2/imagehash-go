package imagehashgo

import (
	"image"
	"image/color"
	"runtime"
	"sync"
)

// ToGrayscale converts an image to a grayscale image (image.Gray)
// using the L mode formula from Pillow:
// L = R * 299/1000 + G * 587/1000 + B * 114/1000
func ToGrayscale(img image.Image) *image.Gray {
	if gray, ok := img.(*image.Gray); ok {
		return gray
	}

	bounds := img.Bounds()
	grayImg := image.NewGray(bounds)

	numCPUs := runtime.NumCPU()
	if numCPUs > 1 && bounds.Dy() > numCPUs {
		var wg sync.WaitGroup
		rowsPerWorker := bounds.Dy() / numCPUs
		if rowsPerWorker == 0 {
			rowsPerWorker = 1
		}

		for i := range numCPUs {
			startY := bounds.Min.Y + i*rowsPerWorker
			endY := startY + rowsPerWorker
			if i == numCPUs-1 {
				endY = bounds.Max.Y
			}

			if startY >= bounds.Max.Y {
				break
			}

			wg.Add(1)
			go func(sY, eY int) {
				defer wg.Done()
				for y := sY; y < eY; y++ {
					for x := bounds.Min.X; x < bounds.Max.X; x++ {
						processPixel(img, grayImg, x, y)
					}
				}
			}(startY, endY)
		}
		wg.Wait()
	} else {
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				processPixel(img, grayImg, x, y)
			}
		}
	}

	return grayImg
}

// ToGrayscaleFast is an optimized version with type-specific fast paths
// For small images (< 64x64), it avoids goroutine overhead
func ToGrayscaleFast(img image.Image) *image.Gray {
	if gray, ok := img.(*image.Gray); ok {
		return gray
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	grayImg := image.NewGray(bounds)

	// For small images, avoid goroutine overhead
	useParallel := width*height > 64*64 && runtime.NumCPU() > 1

	// Type-specific optimizations
	switch typedImg := img.(type) {
	case *image.YCbCr:
		if useParallel {
			processYCbCrParallel(typedImg, grayImg)
		} else {
			processYCbCr(typedImg, grayImg)
		}
	case *image.RGBA:
		if useParallel {
			processRGBAParallel(typedImg, grayImg)
		} else {
			processRGBA(typedImg, grayImg)
		}
	case *image.NRGBA:
		if useParallel {
			processNRGBAParallel(typedImg, grayImg)
		} else {
			processNRGBA(typedImg, grayImg)
		}
	default:
		// Fallback to generic interface
		if useParallel {
			processGenericParallel(img, grayImg)
		} else {
			processGeneric(img, grayImg)
		}
	}

	return grayImg
}

// Type-specific processors for YCbCr (common in JPEG)
func processYCbCr(src *image.YCbCr, dst *image.Gray) {
	bounds := src.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := src.YCbCrAt(x, y).RGBA()
			dst.Pix[(y-bounds.Min.Y)*dst.Stride+(x-bounds.Min.X)] = rgbaToGray(r, g, b, a)
		}
	}
}

func processYCbCrParallel(src *image.YCbCr, dst *image.Gray) {
	bounds := src.Bounds()
	numCPUs := runtime.NumCPU()
	rowsPerWorker := bounds.Dy() / numCPUs
	if rowsPerWorker == 0 {
		rowsPerWorker = 1
	}

	var wg sync.WaitGroup
	for i := range numCPUs {
		startY := bounds.Min.Y + i*rowsPerWorker
		endY := startY + rowsPerWorker
		if i == numCPUs-1 {
			endY = bounds.Max.Y
		}
		if startY >= bounds.Max.Y {
			break
		}

		wg.Add(1)
		go func(sY, eY int) {
			defer wg.Done()
			for y := sY; y < eY; y++ {
				for x := bounds.Min.X; x < bounds.Max.X; x++ {
					r, g, b, a := src.YCbCrAt(x, y).RGBA()
					dst.Pix[(y-bounds.Min.Y)*dst.Stride+(x-bounds.Min.X)] = rgbaToGray(r, g, b, a)
				}
			}
		}(startY, endY)
	}
	wg.Wait()
}

// Type-specific processors for RGBA
func processRGBA(src *image.RGBA, dst *image.Gray) {
	bounds := src.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := src.RGBAAt(x, y).RGBA()
			dst.Pix[(y-bounds.Min.Y)*dst.Stride+(x-bounds.Min.X)] = rgbaToGray(r, g, b, a)
		}
	}
}

func processRGBAParallel(src *image.RGBA, dst *image.Gray) {
	bounds := src.Bounds()
	numCPUs := runtime.NumCPU()
	rowsPerWorker := bounds.Dy() / numCPUs
	if rowsPerWorker == 0 {
		rowsPerWorker = 1
	}

	var wg sync.WaitGroup
	for i := range numCPUs {
		startY := bounds.Min.Y + i*rowsPerWorker
		endY := startY + rowsPerWorker
		if i == numCPUs-1 {
			endY = bounds.Max.Y
		}
		if startY >= bounds.Max.Y {
			break
		}

		wg.Add(1)
		go func(sY, eY int) {
			defer wg.Done()
			for y := sY; y < eY; y++ {
				for x := bounds.Min.X; x < bounds.Max.X; x++ {
					r, g, b, a := src.RGBAAt(x, y).RGBA()
					dst.Pix[(y-bounds.Min.Y)*dst.Stride+(x-bounds.Min.X)] = rgbaToGray(r, g, b, a)
				}
			}
		}(startY, endY)
	}
	wg.Wait()
}

// Type-specific processors for NRGBA (created by imaging library)
func processNRGBA(src *image.NRGBA, dst *image.Gray) {
	bounds := src.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := src.NRGBAAt(x, y).RGBA()
			dst.Pix[(y-bounds.Min.Y)*dst.Stride+(x-bounds.Min.X)] = rgbaToGray(r, g, b, a)
		}
	}
}

func processNRGBAParallel(src *image.NRGBA, dst *image.Gray) {
	bounds := src.Bounds()
	numCPUs := runtime.NumCPU()
	rowsPerWorker := bounds.Dy() / numCPUs
	if rowsPerWorker == 0 {
		rowsPerWorker = 1
	}

	var wg sync.WaitGroup
	for i := range numCPUs {
		startY := bounds.Min.Y + i*rowsPerWorker
		endY := startY + rowsPerWorker
		if i == numCPUs-1 {
			endY = bounds.Max.Y
		}
		if startY >= bounds.Max.Y {
			break
		}

		wg.Add(1)
		go func(sY, eY int) {
			defer wg.Done()
			for y := sY; y < eY; y++ {
				for x := bounds.Min.X; x < bounds.Max.X; x++ {
					r, g, b, a := src.NRGBAAt(x, y).RGBA()
					dst.Pix[(y-bounds.Min.Y)*dst.Stride+(x-bounds.Min.X)] = rgbaToGray(r, g, b, a)
				}
			}
		}(startY, endY)
	}
	wg.Wait()
}

// Generic processor using interface
func processGeneric(src image.Image, dst *image.Gray) {
	bounds := src.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			processPixel(src, dst, x, y)
		}
	}
}

func processGenericParallel(src image.Image, dst *image.Gray) {
	bounds := src.Bounds()
	numCPUs := runtime.NumCPU()
	rowsPerWorker := bounds.Dy() / numCPUs
	if rowsPerWorker == 0 {
		rowsPerWorker = 1
	}

	var wg sync.WaitGroup
	for i := range numCPUs {
		startY := bounds.Min.Y + i*rowsPerWorker
		endY := startY + rowsPerWorker
		if i == numCPUs-1 {
			endY = bounds.Max.Y
		}
		if startY >= bounds.Max.Y {
			break
		}

		wg.Add(1)
		go func(sY, eY int) {
			defer wg.Done()
			for y := sY; y < eY; y++ {
				for x := bounds.Min.X; x < bounds.Max.X; x++ {
					processPixel(src, dst, x, y)
				}
			}
		}(startY, endY)
	}
	wg.Wait()
}

// rgbaToGray converts RGBA values to grayscale using the correct formula
func rgbaToGray(r, g, b, a uint32) uint8 {
	// RGBA returns values in [0, 65535] and they are alpha-premultiplied.
	// Un-premultiply to get the raw RGB values.
	if a > 0 && a < 0xffff {
		r = (r * 0xffff) / a
		g = (g * 0xffff) / a
		b = (b * 0xffff) / a
	}

	// Convert 16-bit to 8-bit
	r8 := uint32(r >> 8)
	g8 := uint32(g >> 8)
	b8 := uint32(b >> 8)

	// Applying the formula: R*0.299 + G*0.587 + B*0.114
	// To avoid floating point, we use: (R*299 + G*587 + B*114 + 500) / 1000
	l := (r8*299 + g8*587 + b8*114 + 500) / 1000
	return uint8(l)
}

func processPixel(img image.Image, grayImg *image.Gray, x, y int) {
	r, g, b, a := img.At(x, y).RGBA()
	// RGBA returns values in [0, 65535] and they are alpha-premultiplied.
	// Pillow's 'L' conversion ignores alpha, so we should un-premultiply
	// to get the raw RGB values.

	if a > 0 && a < 0xffff {
		r = (r * 0xffff) / a
		g = (g * 0xffff) / a
		b = (b * 0xffff) / a
	}

	// Convert 16-bit to 8-bit
	r8 := uint32(r >> 8)
	g8 := uint32(g >> 8)
	b8 := uint32(b >> 8)

	// Applying the formula: R*0.299 + G*0.587 + B*0.114
	// To avoid floating point, we can use: (R*299 + G*587 + B*114 + 500) / 1000
	l := (r8*299 + g8*587 + b8*114 + 500) / 1000
	grayImg.SetGray(x, y, color.Gray{Y: uint8(l)})
}
