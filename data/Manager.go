package data

import (
	"bytes"
	"errors"
	"image"
	"strconv"
	"strings"
	"sync"

	// Package image/png is not used explicitly in the code below,
	// but is imported for its initialization side-effect, which allows
	// image.Decode to understand PNG formatted images.
	_ "image/png"
	"os"
	"path"
	"path/filepath"

	"github.com/sirupsen/logrus"

	"github.com/chimera-rpg/go-client/config"
	"github.com/chimera-rpg/go-common/network"
	"github.com/kettek/apng"
)

type ImageRef struct {
	ID  uint32
	img image.Image
}

// Manager handles access to files on the system.
type Manager struct {
	Conn       *network.Connection
	Log        *logrus.Logger
	DataPath   string // Path for client data (fonts, etc.)
	ConfigPath string // Path for user configuration (style overrides, bindings, etc.)
	Config     config.Config
	CachePath  string // Path for local cache (downloaded PNGs, etc.)
	animations map[uint32]Animation
	audio      map[uint32]Audio
	//images         map[uint32]image.Image
	images         []ImageRef
	imageLock      sync.Mutex
	sounds         map[uint32]SoundEntry
	handleCallback func(netID int, cmd network.Command)
}

// Setup gets the required data/config/cache paths and creates them if needed.
func (m *Manager) Setup(l *logrus.Logger) (err error) {
	m.Log = l
	// Acquire our various paths.
	if err = m.acquireDataPath(); err != nil {
		return
	}
	if err = m.acquireConfigPath(); err != nil {
		return
	}
	if err = m.acquireCachePath(); err != nil {
		return
	}
	// Ensure each exists.
	if _, err = os.Stat(m.DataPath); err != nil {
		// DataPath does not exist!
		return
	}
	if _, err = os.Stat(m.ConfigPath); err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(m.ConfigPath, 0750)
		}
		if err != nil {
			return
		}
	}
	if _, err = os.Stat(m.CachePath); err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(m.CachePath, 0755)
		}
		if err != nil {
			return
		}
	}
	// Also ensure images directory exists.
	imagesPath := path.Join(m.CachePath, "images")
	if _, err = os.Stat(imagesPath); err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(imagesPath, 0755)
		}
		if err != nil {
			return
		}
	}
	// Also ensure sounds directory exists.
	soundsPath := path.Join(m.CachePath, "sounds")
	if _, err = os.Stat(soundsPath); err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(soundsPath, 0755)
		}
		if err != nil {
			return
		}
	}

	// Read in our config.
	if err := m.Config.Read(path.Join(m.ConfigPath, "client.yaml")); err != nil {
		m.Log.Info(err)
	}

	m.animations = make(map[uint32]Animation)
	m.audio = make(map[uint32]Audio)
	//m.images = make(map[uint32]image.Image)
	m.sounds = make(map[uint32]SoundEntry)

	// Collect cached images.
	if err = m.collectCachedImages(); err != nil {
		m.Log.Error("[Manager] ", err)
	}
	m.Log.WithFields(logrus.Fields{
		"Count": len(m.images),
	}).Print("Loaded cached images")

	// Collect cached sounds.
	if err = m.collectCachedSounds(); err != nil {
		m.Log.Error("[Manager] ", err)
	}
	m.Log.WithFields(logrus.Fields{
		"Count": len(m.sounds),
	}).Print("Loaded cached sounds")
	return
}

// GetDataPath gets a path relative to the data path directory.
func (m *Manager) GetDataPath(parts ...string) string {
	return path.Join(m.DataPath, path.Clean("/"+path.Join(parts...)))
}

// GetCachePath gets a path relative to the cache path directory.
func (m *Manager) GetCachePath(parts ...string) string {
	return path.Join(m.CachePath, path.Clean("/"+path.Join(parts...)))
}

// GetConfigPath gets a path relative to the config path directory.
func (m *Manager) GetConfigPath(parts ...string) string {
	return path.Join(m.ConfigPath, path.Clean("/"+path.Join(parts...)))
}

