package audio

import (
	"github.com/faiface/beep/speaker"
	"github.com/sirupsen/logrus"
)

func (instance *Instance) Setup(l *logrus.Logger) (err error) {
	instance.CommandChannel = make(chan CommandI)
	instance.QuitChannel = make(chan bool)
	instance.sounds = make(map[uint32]*Sound)
	instance.playingMusic = make(map[uint32]uint32)
	instance.log = l

	err = speaker.Init(44100, 2048)
	if err != nil {
		return err
	}

	instance.log.Infoln("[Audio] beep Initialized")

	return nil
}
