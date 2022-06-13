package main

import (
	"image"
	"image/jpeg"
	"log"
	"os"
	"strconv"
	"time"
)

func saveToFile(img image.Image) error {
	fname := strconv.FormatInt(time.Now().UnixNano()/int64(time.Millisecond), 10) + ".jpg"
	f, err := os.Create(fname)
	if err != nil {
		panic(err)
	}

	defer f.Close()

	log.Println("saving", fname)
	return jpeg.Encode(f, img, &jpeg.Options{Quality: 60})
}
