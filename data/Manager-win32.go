// +build windows

package data

import (
	"os"
	"path"
)

func (m *Manager) acquireConfigPath() (err error) {
	dir := path.Join(os.Getenv("APPDATA"), "chimera")

	m.ConfigPath = dir
	return
}

func (m *Manager) acquireCachePath() (err error) {
	dir := path.Join(os.Getenv("LOCALAPPDATA"), "chimera")

	m.CachePath = dir
	return
}