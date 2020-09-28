package imdom

import (
	"fmt"
	"log"
	"net/url"

	"github.com/inkyblackness/imgui-go/v2"
)

// Run implements the main program loop of the application. It returns when the platform signals to stop.
func Run(p Platform, r Renderer) {
	imgui.CurrentIO().SetClipboard(clipboard{platform: p})

	// Initialize local state
	showDemoWindow := false
	clearColor := [3]float32{0.0, 0.0, 0.0}
	connectionStatus := "Online"

	// Initialize TickerPlant
	url := url.URL{Scheme: "ws", Host: "localhost:8080", Path: "/ws"}
	tickerPlant, err := NewTickerPlant(url)
	if err != nil {
		log.Fatalln(err)
		return
	}
	go func() {
		tickerPlant.Run()
		defer func() {
			tickerPlant.Close()
			connectionStatus = "Offline"
		}()
	}()

	// Run the window
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

			if imgui.Button("Subscribe trades") {
				tickerPlant.SubscribeOrderBook("ESU0", "CME")
			}

			imgui.Text(fmt.Sprintf("Application average %.3f ms/frame (%.1f FPS)",
				millisPerSecond/imgui.CurrentIO().Framerate(), imgui.CurrentIO().Framerate()))
			imgui.Text(fmt.Sprintf("Connection Status:"))
			imgui.SameLine()
			imgui.Text(fmt.Sprint(connectionStatus))

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
