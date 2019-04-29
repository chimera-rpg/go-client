package data

import (
	"os"
	"path"
	"path/filepath"
)

// Manager handles access to files on the system.
type Manager struct {
	DataPath   string // Path for client data (fonts, etc.)
	ConfigPath string // Path for user configuration (style overrides, bindings, etc.)
	CachePath  string // Path for local cache (downloaded PNGs, etc.)
}

// Setup gets the required data/config/cache paths and creates them if needed.
func (m *Manager) Setup() (err error) {
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
