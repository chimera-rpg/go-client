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
		Style: UI.Style{
			X: UI.Number{
				Percentage: true,
				Value:      10,
			},
			Y: UI.Number{
				Percentage: true,
				Value:      10,
			},
			W: UI.Number{
				Percentage: true,
				Value:      80,
			},
			H: UI.Number{
				Percentage: true,
				Value:      80,
			},
		},
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
		Style: UI.Style{
			ForegroundColor: UI.Color{255, 255, 255, 255, true},
			BackgroundColor: UI.Color{255, 255, 255, 64, true},
			PaddingLeft: UI.Number{
				Percentage: true,
				Value:      5,
			},
			PaddingRight: UI.Number{
				Percentage: true,
				Value:      5,
			},
			PaddingTop: UI.Number{
				Percentage: true,
				Value:      5,
			},
			PaddingBottom: UI.Number{
				Percentage: true,
				Value:      5,
			},
			Origin: UI.ORIGIN_CENTERX | UI.ORIGIN_CENTERY,
			X: UI.Number{
				Value:      50,
				Percentage: true,
			},
			Y: UI.Number{
				Value:      10,
				Percentage: true,
			},
		},
		Value: "Please choose a server:",
	})

	el_img := UI.NewImageElement(UI.ImageElementConfig{
		Style: UI.Style{
			X: UI.Number{
				Value:      50,
				Percentage: true,
			},
			Y: UI.Number{
				Value:      50,
				Percentage: true,
			},
			W: UI.Number{
				Value: 48,
			},
			H: UI.Number{
				Value: 48,
			},
			Origin: UI.ORIGIN_CENTERX | UI.ORIGIN_CENTERY,
		},
		Image: s.Client.GetPNGData("ui/loading.png"),
	})

	s.ServersWindow.AdoptChild(el)
	el.AdoptChild(el_img)

	el_test := UI.NewTextElement(UI.TextElementConfig{
		Style: UI.Style{
			X: UI.Number{
				Value:      50,
				Percentage: true,
			},
			Y: UI.Number{
				Value:      50,
				Percentage: true,
			},
			Origin: UI.ORIGIN_CENTERX | UI.ORIGIN_CENTERY,
		},
		Value: "Test",
	})
	el_img.AdoptChild(el_test)

	s.Client.Print("Please choose a server: ")
	return
}

func (s *List) Close() {
	s.ServersWindow.Destroy()
}
