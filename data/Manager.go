package data

import (
	"image"
	// Package image/png is not used explicitly in the code below,
	// but is imported for its initialization side-effect, which allows
	// image.Decode to understand PNG formatted images.
	_ "image/png"
	"log"
	"os"
	"path"
	"path/filepath"

	"github.com/chimera-rpg/go-common/network"
	"github.com/kettek/apng"
)

// Manager handles access to files on the system.
type Manager struct {
	Conn       *network.Connection
	Log        *log.Logger
	DataPath   string // Path for client data (fonts, etc.)
	ConfigPath string // Path for user configuration (style overrides, bindings, etc.)
	CachePath  string // Path for local cache (downloaded PNGs, etc.)
	animations map[uint32]Animation
}

// Setup gets the required data/config/cache paths and creates them if needed.
func (m *Manager) Setup() (err error) {
	m.Log = log.New(os.Stdout, "Manager: ", log.Ltime)
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
			err = os.MkdirAll(m.ConfigPath, os.ModeDir)
		}
		if err != nil {
			return
		}
	}
	if _, err = os.Stat(m.CachePath); err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(m.CachePath, os.ModeDir)
		}
		if err != nil {
			return
		}
	}
	m.animations = make(map[uint32]Animation)
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

// EnsureAnimation checks if an animation associated with a given ID exists, and if not, sends a network request for the animation.
func (m *Manager) EnsureAnimation(aID uint32) {
	// If animation id is not known, add the animation, then send an animation request.
	if _, animExists := m.animations[aID]; !animExists {
		m.animations[aID] = Animation{
			Faces: make(map[uint32][]AnimationFrame),
		}
		m.Log.Printf("Sending animrequest for %d\n", aID)
		m.Conn.Send(network.CommandAnimation{
			Type:        network.Get,
			AnimationID: aID,
		})
	}

}

// HandleAnimationCommand handles received animation commands and incorporates them into the animations map.
func (m *Manager) HandleAnimationCommand(cmd network.CommandAnimation) error {
	if _, exists := m.animations[cmd.AnimationID]; !exists {
		m.animations[cmd.AnimationID] = Animation{
			AnimationID: cmd.AnimationID,
			Faces:       make(map[uint32][]AnimationFrame),
		}
	}
	m.Log.Printf("Received animrequest %d: %+v\n", cmd.AnimationID, cmd)
	for faceID, frames := range cmd.Faces {
		m.animations[cmd.AnimationID].Faces[faceID] = make([]AnimationFrame, len(frames))
		for frameIndex, frame := range frames {
			m.animations[cmd.AnimationID].Faces[faceID][frameIndex] = AnimationFrame{
				ImageID: frame.ImageID,
				Time:    frame.Time,
			}
		}
	}
	return nil
}
