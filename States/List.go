package States

import (
	"github.com/chimera-rpg/go-client/Client"
	"github.com/chimera-rpg/go-client/UI"
)

type List struct {
	Client.State
	ServersWindow UI.Window
}

func (s *List) Init(t interface{}) (state Client.StateI, nextArgs interface{}, err error) {
	err = s.ServersWindow.Setup(UI.WindowConfig{
		Value: "Server List",
		Style: `
			X 10%
			Y 10%
			W 80%
			H 80%
		`,
		Parent: s.Client.RootWindow,
		RenderFunc: func(w *UI.Window) {
			w.Context.Renderer.SetDrawColor(32, 32, 33, 128)
			w.Context.Renderer.Clear()
		},
	})

	/*
	  Imagine a future where the following was simplified to:

	  UI.TextElementConfig{
	    ForegroundColor: "255 255 255 255",
	    BackgroundColor: "255 255 255 64",
	    padding: "5% 5% 5% 5%",
	    origin: "centerx centery",
	    X: "50%",
	    Y: "10%",
	    Value: "Please choose a server",
	  }
	*/
	el := UI.NewTextElement(UI.TextElementConfig{
		Style: `
			ForegroundColor 255 255 255 255
			BackgroundColor 255 255 255 64
			PaddingLeft 5%
			PaddingRight 5%
			PaddingTop 5%
			PaddingBottom 5%
			Origin CenterX CenterY
			X 50%
			Y 10%
		`,
		Value: "Please choose a server:",
	})

	el_img := UI.NewImageElement(UI.ImageElementConfig{
		Style: `
			X 50%
			Y 50%
			W 48
			H 48
			Origin CenterX CenterY
		`,
		Image: s.Client.GetPNGData("ui/loading.png"),
	})

	s.ServersWindow.AdoptChild(el)
	el.AdoptChild(el_img)

	el_test := UI.NewTextElement(UI.TextElementConfig{
		Style: `
			X 50%
			Y 50%
			Origin CenterX CenterY
		`,
		Value: "Test",
	})
	el_img.AdoptChild(el_test)

	s.Client.Print("Please choose a server: ")
	return
}

func (s *List) Close() {
	s.ServersWindow.Destroy()
}
