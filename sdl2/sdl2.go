package main

import (
	"log"

	"github.com/veandco/go-sdl2/sdl"
)

type color struct {
	r, g, b byte
}

func setPixels(x, y int, c color, pixels []byte) {
	index := (y*800 + x) * 4
	if index < len(pixels)-4 && index >= 0 {
		pixels[index] = c.r
		pixels[index+1] = c.g
		pixels[index+2] = c.b
	}
}

func main() {
	err := sdl.Init(sdl.INIT_EVERYTHING)
	if err != nil {
		log.Fatal(err)
		return
	}
	wind, e := sdl.CreateWindow("game_demo", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		800, 600, sdl.WINDOW_SHOWN)
	if e != nil {
		log.Fatal(e)
	}
	defer wind.Destroy()
	renderer, err := sdl.CreateRenderer(wind, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		log.Fatal(err)
	}
	defer renderer.Destroy()
	tex, err := renderer.CreateTexture(sdl.PIXELFORMAT_ABGR8888, sdl.TEXTUREACCESS_STREAMING, 800, 600)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer tex.Destroy()

	pixels := make([]byte, 800*600*4)
	for y := 0; y < 600; y++ {
		for x := 0; x < 800; x++ {
			setPixels(x, y, color{byte(x % 255), 0, byte(y % 255)}, pixels)
		}
	}
	_ = tex.Update(nil, pixels, 800*4)
	_ = renderer.Copy(tex, nil, nil)
	renderer.Present()

	for {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event.(type) {
			case *sdl.QuitEvent:
				return
			}
		}

		sdl.Delay(16)
	}
}
