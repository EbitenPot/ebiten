package cocoa

// #cgo CFLAGS: -x objective-c -fobjc-arc
// #cgo LDFLAGS: -framework Cocoa -framework OpenGL -framework QuartzCore
//
// #include <stdlib.h>
// #include "input.h"
//
// void Run(size_t width, size_t height, size_t scale, const char* title);
//
import "C"
import (
	"github.com/hajimehoshi/go.ebiten"
	"github.com/hajimehoshi/go.ebiten/graphics/opengl"
	"time"
	"unsafe"
)

type UI struct {
	game           ebiten.Game
	screenWidth    int
	screenHeight   int
	screenScale    int
	graphicsDevice *opengl.Device
	inited         chan bool
	updating       chan bool
	updated        chan bool
	input          chan ebiten.InputState
}

var currentUI *UI

//export ebiten_EbitenOpenGLView_Initialized
func ebiten_EbitenOpenGLView_Initialized() {
	if currentUI.graphicsDevice != nil {
		panic("The graphics device is already initialized")
	}

	currentUI.graphicsDevice = opengl.NewDevice(
		currentUI.screenWidth,
		currentUI.screenHeight,
		currentUI.screenScale)
	currentUI.graphicsDevice.Init()

	currentUI.game.Init(currentUI.graphicsDevice.TextureFactory())

	currentUI.inited <- true
}

//export ebiten_EbitenOpenGLView_Updating
func ebiten_EbitenOpenGLView_Updating() {
	<-currentUI.updating
	currentUI.graphicsDevice.Update(currentUI.game.Draw)
	currentUI.updated <- true
}

//export ebiten_EbitenOpenGLView_InputUpdated
func ebiten_EbitenOpenGLView_InputUpdated(inputType C.int, cx, cy C.int) {
	if inputType == C.InputTypeMouseUp {
		currentUI.input <- ebiten.InputState{-1, -1}
		return
	}

	x, y := int(cx), int(cy)
	x /= currentUI.screenScale
	y /= currentUI.screenScale
	if x < 0 {
		x = 0
	} else if currentUI.screenWidth <= x {
		x = currentUI.screenWidth - 1
	}
	if y < 0 {
		y = 0
	} else if currentUI.screenHeight <= y {
		y = currentUI.screenHeight - 1
	}
	currentUI.input <- ebiten.InputState{x, y}
}

func Run(game ebiten.Game, screenWidth, screenHeight, screenScale int,
	title string) {
	cTitle := C.CString(title)
	defer C.free(unsafe.Pointer(cTitle))

	currentUI = &UI{
		game:         game,
		screenWidth:  screenWidth,
		screenHeight: screenHeight,
		screenScale:  screenScale,
		inited:       make(chan bool),
		updating:     make(chan bool),
		updated:      make(chan bool),
		input:        make(chan ebiten.InputState),
	}

	go func() {
		frameTime := time.Duration(
			int64(time.Second) / int64(ebiten.FPS))
		tick := time.Tick(frameTime)
		gameContext := &GameContext{
			screenWidth:  screenWidth,
			screenHeight: screenHeight,
			inputState:   ebiten.InputState{-1, -1},
		}
		<-currentUI.inited
		for {
			select {
			case gameContext.inputState = <-currentUI.input:
			case <-tick:
				game.Update(gameContext)
			case currentUI.updating <- true:
				<-currentUI.updated
			}
		}
	}()

	C.Run(C.size_t(screenWidth),
		C.size_t(screenHeight),
		C.size_t(screenScale),
		cTitle)
}

type GameContext struct {
	screenWidth  int
	screenHeight int
	inputState   ebiten.InputState
}

func (context *GameContext) ScreenWidth() int {
	return context.screenWidth
}

func (context *GameContext) ScreenHeight() int {
	return context.screenHeight
}

func (context *GameContext) InputState() ebiten.InputState {
	return context.inputState
}
