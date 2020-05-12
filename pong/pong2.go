package main

/*
	1.
	相比于第一个版本，画面流畅
	加入比分，谁先得到3分比赛结束
	加入游戏状态。游戏按pause开始，每得一分暂停一次，得到3分从头开始
	碰撞优化
	2.
	添加一个渐变背景
	3.
	用图片里的balloon代替 白色像素球
*/

import (
	"fmt"
	"gogame/noise"
	"image/png"
	"log"
	"os"
	"time"

	"github.com/veandco/go-sdl2/sdl"
)

// --
type gameState int

const (
	W     int       = 800
	H     int       = 600
	start gameState = iota
	play
)

var state = start

// --
func lerp(b1, b2 byte, pct float32) byte {
	return byte(float32(b1) + pct*(float32(b2)-float32(b1)))
}

func colorLerp(c1, c2 color, pct float32) color {
	return color{lerp(c1.r, c2.r, pct), lerp(c1.g, c2.g, pct), lerp(c1.b, c2.b, pct)}
}

func getGradient(c1, c2 color) []color {
	res := make([]color, 256)
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

func rescale(noise []float32, min, max float32, gradient []color, w, h int) []byte {
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

type color struct {
	r, g, b byte
}
type pos struct {
	x, y float32
}

// 0,1,2,3
var nums = [][]byte{
	{1, 1, 1,
		1, 0, 1,
		1, 0, 1,
		1, 0, 1,
		1, 1, 1,
	},
	{1, 1, 0,
		0, 1, 0,
		0, 1, 0,
		0, 1, 0,
		1, 1, 1,
	},
	{1, 1, 1,
		0, 0, 1,
		1, 1, 1,
		1, 0, 0,
		1, 1, 1,
	},
	{1, 1, 1,
		0, 0, 1,
		1, 1, 1,
		0, 0, 1,
		1, 1, 1,
	},
}

func drawNumber(pos pos, color color, size int, num int, pixels []byte) {
	startX := int(pos.x) - (size*3)/2
	startY := int(pos.y) - (size*5)/2
	for i, v := range nums[num] {
		if v == 1 {
			for y := startY; y < startY+size; y++ {
				for x := startX; x < startX+size; x++ {
					setPixels(x, y, color, pixels)
				}
			}
		}
		startX += size
		if (i+1)%3 == 0 {
			startY += size
			startX -= size * 3
		}
	}

}

type ball struct {
	pos
	radius    float32
	xVelocity float32
	yVelocity float32
	color     color
}

func (ball *ball) draw(pixels []byte) {
	for y := -ball.radius; y < ball.radius; y++ {
		for x := -ball.radius; x < ball.radius; x++ {
			if x*x+y*y < ball.radius*ball.radius {
				setPixels(int(ball.x+x), int(ball.y+y), color{255, 255, 255}, pixels)
			}
		}
	}
}

func (ball *ball) update(leftpaddle *paddle, rightpaddle *paddle, elapsedTime float32) {
	ball.x += ball.xVelocity * elapsedTime
	ball.y += ball.yVelocity * elapsedTime
	//TODO handle collisions
	if ball.y-ball.radius < 0 || ball.y+ball.radius > float32(H) {
		ball.yVelocity = -ball.yVelocity
	}
	if ball.x < 0 {
		rightpaddle.score++
		ball.x = float32(W / 2)
		ball.y = float32(H / 2)
		state = start
	} else if int(ball.x) > W {
		leftpaddle.score++
		ball.x = float32(W / 2)
		ball.y = float32(H / 2)
		state = start
	}
	if ball.x-ball.radius < leftpaddle.x+leftpaddle.w/2 {
		if ball.y-ball.radius > leftpaddle.y-leftpaddle.h/2 &&
			ball.y+ball.radius < leftpaddle.y+leftpaddle.h/2 {
			ball.xVelocity = -ball.xVelocity
			//碰撞优化，ball在碰到paddel后，位置更新到paddle之外
			ball.x = leftpaddle.x + leftpaddle.w/2.0 + ball.radius
		}
	}
	if ball.x+ball.radius > rightpaddle.x-rightpaddle.w/2 {
		if ball.y-ball.radius > rightpaddle.y-rightpaddle.h/2 &&
			ball.y+ball.radius < rightpaddle.y+rightpaddle.h/2 {
			ball.xVelocity = -ball.xVelocity
			//碰撞优化，ball在碰到paddel后，位置更新到paddle之外
			ball.x = rightpaddle.x - rightpaddle.w/2.0 - ball.radius
		}
	}

}

type paddle struct {
	pos
	w     float32
	h     float32
	speed float32
	score int
	color color
}

//lerp 将一个东西显示在[a - b]之间，距离a pct个百分比的位置
func flerp(a, b float32, pct float32) float32 {
	return a + pct*(b-a)
}

func (paddle *paddle) draw(pixels []byte) {
	startX := int(paddle.x - paddle.w/2)
	startY := int(paddle.y - paddle.h/2)
	for y := 0; y < int(paddle.h); y++ {
		for x := 0; x < int(paddle.w); x++ {
			setPixels(startX+x, startY+y, color{255, 255, 255}, pixels)
		}
	}

	//在paddle旁边显示分数
	numX := flerp(paddle.x, 400, 0.2)
	drawNumber(pos{numX, 40}, paddle.color, 10, paddle.score, pixels)
}

func (paddle *paddle) update(keystate []uint8, elapsedTime float32) {
	if keystate[sdl.SCANCODE_UP] != 0 {
		paddle.y -= paddle.speed * elapsedTime
	}
	if keystate[sdl.SCANCODE_DOWN] != 0 {
		paddle.y += paddle.speed * elapsedTime
	}

}

func (paddle *paddle) aiUpdate(ball *ball, elapsedTime float32) {
	paddle.y = ball.y

}

func (paddle *paddle) aiUpdateOfBalloon(balloon *balloon, elapsedTime float32) {
	paddle.y = balloon.y
}
func (tex *balloon) update(leftpaddle *paddle, rightpaddle *paddle, elapsedTime float32) {
	tex.x += tex.xVelocity * elapsedTime
	tex.y += tex.yVelocity * elapsedTime
	//TODO handle collisions
	if tex.y < 0 || tex.y > float32(H) {
		tex.yVelocity = -tex.yVelocity
	}
	if tex.x < 0 {
		rightpaddle.score++
		tex.x = float32(W / 2)
		tex.y = float32(H / 2)
		state = start
	} else if int(tex.x) > W {
		leftpaddle.score++
		tex.x = float32(W / 2)
		tex.y = float32(H / 2)
		state = start
	}
	if tex.x < leftpaddle.x+leftpaddle.w/2 {
		if tex.y > leftpaddle.y-leftpaddle.h/2 &&
			tex.y < leftpaddle.y+leftpaddle.h/2 {
			tex.xVelocity = -tex.xVelocity
			//碰撞优化，ball在碰到paddel后，位置更新到paddle之外
			tex.x = leftpaddle.x + leftpaddle.w/2.0
		}
	}
	if tex.x > rightpaddle.x-rightpaddle.w/2 {
		if tex.y > rightpaddle.y-rightpaddle.h/2 &&
			tex.y < rightpaddle.y+rightpaddle.h/2 {
			tex.xVelocity = -tex.xVelocity
			//碰撞优化，ball在碰到paddel后，位置更新到paddle之外
			tex.x = rightpaddle.x - rightpaddle.w/2.0
		}
	}

}

type texture struct {
	pos
	pixels      []byte
	w, h, pitch int
}

type balloon struct {
	texture
	xVelocity float32
	yVelocity float32
}

func (balloon *balloon) drawAlpha(pixels []byte) {
	for y := 0; y < balloon.h; y++ {
		for x := 0; x < balloon.w; x++ {
			screenY := y + int(balloon.y)
			screenX := x + int(balloon.x)
			if screenX >= 0 && screenX < W && screenY >= 0 && screenY < H {
				texIndex := y*balloon.pitch + x*4
				screenIndex := screenY*W*4 + screenX*4
				srcR := int(balloon.pixels[texIndex])
				srcG := int(balloon.pixels[texIndex+1])
				srcB := int(balloon.pixels[texIndex+2])
				srcA := int(balloon.pixels[texIndex+3])
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

//缩方图片大小(默认缩小3倍)
func (balloon *balloon) scale() *balloon {
	//我们这里先缩小3倍
	resPixels := make([]byte, len(balloon.pixels)/3)

	index := 0
	for i := 0; i < len(balloon.pixels); i += 3 {
		resPixels[index] = balloon.pixels[i]
		index++
	}
	balloon.pixels = resPixels
	balloon.w /= 3
	balloon.h /= 3
	return balloon
}
func loadOneBalloon() balloon {
	file, e := os.Open("../balloons/balloon_green.png")
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
	balloonsTex := texture{pos{float32(W/2) - 60.0, float32(H/2) - 80.0}, balloonsPixels, w, h, w * 4}
	return balloon{balloonsTex, 400, 400}
}
func clear(pixels []byte) {
	for i := range pixels {
		pixels[i] = 0
	}
}

func setPixels(x, y int, c color, pixels []byte) {
	index := (y*W + x) * 4
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

	player1 := paddle{pos{100, 100}, 20, 100, 300, 0, color{255, 255, 255}}
	player2 := paddle{pos{700, 100}, 20, 100, 300, 0, color{255, 255, 255}}
	//ball := ball{pos{300, 300}, 20, 400, 400, color{255, 255, 255}}
	keystate := sdl.GetKeyboardState()

	noises, min, max := noise.MakeNoise(noise.FBM, 0.01, 0.2, 2, 3, W, H)
	gradient := getGradient(color{255, 0, 0}, color{0, 0, 0})
	noisePixels := rescale(noises, min, max, gradient, W, H)

	//加载balloon
	balloon := loadOneBalloon()
	//缩放大小
	balloon.scale()

	var frameStart time.Time
	var elapsedTime float32

	//main
	for {
		frameStart = time.Now()
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event.(type) {
			case *sdl.QuitEvent:
				return
			}
		}
		if state == play {
			player1.update(keystate, elapsedTime)
			player2.aiUpdateOfBalloon(&balloon, elapsedTime)
			balloon.update(&player1, &player2, elapsedTime)
		} else if state == start {
			if keystate[sdl.SCANCODE_SPACE] != 0 {
				if player1.score == 3 || player2.score == 3 {
					player1.score = 0
					player2.score = 0
				}
				state = play
			}
		}

		for i := range noisePixels {
			pixels[i] = noisePixels[i]
		}

		player1.draw(pixels)
		player2.draw(pixels)
		//ball.draw(pixels)
		balloon.drawAlpha(pixels)

		_ = tex.Update(nil, pixels, W*4)
		_ = renderer.Copy(tex, nil, nil)
		renderer.Present()

		elapsedTime = float32(time.Since(frameStart).Seconds() * 1000)
		fmt.Println("每一帧消耗时间: ", elapsedTime)
		if elapsedTime < 5 {
			sdl.Delay(5 - uint32(elapsedTime))
			elapsedTime = float32(time.Since(frameStart).Seconds())
		}
	}
}
