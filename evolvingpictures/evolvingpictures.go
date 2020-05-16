package main

//1.我们不再自己手动绘制texture，直接使用sdl2 自带的texture
//2.并使用GPU渲染
//3.使用仅有的三个素材，渲染出多个气球，并使用package vector3下的向量代替原有的pos
//4.气球移动
//5.鼠标输入
//6.点击气球，气球爆炸，发出声音,爆炸效果
//7.气球爆炸，删除气球
//TODO 8.气球之间的碰撞检测
import (
	"gogame/evolvingpictures/apt"
	"log"
	"time"

	"github.com/veandco/go-sdl2/sdl"
)

const (
	H int = 600
	W int = 800
	D int = 100 //三维坐标中的深度
)

type mouseState struct {
	leftButton  bool
	rightButton bool
	x, y        int
}

func getMouseState() mouseState {
	mouseX, mouseY, mouseButtonState := sdl.GetMouseState()
	leftButton := mouseButtonState & sdl.ButtonLMask()
	rightButton := mouseButtonState & sdl.ButtonRMask()
	return mouseState{!(leftButton == 0), !(rightButton == 0), int(mouseX), int(mouseY)}
}

type audioState struct {
	explodeBytes []byte
	deviceID     sdl.AudioDeviceID
	audioSpec    *sdl.AudioSpec
}

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

func APT2Texture(node1, node2 apt.Node, w, h int, renderer *sdl.Renderer) *sdl.Texture {
	// -1.0 and 1.0
	scale := float32(255 / 2)
	offset := -1.0 * scale
	pixels := make([]byte, w*h*4)
	pixelsIndex := 0
	for yi := 0; yi < h; yi++ {
		y := float32(yi)/float32(h)*2 - 1
		for xi := 0; xi < w; xi++ {
			x := float32(xi)/float32(w)*2 - 1
			c := node1.Eval(x, y)
			c2 := node2.Eval(x, y)
			pixels[pixelsIndex] = byte(c*scale - offset)
			pixelsIndex++
			pixels[pixelsIndex] = byte(c2*scale - offset)
			pixelsIndex++
			pixels[pixelsIndex] = 0 //byte(c*scale-offset)
			pixelsIndex++
			pixelsIndex++ //skip alpha
		}
	}
	return pixelsToTexture(renderer, pixels, w, h)
}

func main() {
	err := sdl.Init(sdl.INIT_EVERYTHING)
	if err != nil {
		panic(err)
	}
	wind, e := sdl.CreateWindow("evolving_pictures", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		int32(W), int32(H), sdl.WINDOW_SHOWN)
	if e != nil {
		panic(e)
	}
	defer wind.Destroy()
	renderer, err := sdl.CreateRenderer(wind, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		log.Fatal(err)
	}
	defer renderer.Destroy()
	sdl.SetHint(sdl.HINT_RENDER_SCALE_QUALITY, "1")
	////--声音
	//explodeBytes, audioSpec := sdl.LoadWAV("explode.wav")
	//err = sdl.OpenAudio(audioSpec, nil)
	//if err != nil {
	//	panic(err)
	//}
	//defer sdl.FreeWAV(explodeBytes)
	//audioState := audioState{explodeBytes, 1, audioSpec}
	////--

	var elapsedTime float32
	//currentMouseState := getMouseState()
	//preMouseState := currentMouseState

	x := apt.OpX{}
	y := apt.OpY{}
	sine := apt.OpSin{}
	plus := apt.OpPlus{}
	sine.Child = &x
	plus.LeftChild = &sine
	plus.RightNode = &y

	texture := APT2Texture(&plus, &sine, W, H, renderer)

	for {
		//currentMouseState = getMouseState()
		frameStart := time.Now()
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event.(type) {
			case *sdl.QuitEvent:
				return
			}
		}

		_ = renderer.Copy(texture, nil, nil)

		renderer.Present()
		elapsedTime = float32(time.Since(frameStart).Seconds() * 1000)
		//fmt.Println("每一帧消耗时间: ", elapsedTime)
		if elapsedTime < 5 {
			sdl.Delay(5 - uint32(elapsedTime))
			elapsedTime = float32(time.Since(frameStart).Seconds() * 1000)
		}
		//preMouseState = currentMouseState
	}
}
