package main

import (
	"log"

	"github.com/veandco/go-sdl2/sdl"
)

type color struct {
	r, g, b byte
}
type pos struct {
	x, y float32
}

type ball struct {
	pos
	radius    int
	xVelocity float32
	yVelocity float32
	color     color
}

func (ball *ball) draw(pixels []byte) {
	for y := -ball.radius; y < ball.radius; y++ {
		for x := -ball.radius; x < ball.radius; x++ {
			if x*x+y*y < ball.radius*ball.radius {
				setPixels(int(ball.x)+x, int(ball.y)+y, color{255, 255, 255}, pixels)
			}
		}
	}
}
func (ball *ball) update(leftpaddle *paddle, rightpaddle *paddle) {
	ball.x += ball.xVelocity
	ball.y += ball.yVelocity
	//TODO handle collisions
	if int(ball.y)-ball.radius < 0 || int(ball.y)+ball.radius > 600 {
		ball.yVelocity = -ball.yVelocity
	}
	if ball.x < 0 || int(ball.x) > 800 {
		ball.x = 300
		ball.y = 300
	}
	if int(ball.x)-ball.radius < int(leftpaddle.x)+leftpaddle.w/2 {
		if int(ball.y)-ball.radius > int(leftpaddle.y)-leftpaddle.h/2 &&
			int(ball.y)+ball.radius < int(leftpaddle.y)+leftpaddle.h/2 {
			ball.xVelocity = -ball.xVelocity
		}
	}
	if int(ball.x)+ball.radius > int(rightpaddle.x)-rightpaddle.w/2 {
		if int(ball.y)-ball.radius > int(rightpaddle.y)-rightpaddle.h/2 &&
			int(ball.y)+ball.radius < int(rightpaddle.y)+rightpaddle.h/2 {
			ball.xVelocity = -ball.xVelocity
		}
	}

}

type paddle struct {
	pos
	w     int
	h     int
	color color
}

func (paddle *paddle) draw(pixels []byte) {
	startX := int(paddle.x) - paddle.w/2
	startY := int(paddle.y) - paddle.h/2
	for y := 0; y < paddle.h; y++ {
		for x := 0; x < paddle.w; x++ {
			setPixels(startX+x, startY+y, color{255, 255, 255}, pixels)
		}
	}
}

func (paddle *paddle) update(keystate []uint8) {
	if keystate[sdl.SCANCODE_UP] != 0 {
		paddle.y -= 5
	}
	if keystate[sdl.SCANCODE_DOWN] != 0 {
		paddle.y += 5
	}

}
func (paddle *paddle) aiUpdate(ball *ball) {
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

	player1 := paddle{pos{100, 100}, 20, 100, color{255, 255, 255}}
	player2 := paddle{pos{700, 100}, 20, 100, color{255, 255, 255}}
	ball := ball{pos{300, 300}, 20, 3, 3, color{255, 255, 255}}
	keystate := sdl.GetKeyboardState()
	for {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event.(type) {
			case *sdl.QuitEvent:
				return
			}
		}
		clear(pixels)
		player1.update(keystate)
		player2.aiUpdate(&ball)
		ball.update(&player1, &player2)
		player1.draw(pixels)
		player2.draw(pixels)
		ball.draw(pixels)

		tex.Update(nil, pixels, 800*4)
		renderer.Copy(tex, nil, nil)
		renderer.Present()
		sdl.Delay(20)
	}
}
