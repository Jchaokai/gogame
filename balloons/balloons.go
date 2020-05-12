package main

import (
	"gogame/noise"
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
	//每个气球有自己缩放比例
	scale float32
}

type pos struct {
	x, y float32
}
type rgba struct {
	r, g, b byte
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

func flerp(a, b, pct float32) float32 {
	return a + (b-a)*pct
}
func blerp(c00, c01, c10, c11, tx, ty float32) float32 {
	return flerp(flerp(c00, c10, tx), flerp(c01, c11, tx), ty)
}

//气球比例缩放
func (tex *texture) drawScaled(scaleX, scaleY float32, pixels []byte) {
	newWidth := int(float32(tex.w) * scaleX)
	newHeight := int(float32(tex.h) * scaleY)
	texW4 := tex.w * 4
	for y := 0; y < newHeight; y++ {
		fy := float32(y) / float32(newHeight) * float32(tex.h-1)
		fyi := int(fy)
		screenY := int(fy*scaleY) + int(tex.y)
		screenIndex := screenY*W*4 + int(tex.x)*4
		ty := fy - float32(fyi)
		for x := 0; x < newWidth; x++ {
			fx := float32(x) / float32(newWidth) * float32(tex.w-1)
			screenX := int(fx*scaleX) + int(tex.x)
			if screenX >= 0 && screenX < W && screenY >= 0 && screenY < H {
				fxi := int(fx)
				c00i := fyi*texW4 + fxi*4
				c01i := fyi*texW4 + (fxi+1)*4
				c10i := (fyi+1)*texW4 + fxi*4
				c11i := (fyi+1)*texW4 + (fxi+1)*4

				tx := fx - float32(fxi)
				for i := 0; i < 4; i++ {
					c00 := float32(tex.pixels[c00i+i])
					c01 := float32(tex.pixels[c01i+i])
					c10 := float32(tex.pixels[c10i+i])
					c11 := float32(tex.pixels[c11i+i])
					pixels[screenIndex] = byte(blerp(c00, c01, c10, c11, tx, ty))
					screenIndex++
				}
			}
		}
	}
}

func loadBalloons() []texture {
	balloonStrs := []string{"balloon_blue.png", "balloon_green.png", "balloon_red.png"}
	balloonTextures := make([]texture, 0)
	for i, value := range balloonStrs {
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
		balloonTextures = append(balloonTextures, texture{pos{float32(i * 60), float32(i * 60)}, balloonsPixels, w, h, w * 4, float32(i + 1)})
	}

	return balloonTextures
}

func lerp(b1, b2 byte, pct float32) byte {
	return byte(float32(b1) + pct*(float32(b2)-float32(b1)))
}

func colorLerp(c1, c2 rgba, pct float32) rgba {
	return rgba{lerp(c1.r, c2.r, pct), lerp(c1.g, c2.g, pct), lerp(c1.b, c2.b, pct)}
}

func getGradient(c1, c2 rgba) []rgba {
	res := make([]rgba, 256)
	for i := range res {
		pct := float32(i) / float32(255)
		res[i] = colorLerp(c1, c2, pct)
	}
	return res
}

func clamp(min, max, v int) int {
	if v < min {
		v = min
	} else if v > max {
		v = max
	}
	return v
}

func rescaleAndDraw(noise []float32, min, max float32, gradient []rgba, w, h int) []byte {
	res := make([]byte, w*h*4)
	scale := 255.0 / (max - min)
	offset := min * scale
	for i := range noise {
		noise[i] = noise[i]*scale - offset
		c := gradient[clamp(0, 255, int(noise[i]))]
		res[i*4] = c.r
		res[i*4+1] = c.g
		res[i*4+2] = c.b
	}
	return res
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
	//--绘制天空背景图
	cloudNoises, min, max := noise.MakeNoise(noise.FBM, 0.01, 0.2, 2, 3, W, H)
	cloudGradient := getGradient(rgba{0, 0, 255}, rgba{255, 255, 255})
	cloudPixels := rescaleAndDraw(cloudNoises, min, max, cloudGradient, W, H)
	cloudTexture := texture{pos{0, 0}, cloudPixels, W, H, W * 4, 1.0}
	//--
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

		cloudTexture.draw(pixels)
		//clear(pixels)
		for _, Tex := range balloonsTexs {
			//Tex.drawAlpha(pixels)
			Tex.drawScaled(Tex.scale, Tex.scale, pixels)
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
