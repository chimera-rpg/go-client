package main

import (
	"flag"
	"github.com/sirupsen/logrus"
	"runtime/debug"

	"github.com/chimera-rpg/go-client/client"
	"github.com/chimera-rpg/go-client/data"
	"github.com/chimera-rpg/go-client/states"
	"github.com/chimera-rpg/go-client/ui"
)

func main() {
	var log = logrus.New()
	var err error
	var dataManager data.Manager
	var clientInstance client.Client
	var uiInstance ui.Instance

	defer func() {
		if r := recover(); r != nil {
			ui.ShowError("%v", r.(error).Error())
			debug.PrintStack()
		}
	}()
	log.SetLevel(logrus.DebugLevel)
	log.Print("Starting Chimera client (golang)")

	if err = dataManager.Setup(log); err != nil {
		ui.ShowError("%s", err)
	}

	// Setup our UI
	if err = uiInstance.Setup(&dataManager); err != nil {
		ui.ShowError("%s", err)
		return
	}
	defer uiInstance.Cleanup()

	ui.GlobalInstance = &uiInstance

	// Setup our Client
	if err = clientInstance.Setup(&dataManager, &uiInstance, log); err != nil {
		ui.ShowError("%s", err)
		return
	}
	defer clientInstance.Destroy()
	// Start the clientInstance's channel listening loop as a coroutine
	go clientInstance.ChannelLoop()

	flag.String("username", "", "username")
	flag.String("password", "", "password")
	flag.String("character", "", "name of character")
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
