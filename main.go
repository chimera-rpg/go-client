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
		{Flags: sdl.MESSAGEBOX_BUTTON_RETURNKEY_DEFAULT, ButtonID: 1, Text: "OH NO"},
	}

	messageboxdata := sdl.MessageBoxData{
		Flags:       flags,
		Window:      nil,
		Title:       "Chimera",
		Message:     fmt.Sprintf(format, a...),
		Buttons:     buttons,
		ColorScheme: nil,
	}

	sdl.ShowMessageBox(&messageboxdata)
}

func showError(format string, a ...interface{}) {
	showWindow(sdl.MESSAGEBOX_ERROR, format, a...)
}
func showWarning(format string, a ...interface{}) {
	showWindow(sdl.MESSAGEBOX_WARNING, format, a...)
}
func showInfo(format string, a ...interface{}) {
	showWindow(sdl.MESSAGEBOX_INFORMATION, format, a...)
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			showError("%v", r.(error).Error())
			debug.PrintStack()
		}
	}()
	log.Print("Starting Chimera client (golang)")

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
		clientInstance.StateChannel <- client.StateMessage{State: &states.Handshake{}, Args: *netPtr}
	} else {
		clientInstance.StateChannel <- client.StateMessage{State: &states.List{}, Args: nil}
	}

	// Start our UI Loop.
	uiInstance.Loop()

	log.Print("Sayonara!")
}