func (m *Manager) acquireDataPath() (err error) {
	var dir string
	// Set our path which should be <parent of cmd>/share/chimera/client.
	if dir, err = filepath.Abs(os.Args[0]); err != nil {
		return
	}
	dir = path.Join(filepath.Dir(filepath.Dir(dir)), "share", "chimera", "client")

	m.DataPath = dir
	return
}

// Sounds returns the manager's sounds
func (m *Manager) Sounds() map[uint32]SoundEntry {
	return m.sounds
}

// collectCachedImages reads the cache directory for images to load.
func (m *Manager) collectCachedImages() (err error) {
	imagesPath := path.Join(m.CachePath, "images")
	err = filepath.Walk(imagesPath, func(filepath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			if strings.HasSuffix(filepath, ".png") {
				shortpath := filepath[len(imagesPath)+1:]
				shortpath = shortpath[:len(shortpath)-len(".png")]
				ui64, err := strconv.ParseUint(shortpath, 10, 32)
				if err != nil {
					m.Log.Warn("[Manager] ", err)
					return nil
				}
				i := uint32(ui64)
				img, err := m.GetImage(filepath)
				if err != nil {
					m.Log.Warn("[Manager] ", err)
					return nil
				}
				m.SetCachedImage(i, img, true)
			}
		}
		return nil
	})
	return
}

// collectCachedSounds reads the cache directory for sounds to load.
func (m *Manager) collectCachedSounds() (err error) {
	soundsPath := path.Join(m.CachePath, "sounds")
	err = filepath.Walk(soundsPath, func(filepath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			var dataType uint8
			knownType := false
			shortpath := filepath[len(soundsPath)+1:]
			if strings.HasSuffix(filepath, ".flac") {
				shortpath = shortpath[:len(shortpath)-len(".flac")]
				dataType = network.SoundFlac
				knownType = true
			} else if strings.HasSuffix(filepath, ".ogg") {
				shortpath = shortpath[:len(shortpath)-len(".ogg")]
				dataType = network.SoundOgg
				knownType = true
			}
			if knownType {
				ui64, err := strconv.ParseUint(shortpath, 10, 32)
				if err != nil {
					m.Log.Warn("[Manager] ", err)
					return nil
				}
				i := uint32(ui64)
				m.sounds[i] = SoundEntry{
					Filepath: filepath,
					Type:     dataType,
				}
			}
		}
		return nil
	})
	return
}

// WriteImage writes image data to the images subdirectory in the cachePath.
func (m *Manager) WriteImage(imageID uint32, imageType uint8, data []byte) error {
	targetPath := path.Join(m.CachePath, "images", strconv.FormatUint(uint64(imageID), 10))
	if imageType == network.GraphicsPng {
		targetPath = targetPath + ".png"
	}
	return m.WriteBytes(targetPath, data)
}

// WriteSound writes sound data to the sounds subdirectory in the cachePath.
func (m *Manager) WriteSound(soundID uint32, soundType uint8, data []byte) error {
	targetPath := path.Join(m.CachePath, "sounds", strconv.FormatUint(uint64(soundID), 10))
	if soundType == network.SoundFlac {
		targetPath = targetPath + ".flac"
	} else if soundType == network.SoundOgg {
		targetPath = targetPath + ".ogg"
	}
	return m.WriteBytes(targetPath, data)
}

// WriteBytes writes bytes to a file path.
func (m *Manager) WriteBytes(file string, data []byte) (err error) {
	var writer *os.File

	writer, err = os.OpenFile(file, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return
	}
	defer writer.Close()
	writer.Write(data)
	return
}

// GetBytes returns the given file as a slice of bytes.
func (m *Manager) GetBytes(file string) (data []byte, err error) {
	var reader *os.File
	var info os.FileInfo
	reader, err = os.Open(file)
	if err != nil {
		return
	}
	info, err = reader.Stat()
	if err != nil {
		return
	}
	data = make([]byte, info.Size())
	_, err = reader.Read(data)
	return
}

