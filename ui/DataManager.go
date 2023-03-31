package ui

import (
	"errors"
	"image"

	"github.com/veandco/go-sdl2/sdl"
)

// DataManager is a ui-contextualized data manager that is used to indirectly call the main client's data manager and cache as needed.
type DataManager struct {
	imageCache    map[uint32]image.Image
	imageTextures map[uint32]*Image
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
		m.imageCache[iID] = img
	}
	return
}

func (m *DataManager) ClearCachedImage(iID uint32) {
	delete(m.imageCache, iID)
	if tex, ok := m.imageTextures[iID]; ok {
		if tex.grayscaleTexture != nil {
			tex.grayscaleTexture.Destroy()
		}
		if tex.outlineTexture != nil {
			tex.outlineTexture.Destroy()
		}
		if tex.regularTexture != nil {
			tex.regularTexture.Destroy()
		}
		delete(m.imageTextures, iID)
	}
}

func (m *DataManager) GetImage(iID uint32) *Image {
	return m.imageTextures[iID]
}

func (m *DataManager) SetRegularTexture(iID uint32, tex *sdl.Texture) {
	if _, ok := m.imageTextures[iID]; !ok {
		m.imageTextures[iID] = &Image{}
	}
	m.imageTextures[iID].regularTexture = tex
}

func (m *DataManager) GetRegularTexture(iID uint32) (text *sdl.Texture, err error) {
	if _, ok := m.imageTextures[iID]; ok {
		return m.imageTextures[iID].regularTexture, nil
	}
	return nil, errors.New("missing")
}

func (m *DataManager) SetGrayscaleTexture(iID uint32, tex *sdl.Texture) {
	if _, ok := m.imageTextures[iID]; !ok {
		m.imageTextures[iID] = &Image{}
	}
	m.imageTextures[iID].grayscaleTexture = tex
}

func (m *DataManager) GetGrayscaleTexture(iID uint32) (text *sdl.Texture, err error) {
	if _, ok := m.imageTextures[iID]; ok {
		return m.imageTextures[iID].grayscaleTexture, nil
	}
	return nil, errors.New("missing")
}

func (m *DataManager) SetOutlineTexture(iID uint32, tex *sdl.Texture) {
	if _, ok := m.imageTextures[iID]; !ok {
		m.imageTextures[iID] = &Image{}
	}
	m.imageTextures[iID].outlineTexture = tex
}

func (m *DataManager) GetOutlineTexture(iID uint32) (text *sdl.Texture, err error) {
	if _, ok := m.imageTextures[iID]; ok {
		return m.imageTextures[iID].outlineTexture, nil
	}
	return nil, errors.New("missing")
}
