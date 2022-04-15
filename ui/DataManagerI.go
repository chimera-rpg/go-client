package ui

import "image"

type DataManagerI interface {
	GetDataPath(...string) string
	GetCachedImage(iID uint32) (img image.Image, err error)
}
