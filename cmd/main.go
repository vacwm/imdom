package main

import (
	"fmt"
	"os"

	"github.com/inkyblackness/imgui-go/v2"
	"github.com/vacwm/imdom/internal/imdom"
	"github.com/vacwm/imdom/internal/platforms"
	"github.com/vacwm/imdom/internal/renderers"
)

func main() {
	context := imgui.CreateContext(nil)
	defer context.Destroy()
	io := imgui.CurrentIO()

	platform, err := platforms.NewGLFW(io)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(-1)
	}
	defer platform.Dispose()

	renderer, err := renderers.NewOpenGL3(io)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(-1)
	}
	defer renderer.Dispose()

	imdom.Run(platform, renderer)
}
