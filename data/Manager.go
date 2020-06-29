package data

import (
	"bytes"
	"image"
	"strconv"
	"strings"

	// Package image/png is not used explicitly in the code below,
	// but is imported for its initialization side-effect, which allows
	// image.Decode to understand PNG formatted images.
	_ "image/png"
	"os"
	"path"
	"path/filepath"

	"github.com/sirupsen/logrus"

	"github.com/chimera-rpg/go-common/network"
	"github.com/kettek/apng"
)

// Manager handles access to files on the system.
type Manager struct {
	Conn       *network.Connection
	Log        *logrus.Logger
	DataPath   string // Path for client data (fonts, etc.)
	ConfigPath string // Path for user configuration (style overrides, bindings, etc.)
	CachePath  string // Path for local cache (downloaded PNGs, etc.)
	animations map[uint32]Animation
	images     map[uint32]image.Image
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
			err = os.MkdirAll(m.ConfigPath, 0640)
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

	m.animations = make(map[uint32]Animation)
	m.images = make(map[uint32]image.Image)

	// Collect cached images.
	if err = m.collectCachedImages(); err != nil {
		m.Log.Error("[Manager] ", err)
	}
	m.Log.WithFields(logrus.Fields{
		"Count": len(m.images),
	}).Print("Loaded cached images")

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
				m.images[i] = img
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

// GetCachedImage returns the cached image associated with the given ID.
func (m *Manager) GetCachedImage(iID uint32) (img image.Image) {
	if img, ok := m.images[iID]; ok {
		return img
	}
	imageData, err := m.GetImage(m.GetDataPath("ui/loading.png"))
	if err != nil {
		return imageData
	}
	return
}

// EnsureAnimation checks if an animation associated with a given ID exists, and if not, sends a network request for the animation.
func (m *Manager) EnsureAnimation(aID uint32) {
	// If animation id is not known, add the animation, then send an animation request.
	if _, animExists := m.animations[aID]; !animExists {
		m.animations[aID] = Animation{
			Faces: make(map[uint32][]AnimationFrame),
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
	if _, exists := m.animations[cmd.AnimationID]; !exists {
		m.animations[cmd.AnimationID] = Animation{
			AnimationID: cmd.AnimationID,
			Faces:       make(map[uint32][]AnimationFrame),
		}
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
			}
			// Request any unknown graphics.
			m.EnsureImage(frame.ImageID)
		}
	}
	return nil
}

// EnsureImage ensures that the given image is available. If it is not, then send a graphics request.
func (m *Manager) EnsureImage(iID uint32) {
	if _, imageExists := m.images[iID]; !imageExists {
		m.images[iID] = image.NewNRGBA(image.Rectangle{image.Point{0, 0}, image.Point{8, 8}})
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

// HandleGraphicsCommand
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
		m.images[cmd.GraphicsID] = image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{8, 8}})
	} else if cmd.Type == network.Set {
		if cmd.DataType == network.GraphicsPng {
			// Decode PNG
			if img, _, err := image.Decode(bytes.NewReader(cmd.Data)); err != nil {
				m.Log.Warn("[Manager] Could not Decode Image")
			} else {
				m.images[cmd.GraphicsID] = img
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
	return nil
}