// GetAPNG returns the given file as an APNG type.
func (m *Manager) GetAPNG(file string) (img apng.APNG, err error) {
	var reader *os.File
	reader, err = os.Open(file)
	if err != nil {
		return
	}
	img, err = apng.DecodeAll(reader)
	return
}

// GetImage returns the given file as an Image (if it is supported).
func (m *Manager) GetImage(file string) (img image.Image, err error) {
	var reader *os.File
	reader, err = os.Open(file)
	if err != nil {
		return
	}
	img, _, err = image.Decode(reader)
	return

}

// GetFace returns the frames for a given animation and face.
func (m *Manager) GetFace(aID uint32, fID uint32) (f []AnimationFrame) {
	anim, animExists := m.animations[aID]
	if !animExists {
		return
	}
	face, faceExists := anim.Faces[fID]
	if !faceExists {
		return
	}
	return face
}

// GetAnimation returns the underlying animation.
func (m *Manager) GetAnimation(aID uint32) Animation {
	anim, animExists := m.animations[aID]
	if !animExists {
		return Animation{}
	}
	return anim
}

// GetCachedImage returns the cached image associated with the given ID.
func (m *Manager) GetCachedImage(iID uint32) (img image.Image, err error) {
	m.imageLock.Lock()
	defer m.imageLock.Unlock()
	for _, ref := range m.images {
		if ref.ID == iID {
			return ref.img, nil
		}
	}
	imageData, err := m.GetImage(m.GetDataPath("ui/loading.png"))
	if err != nil {
		return imageData, errors.New("missing")
	}
	return nil, errors.New("loading")
}

func (m *Manager) SetCachedImage(iID uint32, img image.Image, override bool) {
	m.imageLock.Lock()
	defer m.imageLock.Unlock()
	for _, ref := range m.images {
		if ref.ID == iID {
			if override {
				ref.img = img
			}
			return
		}
	}
	m.images = append(m.images, ImageRef{
		ID:  iID,
		img: img,
	})
}

// GetAudioSound returns the associated Sound for an audioID, soundID, and index.
func (m *Manager) GetAudioSound(audioID, soundID uint32, index int) (Sound, bool) {
	audio, ok := m.audio[audioID]
	if !ok {
		return Sound{}, false
	}
	sound, ok := audio.SoundSets[soundID]
	if !ok {
		return Sound{}, false
	}
	if len(sound) == 0 {
		return Sound{}, false
	}
	return sound[0], true
}

// GetCachedSound returns the cached sound associated with the given ID.
func (m *Manager) GetCachedSound(soundID uint32) (snd SoundEntry) {
	if s, ok := m.sounds[soundID]; ok {
		return s
	}
	return SoundEntry{}
}

// EnsureAnimation checks if an animation associated with a given ID exists, and if not, sends a network request for the animation.
func (m *Manager) EnsureAnimation(aID uint32) {
	// If animation id is not known, add the animation, then send an animation request.
	if _, animExists := m.animations[aID]; !animExists {
		m.animations[aID] = Animation{
			Faces:       make(map[uint32][]AnimationFrame),
			RandomFrame: false,
			Ready:       false,
		}
		m.Log.WithFields(logrus.Fields{
			"ID": aID,
		}).Info("[Manager] Sending Animation Request")
		m.Conn.Send(network.CommandAnimation{
			Type:        network.Get,
			AnimationID: aID,
		})
	}

}

