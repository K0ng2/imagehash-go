# imagehash-go

A perceptual image hashing library for Go.

This project is a port of the popular Python [imagehash](https://github.com/jgraving/imagehash) library, ported to Golang by AI (Gemini 3 Flash). It aims to provide identical hashing results for the supported algorithms.

## Features

- **Average Hash (`ahash`)**: Extremely fast but less robust.
- **Perceptual Hash (`phash`)**: Robust against many image modifications.
- **Difference Hash (`dhash`)**: Fast and relatively robust.
- **Vertical Difference Hash (`dhash_v`)**: Specialized for certain image types.

## Installation

```bash
go get github.com/K0ng2/imagehash-go
```

## Usage

```go
package main

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"

	"github.com/K0ng2/imagehash-go"
)

func main() {
	file, err := os.Open("image.png")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		panic(err)
	}

	// Calculate hashes
	ahash := imagehashgo.AverageHash(img, 8)
	phash := imagehashgo.PerceptualHash(img, 8, 4)
	dhash := imagehashgo.DifferenceHash(img, 8)

	fmt.Printf("Average Hash: %s\n", ahash.ToString())
	fmt.Printf("Perceptual Hash: %s\n", phash.ToString())
	fmt.Printf("Difference Hash: %s\n", dhash.ToString())

	// Calculate Hamming distance
	distance, _ := ahash.Distance(phash)
	fmt.Printf("Distance: %d\n", distance)
}
```

## Supported Algorithms

Currently, this library supports the core algorithms found in the original Python library:
- [x] Average Hashing
- [x] Perceptual Hashing (DCT-based)
- [x] Difference Hashing (Horizontal & Vertical)

> [!NOTE]
> `whash` (Wavelet Hashing) and `colorhash` are not currently supported due to their complex dependencies.

## Testing

Run the Go tests to ensure everything is working as expected:

```bash
go test -v .
```

The tests include verification against known hashes from the original Python implementation using a sample image.

## Credits

- Original Python library: [jgraving/imagehash](https://github.com/jgraving/imagehash)
- Ported to Go by AI (Gemini 3 Flash)
