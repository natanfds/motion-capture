package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"os/exec"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"
)

var blockSize = 8

func CalculateDominantColor(img image.Image, startX, startY, blockSize int) color.Color {
	colorCount := make(map[color.Color]int)
	for y := startY; y < startY+blockSize && y < img.Bounds().Dy(); y++ {
		for x := startX; x < startX+blockSize && x < img.Bounds().Dx(); x++ {
			c := img.At(x, y)
			colorCount[c]++
		}
	}

	var dominantColor color.Color
	maxCount := 0
	for c, count := range colorCount {
		if count > maxCount {
			dominantColor = c
			maxCount = count
		}
	}

	return dominantColor
}

func GetColorMap(img image.Image) *image.RGBA {
	colorMap := image.NewRGBA(img.Bounds())
	for y := 0; y < img.Bounds().Dy(); y += blockSize {
		for x := 0; x < img.Bounds().Dx(); x += blockSize {
			dominantColor := CalculateDominantColor(img, x, y, blockSize)

			for by := y; by < y+blockSize && by < img.Bounds().Dy(); by++ {
				for bx := x; bx < x+blockSize && bx < img.Bounds().Dx(); bx++ {
					colorMap.Set(bx, by, dominantColor)
				}
			}
		}
	}

	return colorMap
}

func run() {

	windowCfg := pixelgl.WindowConfig{
		Title:  "Webcam",
		Bounds: pixel.R(0, 0, 640, 480),
		VSync:  true,
	}

	win, err := pixelgl.NewWindow(windowCfg)
	if err != nil {
		panic(err)
	}

	for !win.Closed() {
		cmd := exec.Command("ffmpeg", "-y", "-f", "video4linux2", "-i", "/dev/video0", "-vframes", "1", "-f", "image2pipe", "-")
		var out bytes.Buffer
		cmd.Stdout = &out

		err := cmd.Run()
		if err != nil {
			fmt.Printf("Erro ao capturar a imagem: %v\n", err)
			return
		}

		imageData := out.Bytes()

		img, err := jpeg.Decode(bytes.NewReader(imageData))

		if err != nil {
			fmt.Printf("Erro ao decodificar a imagem: %v\n", err)
			return
		}

		colorMap := GetColorMap(img)
		pic := pixel.PictureDataFromImage(colorMap)
		sprite := pixel.NewSprite(pic, pic.Bounds())

		win.Clear(colornames.Black)
		sprite.Draw(win, pixel.IM.Moved(win.Bounds().Center()))
		win.Update()

		// Esperar 1/10 de segundo
		time.Sleep(1 * time.Millisecond)

	}
}

func main() {
	pixelgl.Run(run)
}
