package main

/*
	相比于第一个版本，画面流畅
	加入比分，谁先得到3分比赛结束
	加入游戏状态。游戏按pause开始，每得一分暂停一次，得到3分从头开始
	碰撞优化
*/

import (
	"log"
	"time"

	"github.com/veandco/go-sdl2/sdl"
)

// --
type gameState int

const (
	start gameState = iota
	play
)

var state = start

// --

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
	if ball.y-ball.radius < 0 || ball.y+ball.radius > 600 {
		ball.yVelocity = -ball.yVelocity
	}
	if ball.x < 0 {
		rightpaddle.score++
		ball.x = 400
		ball.y = 300
		state = start
	} else if int(ball.x) > 800 {
		leftpaddle.score++
		ball.x = 400
		ball.y = 300
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
func lerp(a, b float32, pct float32) float32 {
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
	numX := lerp(paddle.x, 400, 0.2)
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
func clear(pixels []byte) {
	for i := range pixels {
		pixels[i] = 0
	}
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

	player1 := paddle{pos{100, 100}, 20, 100, 300, 0, color{255, 255, 255}}
	player2 := paddle{pos{700, 100}, 20, 100, 300, 0, color{255, 255, 255}}
	ball := ball{pos{300, 300}, 20, 400, 400, color{255, 255, 255}}
	keystate := sdl.GetKeyboardState()

	var frameStart time.Time
	var elapsedTime float32

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
			player2.aiUpdate(&ball, elapsedTime)
			ball.update(&player1, &player2, elapsedTime)
		} else if state == start {
			if keystate[sdl.SCANCODE_SPACE] != 0 {
				if player1.score == 3 || player2.score == 3 {
					player1.score = 0
					player2.score = 0
				}
				state = play
			}
		}
		clear(pixels)
		player1.draw(pixels)
		player2.draw(pixels)
		ball.draw(pixels)

		_ = tex.Update(nil, pixels, 800*4)
		_ = renderer.Copy(tex, nil, nil)
		renderer.Present()

		elapsedTime = float32(time.Since(frameStart).Seconds())
		// fmt.Println(elapsedTime)
		if elapsedTime < .005 {
			sdl.Delay(5 - uint32(elapsedTime/1000.0))
			elapsedTime = float32(time.Since(frameStart).Seconds())
		}
	}
}
