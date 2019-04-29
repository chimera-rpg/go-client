// +build darwin

package data

import (
	"os"
	"path"
)

func (m *Manager) acquireConfigPath() (err error) {
	dir := path.Join(os.Getenv("HOME"), "Library/Application Support")

	m.ConfigPath = path.Join(dir, "chimera")
	return
}

func (m *Manager) acquireCachePath() (err error) {
	dir := path.Join(os.Getenv("HOME"), "Library/Caches")

	m.CachePath = path.Join(dir, "chimera")
	return
}
