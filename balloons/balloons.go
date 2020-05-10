package main

import (
	"image/png"
	"log"
	"os"

	"github.com/veandco/go-sdl2/sdl"
)

const (
	H int = 600
	W int = 800
)

type texture struct {
	pos
	pixels      []byte
	w, h, pitch int
}

type pos struct {
	x, y float32
}

//纹理的渲染
func (tex *texture) draw(pixels []byte) {
	for y := 0; y < tex.h; y++ {
		for x := 0; x < tex.w; x++ {
			screenY := y + int(tex.y)
			screenX := x + int(tex.x)
			if screenX >= 0 && screenX < W && screenY >= 0 && screenY < H {
				texIndex := y*tex.pitch + x*4
				screenIndex := screenY*W*4 + screenX*4
				pixels[screenIndex] = tex.pixels[texIndex]
				pixels[screenIndex+1] = tex.pixels[texIndex+1]
				pixels[screenIndex+2] = tex.pixels[texIndex+2]
				pixels[screenIndex+3] = tex.pixels[texIndex+3]
			}
		}
	}
}
func (tex *texture) drawAlpha(pixels []byte) {
	for y := 0; y < tex.h; y++ {
		for x := 0; x < tex.w; x++ {
			screenY := y + int(tex.y)
			screenX := x + int(tex.x)
			if screenX >= 0 && screenX < W && screenY >= 0 && screenY < H {
				texIndex := y*tex.pitch + x*4
				screenIndex := screenY*W*4 + screenX*4
				srcR := int(tex.pixels[texIndex])
				srcG := int(tex.pixels[texIndex+1])
				srcB := int(tex.pixels[texIndex+2])
				srcA := int(tex.pixels[texIndex+3])
				dstR := int(pixels[screenIndex])
				dstG := int(pixels[screenIndex+1])
				dstB := int(pixels[screenIndex+2])

				rstR := (srcR*255 + dstR*(255-srcA)) / 255
				rstG := (srcG*255 + dstG*(255-srcA)) / 255
				rstB := (srcB*255 + dstB*(255-srcA)) / 255

				pixels[screenIndex] = byte(rstR)
				pixels[screenIndex+1] = byte(rstG)
				pixels[screenIndex+2] = byte(rstB)
			}
		}
	}
}

func loadBalloons() []texture {
	balloonStrs := []string{"balloon_blue.png", "balloon_green.png", "balloon_red.png"}
	balloonTextures := make([]texture, 0)
	for _, value := range balloonStrs {
		file, e := os.Open(value)
		if e != nil {
			panic(e)
		}
		image, e := png.Decode(file)
		if e != nil {
			panic(e)
		}
		_ = file.Close()
		w := image.Bounds().Max.X
		h := image.Bounds().Max.Y
		balloonsPixels := make([]byte, w*h*4)
		bIndex := 0
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				r, g, b, a := image.At(x, y).RGBA()
				balloonsPixels[bIndex] = byte(r / 256)
				bIndex++
				balloonsPixels[bIndex] = byte(g / 256)
				bIndex++
				balloonsPixels[bIndex] = byte(b / 256)
				bIndex++
				balloonsPixels[bIndex] = byte(a / 256)
				bIndex++
			}
		}
		balloonTextures = append(balloonTextures, texture{pos{20, 20}, balloonsPixels, w, h, w * 4})
	}

	return balloonTextures
}

func clear(pixels []byte) {
	for i := range pixels {
		pixels[i] = 0
	}
}
func main() {
	err := sdl.Init(sdl.INIT_EVERYTHING)
	if err != nil {
		log.Fatal(err)
		return
	}
	wind, e := sdl.CreateWindow("game_demo", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		int32(W), int32(H), sdl.WINDOW_SHOWN)
	if e != nil {
		log.Fatal(e)
	}
	defer wind.Destroy()
	renderer, err := sdl.CreateRenderer(wind, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		log.Fatal(err)
	}
	defer renderer.Destroy()
	tex, err := renderer.CreateTexture(sdl.PIXELFORMAT_ABGR8888, sdl.TEXTUREACCESS_STREAMING, int32(W), int32(H))
	if err != nil {
		log.Fatal(err)
		return
	}
	defer tex.Destroy()

	pixels := make([]byte, W*H*4)
	balloonsTexs := loadBalloons()
	dir := 1
	for {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event.(type) {
			case *sdl.QuitEvent:
				return
			}
		}
		clear(pixels)
		for _, Tex := range balloonsTexs {
			Tex.drawAlpha(pixels)
		}

		balloonsTexs[1].x += float32(2 * dir)
		if balloonsTexs[1].x > 500 || balloonsTexs[1].x < 20 {
			dir = -dir
		}
		_ = tex.Update(nil, pixels, W*4)
		_ = renderer.Copy(tex, nil, nil)
		renderer.Present()
		sdl.Delay(16)
	}
}
