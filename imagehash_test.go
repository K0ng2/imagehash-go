package imagehashgo

import (
	"image"
	"image/color"
	_ "image/png"
	"os"
	"testing"
)

func TestImagePng(t *testing.T) {
	file, err := os.Open("image.png")
	if err != nil {
		t.Skip("image.png not found, skipping file-based test")
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		t.Fatalf("Failed to decode image.png: %v", err)
	}

	tests := []struct {
		name     string
		algo     func(image.Image) *ImageHash
		expected string
	}{
		{
			name:     "AverageHash",
			algo:     func(i image.Image) *ImageHash { return AverageHash(i, 8) },
			expected: "ffefc3c3c3c3c3e7",
		},
		{
			name:     "PerceptualHash",
			algo:     func(i image.Image) *ImageHash { return PerceptualHash(i, 8, 4) },
			expected: "b19b9768cc64cc66",
		},
		{
			name:     "DifferenceHash",
			algo:     func(i image.Image) *ImageHash { return DifferenceHash(i, 8) },
			expected: "12189e3333968e0c",
		},
		{
			name:     "DifferenceHashVertical",
			algo:     func(i image.Image) *ImageHash { return DifferenceHashVertical(i, 8) },
			expected: "04828010426626bd",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := tt.algo(img)
			if hash.ToString() != tt.expected {
				t.Errorf("%s got %s, want %s", tt.name, hash.ToString(), tt.expected)
			}
		})
	}
}

func TestImageHash_Distance(t *testing.T) {
	tests := []struct {
		name     string
		h1       []bool
		h2       []bool
		expected int
		wantErr  bool
	}{
		{
			name:     "identical hashes",
			h1:       []bool{true, false, true, false},
			h2:       []bool{true, false, true, false},
			expected: 0,
			wantErr:  false,
		},
		{
			name:     "different hashes",
			h1:       []bool{true, false, true, false},
			h2:       []bool{false, true, false, true},
			expected: 4,
			wantErr:  false,
		},
		{
			name:     "partial overlap",
			h1:       []bool{true, true, false, false},
			h2:       []bool{true, false, true, false},
			expected: 2,
			wantErr:  false,
		},
		{
			name:    "mismatched size",
			h1:      []bool{true, true},
			h2:      []bool{true, true, false},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			size1 := len(tt.h1)
			size2 := len(tt.h2)
			hash1 := &ImageHash{hash: tt.h1, rows: size1, cols: 1}
			hash2 := &ImageHash{hash: tt.h2, rows: size2, cols: 1}

			dist, err := hash1.Distance(hash2)
			if (err != nil) != tt.wantErr {
				t.Errorf("Distance() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && dist != tt.expected {
				t.Errorf("Distance() = %v, expected %v", dist, tt.expected)
			}
		})
	}
}

func TestImageHash_ToString_RoundTrip(t *testing.T) {
	tests := []struct {
		name string
		hash []bool
		rows int
		cols int
	}{
		{
			name: "8x8 hash",
			hash: make([]bool, 64),
			rows: 8,
			cols: 8,
		},
		{
			name: "random pattern",
			hash: []bool{true, false, true, true, false, true, false, false},
			rows: 2,
			cols: 4,
		},
	}

	// Initialize some values for random pattern
	tests[0].hash[0] = true
	tests[0].hash[31] = true
	tests[0].hash[63] = true

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &ImageHash{hash: tt.hash, rows: tt.rows, cols: tt.cols}
			s := h.ToString()
			h2, err := HexToHash(s)
			if err != nil {
				t.Fatalf("HexToHash() error = %v", err)
			}
			if len(h.hash) != len(h2.hash) {
				t.Errorf("Round-trip failed: got length %d, want %d", len(h2.hash), len(h.hash))
			}
			for i := range h.hash {
				if h.hash[i] != h2.hash[i] {
					t.Errorf("Round-trip failed at bit %d: got %v, want %v", i, h2.hash[i], h.hash[i])
				}
			}
		})
	}
}

func TestAverageHash_SolidColor(t *testing.T) {
	// Create a red image
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for y := range 100 {
		for x := range 100 {
			img.Set(x, y, color.RGBA{255, 0, 0, 255})
		}
	}

	hash := AverageHash(img, 8)
	// For solid color, pixel value equals average, so it's NOT greater than average.
	// All bits should be false.
	for i, b := range hash.hash {
		if b {
			t.Errorf("AverageHash bit %d should be false for solid color", i)
		}
	}
}

func TestHashingAlgorithms_SmokeTest(t *testing.T) {
	// Create a simple gradient image
	img := image.NewRGBA(image.Rect(0, 0, 16, 16))
	for y := range 16 {
		for x := range 16 {
			c := uint8(x * 16)
			img.Set(x, y, color.RGBA{c, c, c, 255})
		}
	}

	// Just ensure they don't panic and return something reasonable
	t.Run("AverageHash", func(t *testing.T) {
		h := AverageHash(img, 8)
		if len(h.hash) != 64 {
			t.Errorf("Expected 64 bits, got %d", len(h.hash))
		}
	})

	t.Run("PerceptualHash", func(t *testing.T) {
		h := PerceptualHash(img, 8, 4)
		if len(h.hash) != 64 {
			t.Errorf("Expected 64 bits, got %d", len(h.hash))
		}
	})

	t.Run("DifferenceHash", func(t *testing.T) {
		h := DifferenceHash(img, 8)
		if len(h.hash) != 64 {
			t.Errorf("Expected 64 bits, got %d", len(h.hash))
		}
	})

	t.Run("DifferenceHashVertical", func(t *testing.T) {
		h := DifferenceHashVertical(img, 8)
		if len(h.hash) != 64 {
			t.Errorf("Expected 64 bits, got %d", len(h.hash))
		}
	})
}

func getBenchImage() image.Image {
	file, err := os.Open("image.png")
	if err == nil {
		defer file.Close()
		img, _, err := image.Decode(file)
		if err == nil {
			return img
		}
	}
	// Fallback to a generated image if image.png is not found
	img := image.NewRGBA(image.Rect(0, 0, 512, 512))
	for y := range 512 {
		for x := range 512 {
			c := uint8((x + y) % 256)
			img.Set(x, y, color.RGBA{c, c, c, 255})
		}
	}
	return img
}

func BenchmarkAverageHash(b *testing.B) {
	img := getBenchImage()

	for b.Loop() {
		AverageHash(img, 8)
	}
}

func BenchmarkPerceptualHash(b *testing.B) {
	img := getBenchImage()

	for b.Loop() {
		PerceptualHash(img, 8, 4)
	}
}

func BenchmarkDifferenceHash(b *testing.B) {
	img := getBenchImage()

	for b.Loop() {
		DifferenceHash(img, 8)
	}
}

func BenchmarkDifferenceHashVertical(b *testing.B) {
	img := getBenchImage()

	for b.Loop() {
		DifferenceHashVertical(img, 8)
	}
}
