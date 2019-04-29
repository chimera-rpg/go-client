// +build !windows,!darwin

package data

import (
	"os"
	"path"
)

func (m *Manager) acquireConfigPath() (err error) {
	dir := os.Getenv("XDG_CONFIG_HOME")
	if dir == "" {
		dir = path.Join(os.Getenv("HOME"), ".config")
	}

	m.ConfigPath = path.Join(dir, "chimera")
	return
}

func (m *Manager) acquireCachePath() (err error) {
	dir := os.Getenv("XDG_CACHE_HOME")
	if dir == "" {
		dir = path.Join(os.Getenv("HOME"), ".cache")
	}

	m.CachePath = path.Join(dir, "chimera")
	return
}
