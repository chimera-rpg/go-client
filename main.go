package main

import (
	"runtime"
	"runtime/debug"

	"github.com/sirupsen/logrus"

	"net/http"
	_ "net/http/pprof"

	"github.com/chimera-rpg/go-client/audio"
	"github.com/chimera-rpg/go-client/client"
	"github.com/chimera-rpg/go-client/data"
	"github.com/chimera-rpg/go-client/states/list"
	"github.com/chimera-rpg/go-client/ui"
)

func main() {
	var log = logrus.New()
	log.SetLevel(logrus.DebugLevel)
	log.SetFormatter(&logrus.TextFormatter{ForceColors: true}) // It would be ideal to only force colors on Windows 10+ -- checking this is possible with x/sys/windows/registry, though we'd need OS-specific source files for log initialization.
	var err error
	var dataManager data.Manager
	var clientInstance client.Client
	var uiInstance ui.Instance
	var audioInstance audio.Instance

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

	// Setup our Audio
	if err = audioInstance.Setup(log); err != nil {
		ui.ShowError("%s", err)
	} else {
		go audioInstance.Loop()
		defer audioInstance.Quit()
		audio.GlobalInstance = &audioInstance
		// FIXME: This isn't the right place for this.
		for k, v := range dataManager.Sounds() {
			audioInstance.CommandChannel <- audio.CommandNewSound{
				ID:       k,
				Type:     v.Type,
				Filepath: v.Filepath,
			}
		}
	}

	// Preload our graphics. FIXME: This isn't the right place for this either.
	for _, v := range dataManager.Images() {
		uiInstance.ImageLoadChan <- ui.UpdateImageID(v.ID)
	}

	// Setup our Client
	if err = clientInstance.Setup(&dataManager, &uiInstance, &audioInstance, log); err != nil {
		ui.ShowError("%s", err)
		return
	}
	defer clientInstance.Destroy()
	// Start the clientInstance's channel listening loop as a coroutine
	go clientInstance.ChannelLoop()

	if clientInstance.Flags.Profile {
		log.Print("Starting profiling on port 6060")
		go func() {
			runtime.SetBlockProfileRate(1)
			log.Println(http.ListenAndServe("localhost:6060", nil))
		}()
	}

	// Automatically attempt to connect if the server flag was passed
	if len(clientInstance.Flags.Connect) > 0 {
		clientInstance.StateChannel <- client.StateMessage{State: &list.Handshake{}, Args: clientInstance.Flags.Connect}
	} else {
		clientInstance.StateChannel <- client.StateMessage{State: &list.List{}, Args: nil}
	}

	// Start our UI Loop.
	uiInstance.Loop()

	log.Print("Sayonara!")
}