// HandleAnimationCommand handles received animation commands and incorporates them into the animations map.
func (m *Manager) HandleAnimationCommand(cmd network.CommandAnimation) error {
	if anim, exists := m.animations[cmd.AnimationID]; !exists {
		m.animations[cmd.AnimationID] = Animation{
			AnimationID: cmd.AnimationID,
			RandomFrame: cmd.RandomFrame,
			Ready:       true,
			Faces:       make(map[uint32][]AnimationFrame),
		}
	} else {
		anim.Ready = true
		anim.RandomFrame = cmd.RandomFrame
		m.animations[cmd.AnimationID] = anim
	}
	m.Log.WithFields(logrus.Fields{
		"ID":        cmd.AnimationID,
		"FaceCount": len(cmd.Faces),
	}).Info("[Manager] Received Animation")
	for faceID, frames := range cmd.Faces {
		m.animations[cmd.AnimationID].Faces[faceID] = make([]AnimationFrame, len(frames))
		for frameIndex, frame := range frames {
			m.animations[cmd.AnimationID].Faces[faceID][frameIndex] = AnimationFrame{
				ImageID: frame.ImageID,
				Time:    frame.Time,
				X:       frame.X,
				Y:       frame.Y,
			}
			// Request any unknown graphics.
			m.EnsureImage(frame.ImageID)
		}
	}
	if m.handleCallback != nil {
		m.handleCallback(network.TypeAnimation, cmd)
	}
	return nil
}

// EnsureImage ensures that the given image is available. If it is not, then send a graphics request.
func (m *Manager) EnsureImage(iID uint32) {
	exists := false
	for _, ref := range m.images {
		if iID == ref.ID {
			exists = true
			break
		}
	}
	if !exists {
		imageData, err := m.GetImage(m.GetDataPath("ui/loading.png"))
		if err != nil {
			m.SetCachedImage(iID, image.NewNRGBA(image.Rectangle{image.Point{0, 0}, image.Point{8, 8}}), false)
		} else {
			m.SetCachedImage(iID, imageData, false)
		}

		// Send request.
		m.Log.WithFields(logrus.Fields{
			"ID":       iID,
			"DataType": network.GraphicsPng,
		}).Info("[Manager] Sending Graphics Request")
		m.Conn.Send(network.CommandGraphics{
			Type:       network.Get,
			GraphicsID: iID,
			DataType:   network.GraphicsPng,
		})
	}
}

// HandleGraphicsCommand handles CommandGraphics.
func (m *Manager) HandleGraphicsCommand(cmd network.CommandGraphics) error {
	m.Log.WithFields(logrus.Fields{
		"ID":       cmd.GraphicsID,
		"Type":     cmd.Type,
		"DataType": cmd.DataType,
		"Length":   len(cmd.Data),
	}).Info("[Manager] Received Graphics")
	if cmd.Type == network.Nokay {
		m.Log.Warn("[Manager] Server sent missing image")
		// FIXME: We should have some sort of "missing image" reference here.
		imageData, err := m.GetImage(m.GetDataPath("ui/missing.png"))
		if err != nil {
			m.SetCachedImage(cmd.GraphicsID, image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{8, 8}}), true)
		} else {
			m.SetCachedImage(cmd.GraphicsID, imageData, true)
		}
	} else if cmd.Type == network.Set {
		if cmd.DataType == network.GraphicsPng {
			// Decode PNG
			if img, _, err := image.Decode(bytes.NewReader(cmd.Data)); err != nil {
				m.Log.Warn("[Manager] Could not Decode Image")
			} else {
				m.SetCachedImage(cmd.GraphicsID, img, true)
			}
			// Also write the image to disk for future use.
			if err := m.WriteImage(cmd.GraphicsID, cmd.DataType, cmd.Data); err != nil {
				m.Log.Warn("[Manager] ", err)
			}
		} else {
			m.Log.Warn("[Manager] Unhandled Graphics Type")
		}
	} else {
		m.Log.Warn("[Manager] Bogus Graphics Message")
	}
	if m.handleCallback != nil {
		m.handleCallback(network.TypeGraphics, cmd)
	}
	return nil
}

// EnsureAudio checks if an audio associated with a given ID exists, and if not, sends a network request for the audio.
func (m *Manager) EnsureAudio(aID uint32) {
	if _, audioExists := m.audio[aID]; !audioExists {
		m.audio[aID] = Audio{
			SoundSets: make(map[uint32][]Sound),
			Pending:   true,
		}
		m.Log.WithFields(logrus.Fields{
			"ID": aID,
		}).Info("[Manager] Sending Audio Request")
		m.Conn.Send(network.CommandAudio{
			Type:    network.Get,
			AudioID: aID,
		})
	}
}

