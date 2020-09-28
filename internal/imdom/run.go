package imdom

import (
	"fmt"
	"time"

	"github.com/inkyblackness/imgui-go/v2"
)

// Platform covers mouse/keyboard/gamepad inputs, cursor shape, timing, windowing.
type Platform interface {
	// ShouldStop is regularly called as the abort condition for the program loop.
	ShouldStop() bool
	// ProcessEvents is called once per render loop to dispatch any pending events.
	ProcessEvents()
	// DisplaySize returns the dimension of the display.
	DisplaySize() [2]float32
	// FramebufferSize returns the dimension of the framebuffer.
	FramebufferSize() [2]float32
	// NewFrame marks the begin of a render pass. It must update the imgui IO state according to user input.
	NewFrame()
	// PostRender marks the completion of one render pass. Typically this causes the display buffer to be swapped.
	PostRender()
	// ClipboardText returns the current text of the clipboard, if available.
	ClipboardText() (string, error)
	// SetClipboardText sets the text as the current text of the clipboard.
	SetClipboardText(text string)
}

type clipboard struct {
	platform Platform
}

// Text retrieves the current text of the clipboard, if available.
func (board clipboard) Text() (string, error) {
	return board.platform.ClipboardText()
}

// SetText sets the text as the current text of the clipboard.
func (board clipboard) SetText(text string) {
	board.platform.SetClipboardText(text)
}

// Renderer covers rending imgui draw data.
type Renderer interface {
	// PreRender causes the display buffer to prepare for new output.
	PreRender(clearColor [3]float32)
	// Render draws the provided imgui draw data.
	Render(displaySize [2]float32, framebufferSize [2]float32, drawData imgui.DrawData)
}

const (
	millisPerSecond = 1000
	sleepDuration   = time.Millisecond * 25
)

// Run implements the main program loop of the application. It returns when the platform signals to stop.
func Run(p Platform, r Renderer) {
	imgui.CurrentIO().SetClipboard(clipboard{platform: p})

	showDemoWindow := false
	clearColor := [3]float32{0.0, 0.0, 0.0}

	for !p.ShouldStop() {
		p.ProcessEvents()

		// Signal start of a new frame
		p.NewFrame()
		imgui.NewFrame()

		// Show a simple window.
		{
			imgui.Begin("Welcome")
			imgui.Text("Hello!")                         // Display some text
			imgui.ColorEdit3("clear color", &clearColor) // Edit 3 floats representing a color

			imgui.Checkbox("Demo Window", &showDemoWindow) // Edit bools storing our window open/close state

			imgui.Text(fmt.Sprintf("Application average %.3f ms/frame (%.1f FPS)",
				millisPerSecond/imgui.CurrentIO().Framerate(), imgui.CurrentIO().Framerate()))
			imgui.End()
		}

		// Show the ImGui demo window. Most of the sample code is in imgui.ShowDemoWindow().
		if showDemoWindow {
			// Normally user code doesn't need/want to call this because positions are saved in .ini file anyway.
			// Here we just want to make the demo initial state a bit more friendly!
			const demoX = 650
			const demoY = 20
			imgui.SetNextWindowPosV(imgui.Vec2{X: demoX, Y: demoY}, imgui.ConditionFirstUseEver, imgui.Vec2{})

			imgui.ShowDemoWindow(&showDemoWindow)
		}

		// Rendering
		imgui.Render() // This call only creates the draw data list. Actualy rendering to framebuffer is done below.

		r.PreRender(clearColor)
		r.Render(p.DisplaySize(), p.FramebufferSize(), imgui.RenderedDrawData())
		p.PostRender()

		// sleep to avoid 100% CPU usage
		//<-time.After(sleepDuration)
	}
}
