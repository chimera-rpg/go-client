package ui

import (
	"fmt"
	"image"

	"github.com/veandco/go-sdl2/sdl"
)

type ImageTextures struct {
	grayscale *sdl.Texture
	regular   *sdl.Texture
}

// DataManager is a ui-contextualized data manager that is used to indirectly call the main client's data manager and cache as needed.
type DataManager struct {
	imageCache    map[uint32]image.Image
	imageTextures map[uint32]*ImageTextures
	manager       DataManagerI
}

// GetDataPath just gets the normal DataManager's path.
func (m *DataManager) GetDataPath(s ...string) string {
	return m.manager.GetDataPath(s...)
}

// GetCachedImage returns a ui-stored version of the image if it exists, otherwise it calls the main client's method.
func (m *DataManager) GetCachedImage(iID uint32) (img image.Image, err error) {
	if img, ok := m.imageCache[iID]; ok {
		return img, nil
	}
	img, err = m.manager.GetCachedImage(iID)
	if err == nil {
		fmt.Println("cached local")
		m.imageCache[iID] = img
	}
	return
}
