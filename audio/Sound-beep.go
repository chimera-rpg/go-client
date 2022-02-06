package audio

import (
	"errors"
	"log"
	"os"

	"github.com/chimera-rpg/go-common/network"
	"github.com/faiface/beep"
	"github.com/faiface/beep/effects"
	"github.com/faiface/beep/flac"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/vorbis"
)

type Sound struct {
	filepath    string
	soundType   uint8
	format      beep.Format
	streamer    beep.StreamSeekCloser
	soundBuffer *beep.Buffer
	instances   map[uint32]*SoundInstance
}

type SoundInstance struct {
	control   *beep.Ctrl
	volume    *effects.Volume
	resampler *beep.Resampler
}

func newSoundFromCommand(c CommandNewSound) *Sound {
	s := Sound{
		instances: make(map[uint32]*SoundInstance),
		filepath:  c.Filepath,
		soundType: c.Type,
	}

	streamer, format, err := s.decode()
	if err != nil {
		return &s
	}
	s.streamer = streamer
	s.format = format

	return &s
}

func (s *Sound) decode() (beep.StreamSeekCloser, beep.Format, error) {
	f, err := os.Open(s.filepath)
	if err != nil {
		return nil, beep.Format{}, err
	}
	if s.soundType == network.SoundFlac {
		streamer, format, err := flac.Decode(f)
		if err != nil {
			log.Fatal(err)
		}
		return streamer, format, nil
	} else if s.soundType == network.SoundOgg {
		streamer, format, err := vorbis.Decode(f)
		if err != nil {
			log.Fatal(err)
		}
		return streamer, format, nil
	}
	return nil, beep.Format{}, errors.New("unsupported file")
}

func (s *Sound) playAsSound(volume float32) {
	if s.soundBuffer == nil {
		resampler := beep.Resample(4, s.format.SampleRate, beep.SampleRate(44100), s.streamer)
		s.soundBuffer = beep.NewBuffer(s.format)
		s.soundBuffer.Append(resampler)
	}
	sound := s.soundBuffer.Streamer(0, s.soundBuffer.Len())
	vol := &effects.Volume{Streamer: sound, Base: float64(volume)}
	speaker.Play(vol)
}

func (s *Sound) playAsMusic(id uint32, volume float32, loop int) *SoundInstance {
	if si, ok := s.instances[id]; ok {
		return si
	}

	streamer, format, err := s.decode()
	if err != nil {
		return nil
	}

	if loop == 0 {
		loop = -1
	}
	si := &SoundInstance{}
	si.control = &beep.Ctrl{
		Streamer: beep.Loop(loop, streamer),
	}
	si.resampler = beep.Resample(4, format.SampleRate, beep.SampleRate(44100), si.control)
	si.volume = &effects.Volume{Streamer: si.resampler, Base: float64(volume)}

	s.instances[id] = si

	speaker.Play(si.volume)

	return si
}

func (s *Sound) stopMusic(id uint32) {
	si, ok := s.instances[id]
	if !ok {
		return
	}
	si.control.Paused = true
	delete(s.instances, id)
}
