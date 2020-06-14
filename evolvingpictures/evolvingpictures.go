package main


import (
	"gogame/evolvingpictures/apt"
	"log"
	"math/rand"
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

func APT2Texture(redNode, greenNode, blueNode apt.Node, w, h int, renderer *sdl.Renderer) *sdl.Texture {
	// -1.0 and 1.0
	scale := float32(255 / 2)
	offset := -1.0 * scale
	pixels := make([]byte, w*h*4)
	pixelsIndex := 0
	for yi := 0; yi < h; yi++ {
		y := float32(yi)/float32(h)*2 - 1
		for xi := 0; xi < w; xi++ {
			x := float32(xi)/float32(w)*2 - 1
			r := redNode.Eval(x, y)
			g := greenNode.Eval(x, y)
			b := blueNode.Eval(x, y)
			pixels[pixelsIndex] = byte(r*scale - offset)
			pixelsIndex++
			pixels[pixelsIndex] = byte(g*scale - offset)
			pixelsIndex++
			pixels[pixelsIndex] = byte(b*scale - offset)
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

	//随机的抽象树
	rand.Seed(time.Now().UTC().UnixNano())
	aptR := apt.GetRandomNode()
	aptG := apt.GetRandomNode()
	aptB := apt.GetRandomNode()

	num := rand.Intn(20)
	for i := 0; i < num; i++{
	    aptR.AddRandom(apt.GetRandomNode())
	}
	num = rand.Intn(20)
	for i := 0; i < num; i++{
		aptG.AddRandom(apt.GetRandomNode())
	}
	num = rand.Intn(20)
	for i := 0; i < num; i++{
		aptB.AddRandom(apt.GetRandomNode())
	}
	for{
		_, nilCount := aptR.NodeCount()
		if nilCount == 0 {
			break
		}
		aptR.AddRandom(apt.GetRandomLeaf())
	}
	for{
		_, nilCount := aptG.NodeCount()
		if nilCount == 0 {
			break
		}
		aptG.AddRandom(apt.GetRandomLeaf())
	}
	for{
		_, nilCount := aptB.NodeCount()
		if nilCount == 0 {
			break
		}
		aptB.AddRandom(apt.GetRandomLeaf())
	}


	texture := APT2Texture(aptR,aptG,aptB, 640, 480, renderer)

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
