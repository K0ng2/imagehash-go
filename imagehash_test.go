package imagehashgo

import (
	"image"
	"image/color"
	"testing"
)

func TestAverageHash_Solid(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			img.Set(x, y, color.RGBA{R: 100, G: 100, B: 100, A: 255})
		}
	}

	hash := AverageHash(img, 8)
	// For solid color, all bits should be 0 (false) because pixels are not GREATER than avg
	for i, b := range hash.hash {
		if b {
			t.Errorf("Bit %d should be false for solid image", i)
		}
	}

	expectedHex := "0000000000000000"
	if hash.ToString() != expectedHex {
		t.Errorf("Expected hex %s, got %s", expectedHex, hash.ToString())
	}
}

func TestDistance(t *testing.T) {
	h1 := NewImageHash([]bool{true, true, false, false}, 2, 2)
	h2 := NewImageHash([]bool{true, false, true, false}, 2, 2)

	dist, err := h1.Distance(h2)
	if err != nil {
		t.Fatal(err)
	}
	if dist != 2 {
		t.Errorf("Expected distance 2, got %d", dist)
	}
}
