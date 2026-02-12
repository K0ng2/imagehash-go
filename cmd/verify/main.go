package main

import (
	"fmt"
	"image"
	_ "image/png"
	"os"

	imagehashgo "github.com/K0ng2/imagehash-go"
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

	ahash := imagehashgo.AverageHash(img, 8)
	phash := imagehashgo.PerceptualHash(img, 8, 4)
	dhash := imagehashgo.DifferenceHash(img, 8)
	dhashV := imagehashgo.DifferenceHashVertical(img, 8)

	fmt.Printf("ahash: %s\n", ahash.ToString())
	fmt.Printf("phash: %s\n", phash.ToString())
	fmt.Printf("dhash: %s\n", dhash.ToString())
	fmt.Printf("dhash_v: %s\n", dhashV.ToString())
}
