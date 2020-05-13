package main

//1.我们不再自己手动绘制texture，直接使用sdl2 自带的texture
//2.并使用GPU渲染
//3.使用仅有的三个素材，渲染出50个气球，并使用package vector3下的向量代替原有的pos
//4.移动气球

import (
	"fmt"
	"gogame/noise"
	. "gogame/vector3" //有了. 不需要加包名引用
	"image/png"
	"log"
	"math/rand"
	"os"
	"sort"
	"time"

	"github.com/veandco/go-sdl2/sdl"
)

const (
	H int = 600
	W int = 800
	D int = 100 //三维坐标中的深度
)

type balloon struct {
	tex *sdl.Texture
	//不再使用pos，使用vector
	pos Vector3 //位置
	dir Vector3 //方向
	//每个气球有自己缩放比例
	//scale float32
	w, h int
}

func (balloon *balloon) draw(render *sdl.Renderer) {
	scale := (balloon.pos.Z/200 + 0.5) / 2
	newW := int32(float32(balloon.w) * scale)
	newH := int32(float32(balloon.h) * scale)
	x := int32(balloon.pos.X - float32(newW)/2)
	y := int32(balloon.pos.Y - float32(newH)/2)
	rect := &sdl.Rect{X: x, Y: y, W: newW, H: newH}
	_ = render.Copy(balloon.tex, nil, rect)

}

//更新移动气球
func (balloon *balloon) Update(elapsedTime float32) {
	//possible position 原有的位置向量 + 方向向量 * 时间
	p := Add(balloon.pos, Mult(balloon.dir, elapsedTime))
	if p.X < 0 || p.X > float32(W) {
		balloon.dir.X = -balloon.dir.X
	}
	if p.Y < 0 || p.X > float32(H) {
		balloon.dir.Y = -balloon.dir.Y
	}
	if p.Z < 0 || p.Z > float32(D) {
		//气球会忽大忽小
		balloon.dir.Z = -balloon.dir.Z
	}

	balloon.pos = Add(balloon.pos, Mult(balloon.dir, elapsedTime))

}

//不再使用pos，使用vector
//type pos struct {
//	x, y float32
//}

type rgba struct {
	r, g, b byte
}

func pixelsToTexture(render *sdl.Renderer, pixels []byte, w, h int) *sdl.Texture {
	tex, err := render.CreateTexture(sdl.PIXELFORMAT_ABGR8888, sdl.TEXTUREACCESS_STREAMING, int32(w), int32(h))
	if err != nil {
		panic(err)
	}
	_ = tex.Update(nil, pixels, w*4)
	return tex
}

func loadBalloons(render *sdl.Renderer, numBalloons int) []*balloon {
	balloonStrs := []string{"balloon_blue.png", "balloon_green.png", "balloon_red.png"}
	balloonsTexture := make([]*sdl.Texture, len(balloonStrs))
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
		tex := pixelsToTexture(render, balloonsPixels, w, h)
		err := tex.SetBlendMode(sdl.BLENDMODE_BLEND)
		if err != nil {
			panic(err) //可能用户的硬件不支持
		}
		balloonsTexture[i] = tex
	}
	balloons := make([]*balloon, numBalloons)
	for i := range balloons {
		tex := balloonsTexture[i%3]
		//生成随机位置
		pos := Vector3{X: rand.Float32() * float32(W), Y: rand.Float32() * float32(H), Z: rand.Float32() * float32(D)}
		dir := Vector3{X: rand.Float32() * 0.12, Y: rand.Float32() * 0.12, Z: rand.Float32() * 0.12}
		_, _, width, height, err := tex.Query()
		if err != nil {
			panic(err)
		}
		balloons[i] = &balloon{tex, pos, dir, int(width), int(height)}
	}
	return balloons
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
	sdl.SetHint(sdl.HINT_RENDER_SCALE_QUALITY, "1")

	//--绘制天空背景图
	cloudNoises, min, max := noise.MakeNoise(noise.FBM, 0.01, 0.2, 2, 3, W, H)
	cloudGradient := getGradient(rgba{0, 0, 255}, rgba{255, 255, 255})
	cloudPixels := rescaleAndDraw(cloudNoises, min, max, cloudGradient, W, H)
	cloudTexture := pixelsToTexture(renderer, cloudPixels, W, H)
	//--
	balloons := loadBalloons(renderer, 20)
	var elapsedTime float32
	for {
		frameStart := time.Now()
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event.(type) {
			case *sdl.QuitEvent:
				return
			}
		}
		_ = renderer.Copy(cloudTexture, nil, nil)

		for _, balloon := range balloons {
			balloon.Update(elapsedTime)
		}
		//Z轴排序
		sort.Slice(balloons, func(i, j int) bool {
			return balloons[i].pos.Z > balloons[j].pos.Z
		})

		for _, balloon := range balloons {
			balloon.draw(renderer)
		}

		renderer.Present()
		elapsedTime = float32(time.Since(frameStart).Seconds() * 1000)
		fmt.Println("每一帧消耗时间: ", elapsedTime)
		if elapsedTime < 5 {
			sdl.Delay(5 - uint32(elapsedTime))
			elapsedTime = float32(time.Since(frameStart).Seconds() * 1000)
		}
	}
}