// HandleAudioCommand handles receiving audio commands and incorporates them into the audio map.
func (m *Manager) HandleAudioCommand(cmd network.CommandAudio) error {
	if audio, exists := m.audio[cmd.AudioID]; !exists {
		m.audio[cmd.AudioID] = Audio{
			AudioID:   cmd.AudioID,
			SoundSets: make(map[uint32][]Sound),
			Pending:   len(cmd.Sounds) != 0,
		}
	} else {
		if len(cmd.Sounds) == 0 {
			audio.Pending = false
			m.audio[cmd.AudioID] = audio
		}
	}
	m.Log.WithFields(logrus.Fields{
		"ID":     cmd.AudioID,
		"Sounds": len(cmd.Sounds),
	}).Info("[Manager] Received Audio")
	for soundSetID, soundSets := range cmd.Sounds {
		m.audio[cmd.AudioID].SoundSets[soundSetID] = make([]Sound, len(soundSets))
		for soundIndex, sound := range soundSets {
			s := Sound{
				SoundID: sound.SoundID,
				Text:    sound.Text,
			}
			m.audio[cmd.AudioID].SoundSets[soundSetID][soundIndex] = s
			// Request any unknown sounds.
			m.EnsureSound(sound.SoundID)
		}
	}
	if m.handleCallback != nil {
		m.handleCallback(network.TypeAudio, cmd)
	}
	return nil
}

// EnsureSound ensures that the given sound is available. If it is not, then send a sound request.
func (m *Manager) EnsureSound(iID uint32) bool {
	if _, soundExists := m.sounds[iID]; !soundExists {
		// Add an empty bytes list as placeholder for now.
		m.sounds[iID] = SoundEntry{
			Pending: true,
		}

		// Send request.
		m.Log.WithFields(logrus.Fields{
			"ID": iID,
		}).Info("[Manager] Sending Sound Request")
		m.Conn.Send(network.CommandSound{
			Type:    network.Get,
			SoundID: iID,
		})
		return false
	}
	return true
}

// HandleSoundCommand handles CommandGraphics.
func (m *Manager) HandleSoundCommand(cmd network.CommandSound) error {
	m.Log.WithFields(logrus.Fields{
		"ID":       cmd.SoundID,
		"Type":     cmd.Type,
		"DataType": cmd.DataType,
		"Length":   len(cmd.Data),
	}).Info("[Manager] Received Sound")
	if cmd.Type == network.Nokay {
		m.Log.Warn("[Manager] Server sent missing sound")
		m.sounds[cmd.SoundID] = SoundEntry{
			Pending: false,
		}
	} else if cmd.Type == network.Set {
		if cmd.DataType == network.SoundFlac || cmd.DataType == network.SoundOgg {
			// Write the sound to disk for future use.
			if err := m.WriteSound(cmd.SoundID, cmd.DataType, cmd.Data); err != nil {
				m.Log.Warn("[Manager] ", err)
			}

			// Acquire file path and get our reader. This is a rough duplicate of our initial caching code.
			targetPath := path.Join(m.CachePath, "sounds", strconv.FormatUint(uint64(cmd.SoundID), 10))
			if cmd.DataType == network.SoundFlac {
				targetPath = targetPath + ".flac"
			} else if cmd.DataType == network.SoundOgg {
				targetPath = targetPath + ".ogg"
			}
			m.sounds[cmd.SoundID] = SoundEntry{
				Filepath: targetPath,
				Type:     cmd.DataType,
			}
		} else {
			m.Log.Warn("[Manager] Unhandled Sound Type")
		}
	} else {
		m.Log.Warn("[Manager] Bogus Sound Message")
	}
	if m.handleCallback != nil {
		m.handleCallback(network.TypeSound, cmd)
	}
	return nil
}

// SetHandleCallback sets the handleCallback to the passed func.
func (m *Manager) SetHandleCallback(handleCallback func(netID int, cmd network.Command)) {
	m.handleCallback = handleCallback
}
