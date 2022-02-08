package audio

import (
	"github.com/sirupsen/logrus"
)

// Instance is the managing instance of the Audio system.
type Instance struct {
	log            *logrus.Logger
	sounds         map[uint32]*Sound
	playingMusic   map[uint32]uint32
	CommandChannel chan CommandI
	QuitChannel    chan bool
}

// GlobalInstance is the reference to the instantiated Instance.
var GlobalInstance *Instance

// Loop is the loop for the audio instance.
func (instance *Instance) Loop() {
	for {
		select {
		case <-instance.QuitChannel:
			instance.log.Println("Quit")
			// TODO: Cleanup.
			return
		case cmd := <-instance.CommandChannel:
			switch c := cmd.(type) {
			case CommandNewSound:
				if _, ok := instance.sounds[c.ID]; !ok {
					snd := newSoundFromCommand(c)
					instance.sounds[c.ID] = snd
				}
			case CommandPlaySound:
				if snd, ok := instance.sounds[c.ID]; ok {
					snd.playAsSound(c.Volume)
				} else {
					instance.log.Errorf("[Audio] missing sound %d", c.ID)
				}
			case CommandStopSound:
			case CommandPlayMusic:
				if snd, ok := instance.sounds[c.ID]; ok {
					snd.playAsMusic(c.PlaybackID, c.Volume, 0)
					instance.playingMusic[c.PlaybackID] = c.ID
				} else {
					instance.log.Errorf("[Audio] missing sound %d", c.ID)
				}
			case CommandStopMusic:
				if sndId, ok := instance.playingMusic[c.PlaybackID]; ok {
					if snd, ok := instance.sounds[sndId]; ok {
						snd.stopMusic(c.PlaybackID)
						delete(instance.playingMusic, c.PlaybackID)
					} else {
						instance.log.Errorf("[Audio] missing sound %d", sndId)
					}
				} else {
					instance.log.Errorf("[Audio] missing music playback %d", c.PlaybackID)
				}
			case CommandStopAllMusic:
				for playbackID, sndId := range instance.playingMusic {
					if snd, ok := instance.sounds[sndId]; ok {
						snd.stopMusic(playbackID)
					}
				}
				instance.playingMusic = make(map[uint32]uint32)
			}
		}
	}
}

// Quit sends to instance's QuitChannel.
func (instance *Instance) Quit() {
	go func() {
		instance.QuitChannel <- true
	}()
}
