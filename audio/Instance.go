package audio

import (
	"github.com/sirupsen/logrus"
)

// Instance is the managing instance of the Audio system.
type Instance struct {
	log            *logrus.Logger
	sounds         map[uint32]Sound
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
					snd := Sound{}
					// Decode bytes into a full PCM buffer.
					/*bytes, err := instance.decodeBytes(c.Bytes)
					if err != nil {
						// return
					}*/
					// Resample to our audio device's format...?
					if err := snd.fromBytes(c.Bytes); err != nil {
						instance.log.Errorln(err)
					}
					instance.sounds[c.ID] = snd
				}
			case CommandPlaySound:
				if snd, ok := instance.sounds[c.ID]; ok {
					snd.play(c.Volume)
				} else {
					instance.log.Errorf("[Audio] missing sound %d", c.ID)
				}
			case CommandStopSound:
			case CommandPlayMusic:
			case CommandStopMusic:
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
