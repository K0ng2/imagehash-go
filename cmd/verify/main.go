package main

import (
	"fmt"
	"image"
	_ "image/png"
	"os"

	imagehashgo "github.com/K0ng2/imagehash-go"
	"github.com/corona10/goimagehash"
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

	ahash1, err := goimagehash.AverageHash(img)
	if err != nil {
		panic(err)
	}
	phash1, err := goimagehash.ExtPerceptionHash(img, 8, 8)
	if err != nil {
		panic(err)
	}
	dhash1, err := goimagehash.DifferenceHash(img)
	if err != nil {
		panic(err)
	}

	fmt.Printf("ahash: %s\n", ahash.ToString())
	fmt.Printf("phash: %s\n", phash.ToString())
	fmt.Printf("dhash: %s\n", dhash.ToString())
	fmt.Printf("dhash_v: %s\n", dhashV.ToString())

	fmt.Printf("ahash1: %s\n", ahash1.ToString())
	fmt.Printf("phash1: %s\n", phash1.ToString())
	fmt.Printf("dhash1: %s\n", dhash1.ToString())
}
