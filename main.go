package main

import (
	"flag"
	"log"
	"runtime/debug"

	"github.com/chimera-rpg/go-client/client"
	"github.com/chimera-rpg/go-client/states"
	"github.com/chimera-rpg/go-client/ui"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			ui.ShowError("%v", r.(error).Error())
			debug.PrintStack()
		}
	}()
	log.Print("Starting Chimera client (golang)")

	clientInstance, err := client.NewClient()
	if err != nil {
		ui.ShowError("%s", err)
		return
	}
	defer clientInstance.Destroy()

	uiInstance, err := ui.NewInstance()
	if err != nil {
		ui.ShowError("%s", err)
		return
	}
	defer uiInstance.Cleanup()
	ui.GlobalInstance = uiInstance

	// Setup our UI
	uiInstance.Setup(clientInstance.DataRoot)

	// Setup our Client
	if err = clientInstance.Setup(uiInstance); err != nil {
		ui.ShowError("%s", err)
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
