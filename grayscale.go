package imagehashgo

import (
	"image"
	"image/color"
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

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
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
	}

	return grayImg
}
