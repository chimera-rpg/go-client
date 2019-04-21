package main

import (
	"flag"
	"fmt"
	"github.com/chimera-rpg/go-client/Client"
	"github.com/chimera-rpg/go-client/States"
	"github.com/chimera-rpg/go-client/UI"
	"github.com/veandco/go-sdl2/sdl"
	"log"
	"runtime/debug"
)

func showWindow(flags uint32, format string, a ...interface{}) {
	buttons := []sdl.MessageBoxButtonData{
		{sdl.MESSAGEBOX_BUTTON_RETURNKEY_DEFAULT, 1, "OH NO"},
	}

	messageboxdata := sdl.MessageBoxData{
		flags,
		nil,
		"Chimera",
		fmt.Sprintf(format, a...),
		buttons,
		nil,
	}

	sdl.ShowMessageBox(&messageboxdata)
}

func showError(format string, a ...interface{}) {
	showWindow(sdl.MESSAGEBOX_ERROR, format, a)
}
func showWarning(format string, a ...interface{}) {
	showWindow(sdl.MESSAGEBOX_WARNING, format, a)
}
func showInfo(format string, a ...interface{}) {
	showWindow(sdl.MESSAGEBOX_INFORMATION, format, a)
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			showError("%v", r.(error).Error())
			debug.PrintStack()
		}
	}()
	log.Print("Starting Chimera client (golang)")

	client, err := Client.NewClient()
	defer client.Destroy()
	if err != nil {
		showError("%s", err)
		return
	}

	ui, err := UI.NewInstance()
	defer ui.Cleanup()
	if err != nil {
		showError("%s", err)
		return
	}
	UI.GlobalInstance = ui

	// Setup our UI
	ui.Setup(client.DataRoot)

	// Setup our Client
	if err = client.Setup(ui); err != nil {
		showError("%s", err)
		return
	}

	// Start the client's channel listening loop as a coroutine
	go client.ChannelLoop()

	netPtr := flag.String("connect", "", "SERVER:PORT")
	flag.Parse()

	// Automatically attempt to connect if the server flag was passed
	if len(*netPtr) > 0 {
		client.StateChannel <- Client.StateMessage{&States.Handshake{}, *netPtr}
	} else {
		client.StateChannel <- Client.StateMessage{&States.List{}, nil}
	}

	// Start our UI Loop.
	ui.Loop()

	log.Print("Sayonara!")
}
