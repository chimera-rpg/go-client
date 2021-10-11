//go:build !mobile

package audio

import (
	"github.com/sirupsen/logrus"
	"github.com/veandco/go-sdl2/mix"
)

// Setup initializes SDL2_mixer and opens the default audio device at 44100 KHz, signed 16-bit samples in system order, 2 channels, and 1024 chunk size. TODO: Make configurable via Config because why not.
func (instance *Instance) Setup(l *logrus.Logger) (err error) {
	instance.CommandChannel = make(chan CommandI)
	instance.QuitChannel = make(chan bool)
	instance.sounds = make(map[uint32]Sound)
	instance.log = l
	err = mix.Init(mix.INIT_FLAC)
	if err != nil {
		return err
	}
	// TODO: Use SDL_GetAudioDeviceName and SDL_OpenAudioDevice to allow for specific audio card selection.
	err = mix.OpenAudio(44100, mix.DEFAULT_FORMAT, 2, 1024)
	if err != nil {
		return err
	}

	mix.AllocateChannels(128) // For now just use 128 channels.
	instance.log.Infoln("[Audio] Initialized")

	return nil
}
