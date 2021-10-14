//go:build !mobile

package audio

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"unsafe"

	"github.com/jfreymuth/oggvorbis"
	"github.com/mewkiz/flac"
	"github.com/veandco/go-sdl2/mix"
	"github.com/veandco/go-sdl2/sdl"
)

type Sound struct {
	chunk *mix.Chunk
	bytes []byte
}

func (s *Sound) fromBytes(data []byte) error {
	if len(data) < 4 {
		return fmt.Errorf("not enough bytes")
	}
	if string(data[:4]) == "fLaC" {
		reader := bytes.NewReader(data)
		stream, err := flac.Parse(reader)
		if err != nil {
			return err
		}
		// Read in our full FLAC into a buffer.
		fullbuf := new(bytes.Buffer)
		for {
			f, err := stream.ParseNext()
			if err == io.EOF {
				// Done.
				break
			}
			// Interleave the samples.
			for i := 0; i < int(f.BlockSize); i++ {
				for _, sub := range f.Subframes {
					sample := int16(sub.Samples[i])
					binary.Write(fullbuf, binary.LittleEndian, sample)
				}
			}
		}
		data = fullbuf.Bytes()
		// At this point fullbuf contains our decoded FLAC stream.
		if stream.Info.SampleRate != 44100 || stream.Info.NChannels != 2 {
			var srcFormat sdl.AudioFormat
			switch stream.Info.BitsPerSample {
			case 16:
				srcFormat = sdl.AUDIO_S16SYS
			}
			var audioCvt sdl.AudioCVT
			sdl.BuildAudioCVT(&audioCvt, srcFormat, stream.Info.NChannels, int(stream.Info.SampleRate), mix.DEFAULT_FORMAT, 2, 44100)

			audioCvt.Len = int32(stream.Info.NSamples * 2 * 2)

			// Allocate our underlying C-backed buffer.
			audioCvt.AllocBuf(uintptr(audioCvt.Len * audioCvt.LenMult))

			// Copy
			s := unsafe.Slice((*byte)(audioCvt.Buf), len(data))
			copy(s, data)
			if err := sdl.ConvertAudio(&audioCvt); err != nil {
				return err
			}
			data = audioCvt.BufAsSlice()
			audioCvt.FreeBuf()
		}

		s.bytes = make([]byte, len(data))
		copy(s.bytes, data)
		chunk, err := mix.QuickLoadRAW((*uint8)(unsafe.Pointer(&s.bytes[0])), uint32(len(s.bytes)))
		if err != nil {
			return err
		}
		s.chunk = chunk
		return nil
	} else if string(data[:4]) == "OggS" { // Music
		// This is terrible, but we decode ogg into PCM and store it in RAM. This is due to the idiocy that is SDL_mixer's single music track limitation. TODO: Keep track of when music was last played back and unload it beyond a certain time.
		reader := bytes.NewReader(data)
		r, err := oggvorbis.NewReader(reader)
		if err != nil {
			return err
		}
		fmt.Println("ogg stuff", r.SampleRate(), r.Channels(), r.Length())

		return nil
	}
	return fmt.Errorf("unsupported format")
}

func (s *Sound) play(volume float32) {
	if s.chunk != nil {
		v := int(volume * 255)
		s.chunk.Volume(v)
		s.chunk.Play(-1, 0)
	}
}

func (s *Sound) playLoop(volume float32, loop int) {
}
