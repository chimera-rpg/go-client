package data

import (
	"os"
	"path"
	"path/filepath"
)

// Manager handles access to files on the system.
type Manager struct {
    DataPath string     // Path for client data (fonts, etc.)
    ConfigPath string   // Path for user configuration (style overrides, bindings, etc.)
    CachePath string    // Path for local cache (downloaded PNGs, etc.)
}

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

func (m *Manager) GetDataPath(parts ...string) string {
	return path.Join(m.DataPath, path.Join(parts...))
}