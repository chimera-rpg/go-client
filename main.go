package main

import (
	"flag"
	"fmt"
	"log"
	"runtime/debug"

	"github.com/chimera-rpg/go-client/client"
	"github.com/chimera-rpg/go-client/states"
	"github.com/chimera-rpg/go-client/ui"
	"github.com/veandco/go-sdl2/sdl"
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
	log.Print("Starting Chimera clientInstance (golang)")

	clientInstance, err := client.NewClient()
	defer clientInstance.Destroy()
	if err != nil {
		showError("%s", err)
		return
	}

	uiInstance, err := ui.NewInstance()
	defer uiInstance.Cleanup()
	if err != nil {
		showError("%s", err)
		return
	}
	ui.GlobalInstance = uiInstance

	// Setup our UI
	uiInstance.Setup(clientInstance.DataRoot)

	// Setup our Client
	if err = clientInstance.Setup(uiInstance); err != nil {
		showError("%s", err)
		return
	}
	// Start the clientInstance's channel listening loop as a coroutine
	go clientInstance.ChannelLoop()

	netPtr := flag.String("connect", "", "SERVER:PORT")
	flag.Parse()
	// Automatically attempt to connect if the server flag was passed
	if len(*netPtr) > 0 {
		clientInstance.StateChannel <- client.StateMessage{&states.Handshake{}, *netPtr}
	} else {
		clientInstance.StateChannel <- client.StateMessage{&states.List{}, nil}
	}

	// Start our UI Loop.
	uiInstance.Loop()

	log.Print("Sayonara!")
}
